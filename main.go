package main

import (
	"log"

	"gemini-discord.mashu.idv.tw/discord"
	"gemini-discord.mashu.idv.tw/genai"
	"gemini-discord.mashu.idv.tw/repository/chat_session"
	viperx "gemini-discord.mashu.idv.tw/viper"
	"github.com/spf13/viper"
)

func main() {
	viperx.Setup()

	botToken := viper.GetString("discord.botToken")
	geminiApiKey := viper.GetString("gemini.apiKey")
	log.Printf("botToken: %s\n", botToken)
	log.Printf("geminiApiKey: %s\n", geminiApiKey)

	genaiClient := gemini.NewGenaiClient(geminiApiKey)

	textModel := genaiClient.GenerativeModel("gemini-pro")
	multiModel := genaiClient.GenerativeModel("gemini-pro-vision")

	chatSessRepository := chat_session.NewMemoryChatSessionRepository()

	discord.Setup(botToken, textModel, multiModel, chatSessRepository)

	genaiClient.Close()
}
