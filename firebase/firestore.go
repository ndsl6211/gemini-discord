package firebase

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

func NewFirestoreClient(projectId string) *firestore.Client {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectId}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

  return client
}
