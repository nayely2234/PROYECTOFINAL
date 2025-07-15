package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client

func InitFirebase() {
	ctx := context.Background()

	// ✅ Leer directamente desde el archivo .json
	opt := option.WithCredentialsFile("basebiblioteca-fe7d5-firebase-adminsdk-fbsvc-fa72da49bb.json")

	config := &firebase.Config{
		ProjectID: "basebiblioteca-fe7d5",
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Error inicializando Firebase: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error inicializando Firestore: %v", err)
	}

	FirestoreClient = client
	log.Println("✅ Conexión con Firebase Firestore exitosa")
}
