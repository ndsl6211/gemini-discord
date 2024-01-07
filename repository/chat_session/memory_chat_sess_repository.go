package chat_session

import (
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type memoryChatSessionRepository struct {
  chatSessionMap map[string]*genai.ChatSession
}

func (r *memoryChatSessionRepository) GetById(id string) (*genai.ChatSession, error) {
  if session, ok := r.chatSessionMap[id]; ok {
    return session, nil
  }

	return nil, fmt.Errorf("chat session of id %s not found", id)
}

func (r *memoryChatSessionRepository) Save(id string, session *genai.ChatSession) error {
  r.chatSessionMap[id] = session
  
	return nil
}

func (r *memoryChatSessionRepository) DeleteById(id string) error {
  delete(r.chatSessionMap, id)
  
  return nil
}

func NewMemoryChatSessionRepository() ChatSessionRepository {
	return &memoryChatSessionRepository{
    chatSessionMap: make(map[string]*genai.ChatSession, 0),
  }
}
