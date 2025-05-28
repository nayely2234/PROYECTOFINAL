package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client

func InitFirebase() {
	ctx := context.Background()

	// Ruta al archivo de credenciales JSON
	sa := option.WithCredentialsFile("basebiblioteca-fe7d5-firebase-adminsdk-fbsvc-fa72da49bb.json")

	// Configuración explícita con el project ID
	config := &firebase.Config{
		ProjectID: "basebiblioteca-fe7d5",
	}

	// Inicializar la app con configuración y credenciales
	app, err := firebase.NewApp(ctx, config, sa)
	if err != nil {
		log.Fatalf("Error inicializando Firebase: %v", err)
	}

	// Inicializar cliente Firestore
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error inicializando Firestore: %v", err)
	}

	FirestoreClient = client
	log.Println("✅ Conexión con Firebase Firestore exitosa")
}
