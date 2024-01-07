package discord

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"gemini-discord.mashu.idv.tw/repository/chat_session"
	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
)

type commandHandler struct {
	interactionCreateHandlerMap map[string]func(
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
		optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	) error

	messageCreateHandlerMap map[string]func(
		s *discordgo.Session,
		m *discordgo.MessageCreate,
	) error

	textModel  *genai.GenerativeModel
	imageModel *genai.GenerativeModel

	chatSessRepository chat_session.ChatSessionRepository
}

func (h *commandHandler) iGeminiChat(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "New chat session started",
		},
	})

	return nil
}

func (h *commandHandler) mGeminiChat(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
) error {
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return err
	}
	if ch.IsThread() {
		s.ChannelMessageSend(ch.ID, "cannot run this command in a thread")
		return nil
	}

	// "thread" is seen as a Discord channel
	thread, err := s.MessageThreadStart(
		m.ChannelID,
		m.ID,
		"Chat Session",
		60,
	)
	if err != nil {
		log.Printf("failed to create new thread: %v\n", err)
	}
	log.Printf("thread created with ID: %s\n", thread.ID)

  sess := h.textModel.StartChat()

  initMessage :=
    "你是一位生活小助理, 可以協助user回答各式各樣的問題.\n" +
    "你現在位於一個聊天室中, 將會有多個user跟你聊天.\n" +
    "你必須成功的識別每個跟你說話的人, 傳訊息給你的人, 其聊天格式會以以下方式呈現:\n" +
    "[{USER_NAME}]: {MESSAGE}.\n" +
    "但是你在回答的時候並不要使用這個格式, 直接回答你要講的就好.\n\n" +
    "請注意, USER_NAME的格式會像是這樣: Johnny1234\n" +
    "後面的數字是discord的ID, 但請在回答時忽這個ID\n"
  sess.SendMessage(context.Background(), genai.Text(initMessage))
	h.chatSessRepository.Save(thread.ID, sess)

	return nil
}

func (h *commandHandler) geminiPhoto(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
) error {
	if err := h.sendGeneratingMessage(s, i); err != nil {
		return fmt.Errorf("failed to send discord response: %v", err)
	}

	inputText := optionMap["text"].StringValue()

	attachmentId := optionMap["image"].Value.(string)
	attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentId]

	imageType, err := h.getImageType(attachment)
	if err != nil {
		return fmt.Errorf("failed to get image type: %v", err)
	}

	imagePath, err := h.downloadImage(attachment.URL, attachment.ID, imageType)
	if err != nil {
		return fmt.Errorf("failed to download image from command: %v", err)
	}

	generatedRes, err := h.generatePhotoResponse(imageType, imagePath, inputText)
	if err != nil {
		return fmt.Errorf("failed to generate response: %v", err)
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: generatedRes,
	})

	return nil
}

func (h *commandHandler) sendGeneratingMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "generating...",
		},
	}); err != nil {
		return err
	}

	return nil
}

func (h *commandHandler) getImageType(
	attachment *discordgo.MessageAttachment,
) (string, error) {
	parts := strings.Split(attachment.ContentType, "/")
	if len(parts) == 2 && parts[1] != "" {
		return parts[1], nil
	} else {
		return "", fmt.Errorf("unknown content type: %s", attachment.ContentType)
	}
}

func (h *commandHandler) downloadImage(
	imageUrl string,
	filename string,
	imageType string,
) (string, error) {
	res, err := http.Get(imageUrl)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status %s", res.StatusCode)
	}

	filepath := fmt.Sprintf("images/%s.%s", filename, imageType)
	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %v", err)
	}

	if _, err := io.Copy(out, res.Body); err != nil {
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}

	return filepath, nil
}

func (h *commandHandler) generateChatResponse(
	inputText string,
	historyParts []genai.Part,
) (string, error) {
	prompt := append(historyParts, genai.Text(inputText))

	ctx := context.Background()
	res, err := h.textModel.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("failed to generate content from gemini model: %v", err)
	}
	fmt.Printf("%+v\n", res)

	generatedResponse := fmt.Sprintf("%v", res.Candidates[0].Content.Parts[0])
	return generatedResponse, nil
}

func (h *commandHandler) generatePhotoResponse(
	imageType string,
	imagePath string,
	inputText string,
) (string, error) {
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image data from file %s: %v", imagePath, err)
	}

	prompt := []genai.Part{
		genai.ImageData(imageType, imgData),
		genai.Text(inputText),
	}

	ctx := context.Background()
	res, err := h.imageModel.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("failed to generate content from gemini model: %v", err)
	}
	fmt.Printf("%+v\n", res)

	generatedResponse := fmt.Sprintf("%v", res.Candidates[0].Content.Parts[0])
	return generatedResponse, nil
}

func newCommandHandler(
	textModel *genai.GenerativeModel,
	imageModel *genai.GenerativeModel,
  chatSessRepository chat_session.ChatSessionRepository,
) *commandHandler {
	ch := &commandHandler{
		interactionCreateHandlerMap: make(map[string]func(
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
			optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
		) error),
		messageCreateHandlerMap: make(map[string]func(
			s *discordgo.Session,
			m *discordgo.MessageCreate,
		) error),
		textModel:  textModel,
		imageModel: imageModel,
    chatSessRepository: chatSessRepository,
	}

	ch.messageCreateHandlerMap["gemini-chat"] = ch.mGeminiChat

	ch.interactionCreateHandlerMap["gemini-chat"] = ch.iGeminiChat
	ch.interactionCreateHandlerMap["gemini-photo"] = ch.geminiPhoto

	return ch
}
