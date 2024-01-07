package discord

import (
	"log"
	"os"
	"os/signal"

	"gemini-discord.mashu.idv.tw/repository/chat_session"
	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
)

func Setup(
	botToken string,
	textModel *genai.GenerativeModel,
	imageModel *genai.GenerativeModel,
	chatSessRepository chat_session.ChatSessionRepository,
) {
	s, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	log.Println("Discord session created")

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	ch := newCommandHandler(textModel, imageModel, chatSessRepository)
	mh := newMessageHandler(textModel, chatSessRepository)
	s.AddHandler(interactionCreateHandler(ch))
	s.AddHandler(slashCommandMessageCreateHandler(ch))
	s.AddHandler(normalMessageCreateHandler(mh))

	if err := s.Open(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	log.Println("Discord session opened")

	defer s.Close()

	registerCommands(s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Ctrl+C pressed")

	removeCommands(s)
}

func interactionCreateHandler(
	ch *commandHandler,
) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Printf("[InteractionCreate]\n")

		options := i.ApplicationCommandData().Options
		optionMap := make(
			map[string]*discordgo.ApplicationCommandInteractionDataOption,
			len(options),
		)
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}

		if h, ok := ch.interactionCreateHandlerMap[i.ApplicationCommandData().Name]; ok {
			if err := h(s, i, optionMap); err != nil {
				log.Printf("something goes wrong while handling command from InteractionCreate event: %v\n", err)

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "something goes wrong",
					},
				})
			}
		}
	}
}

func slashCommandMessageCreateHandler(
	ch *commandHandler,
) func(*discordgo.Session, *discordgo.MessageCreate) {
	isDefinedSlashCommand := func(m *discordgo.MessageCreate) bool {
		return m.Interaction != nil && m.Interaction.Type == discordgo.InteractionApplicationCommand
	}

	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Printf("[SlashCommandMessageCreate] SenderID: %+v, Channel: %s, Content: %s\n", m.Author.ID, m.ChannelID, m.Content)

		if isDefinedSlashCommand(m) {
			if h, ok := ch.messageCreateHandlerMap[m.Interaction.Name]; ok {
				if err := h(s, m); err != nil {
					log.Printf("[SlashCommandMessageCreate] something goes wrong while handling command from SlashMessageCreate event: %v\n", err)
				}
			}
		} else {
      log.Printf("[SlashCommandMessageCreate] message is not a slash command, ignored.\n")
    }
	}
}

func normalMessageCreateHandler(
	mh *messageHandler,
) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		log.Printf("[NormalMessageCreate] Sender: %s, Channel: %s\n", m.Author.Username, m.ChannelID)

		if m.Author.Bot {
      log.Printf("[NormalMessageCreate] message is from a bot, ignored.\n")
			return
		}

    if sess, yes := mh.isGeminiChatSession(m); yes {
      if err := mh.geminiResponse(s, m, sess); err != nil {
        log.Printf("[NormalMessageCreate] something goes wrong while handling message from NormalMessageCreate event: %v\n", err)
      }
    } else {
      log.Printf("[NormalMessageCreate] unknown chat session, ignored.\n")
    }
	}
}

func threadDeleteHandler(
  mh *messageHandler,
) func(*discordgo.Session, *discordgo.ThreadDelete) {
  return func(s *discordgo.Session, t *discordgo.ThreadDelete) {
    log.Printf("[ThreadDelete] ThreadID: %s\n", t.ID)

    if err := mh.chatSessRepository.DeleteById(t.ID); err != nil {
      log.Printf("[ThreadDelete] something goes wrong while handling ThreadDelete event: %v\n", err)
    }
  }
}
