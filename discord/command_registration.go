package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	registeredCommands []*discordgo.ApplicationCommand
)

func registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
    {
      Name: "gemini-chat",
      Description: "chat with gemini WITHOUT context",
    },
		{
			Name:        "gemini-photo",
			Description: "give an image, and return the description of this image",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "text",
					Description: "input text",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "image",
					Description: "input image",
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Required:    true,
				},
			},
		},
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	log.Println("Command registered")
}

func removeCommands(s *discordgo.Session) {
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
	log.Println("Registered commands removed")
}
