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

	// ✅ Lee el JSON desde la variable de entorno
	credJSON := []byte(os.Getenv("FIREBASE_CREDENTIALS"))

	// Verifica que exista
	if len(credJSON) == 0 {
		log.Fatal("FIREBASE_CREDENTIALS no está definida o está vacía")
	}

	// Configura la opción con JSON en memoria
	opt := option.WithCredentialsJSON(credJSON)

	// Config explícita con tu Project ID (opcional)
	config := &firebase.Config{
		ProjectID: "basebiblioteca-fe7d5",
	}

	// Inicializa la App
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Error inicializando Firebase: %v", err)
	}

	// Inicializa el cliente Firestore
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error inicializando Firestore: %v", err)
	}

	FirestoreClient = client
	log.Println("✅ Conexión con Firebase Firestore exitosa")
}
