package discord

import (
	"context"
	"fmt"

	"gemini-discord.mashu.idv.tw/repository/chat_session"
	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
)

type messageHandler struct {
	textModel *genai.GenerativeModel

	chatSessRepository chat_session.ChatSessionRepository
}

func (h *messageHandler) isGeminiChatSession(m *discordgo.MessageCreate) (*genai.ChatSession, bool) {
	sess, err := h.chatSessRepository.GetById(m.ChannelID)
	if err != nil || sess == nil {
		return nil, false
	}

	return sess, true
}

func (h *messageHandler) geminiResponse(
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	sess *genai.ChatSession,
) error {
	ctx := context.Background()

  if err := s.ChannelTyping(m.ChannelID); err != nil {
    return fmt.Errorf("failed to send typing indicator: %v", err)
  }

	generatedRes, err := sess.SendMessage(
    ctx,
    genai.Text(fmt.Sprintf("[%s]: %s", m.Author.Username, m.Content)),
  )
	if err != nil {
		return fmt.Errorf("failed to generate response: %+v", err.Error())
	}

	if _, err = s.ChannelMessageSend(
    m.ChannelID,
    fmt.Sprintf("%s", generatedRes.Candidates[0].Content.Parts[0]),
  ); err != nil {
		return fmt.Errorf("failed to send message to channel %s", m.ChannelID)
	}

	return nil
}

func newMessageHandler(
	textModel *genai.GenerativeModel,
	chatSessRepository chat_session.ChatSessionRepository,
) *messageHandler {
	return &messageHandler{
		textModel:          textModel,
		chatSessRepository: chatSessRepository,
	}
}
