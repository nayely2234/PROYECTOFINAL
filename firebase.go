package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client

func InitFirebase() {
	ctx := context.Background()

	// ✅ Leer credenciales desde variable de entorno
	credJSON := []byte(os.Getenv("FIREBASE_CREDENTIALS"))
	if len(credJSON) == 0 {
		log.Fatal("❌ Variable FIREBASE_CREDENTIALS vacía o no definida")
	}
	opt := option.WithCredentialsJSON(credJSON)

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
