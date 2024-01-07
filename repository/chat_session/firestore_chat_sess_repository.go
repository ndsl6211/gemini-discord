package chat_session

import (
	"cloud.google.com/go/firestore"
	"github.com/google/generative-ai-go/genai"
)

type firestoreChatSessionRepository struct {
	firestoreClient *firestore.Client
}

func (r *firestoreChatSessionRepository) GetById(id string) (*genai.ChatSession, error) {
	return nil, nil
}

func (r *firestoreChatSessionRepository) Save(id string, session *genai.ChatSession) error {
	return nil
}

func (r *firestoreChatSessionRepository) DeleteById(id string) error {
  return nil
}

func NewFirestoreChatSessionRepository(firestoreClient *firestore.Client) ChatSessionRepository {
	return &firestoreChatSessionRepository{firestoreClient}
}
