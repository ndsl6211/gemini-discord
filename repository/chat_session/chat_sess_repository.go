package chat_session

import "github.com/google/generative-ai-go/genai"

type ChatSessionRepository interface {
  GetById(id string) (*genai.ChatSession, error)
  Save(id string, messages *genai.ChatSession) error
  DeleteById(id string) error
}
