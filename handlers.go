// ======= handlers.go =======
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"
)

type Persona struct {
	Nombre     string `firestore:"nombre"`
	Cedula     string `firestore:"cedula"`
	Ano        string `firestore:"ano"`
	Contrasena string `firestore:"contrasena"`
}

type Prestamo struct {
	Usuario string `firestore:"usuario"`
	Libro   string `firestore:"libro"`
	Fecha   string `firestore:"fecha"`
}

// ===================== REGISTRAR =====================
func RegistrarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "registrar.html", nil)
		return
	}

	nombre := r.FormValue("nombre")
	cedula := r.FormValue("cedula")
	ano := r.FormValue("ano")
	contrasena := r.FormValue("contrasena")

	if nombre == "" || cedula == "" || ano == "" || contrasena == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	doc := map[string]interface{}{
		"nombre":     nombre,
		"cedula":     cedula,
		"ano":        ano,
		"contrasena": contrasena,
	}

	_, _, err := FirestoreClient.Collection("persona").Add(context.Background(), doc)
	if err != nil {
		http.Error(w, "Error al guardar en Firebase", http.StatusInternalServerError)
		log.Println("Error al registrar usuario:", err)
		return
	}

	log.Println("✅ Usuario registrado:", nombre)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ===================== LOGIN =====================
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "login.html", nil)
		return
	}

	nombre := r.FormValue("nombre")
	contrasena := r.FormValue("contrasena")

	if nombre == "" || contrasena == "" {
		http.Error(w, "Campos requeridos", http.StatusBadRequest)
		return
	}

	iter := FirestoreClient.Collection("persona").
		Where("nombre", "==", nombre).
		Where("contrasena", "==", contrasena).
		Documents(r.Context())

	_, err := iter.Next()
	if err != nil {
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "usuario",
		Value: nombre,
		Path:  "/",
	})

	log.Println("✅ Sesión iniciada:", nombre)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "usuario",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ===================== LISTAR USUARIOS =====================
func PersonasHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	iter := FirestoreClient.Collection("persona").Documents(ctx)

	var personas []Persona
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var p Persona
		doc.DataTo(&p)
		personas = append(personas, p)
	}

	usuario := ""
	if cookie, err := r.Cookie("usuario"); err == nil {
		usuario = cookie.Value
	}

	data := struct {
		Personas []Persona
		Año      int
		Usuario  string
	}{
		Personas: personas,
		Año:      time.Now().Year(),
		Usuario:  usuario,
	}

	renderTemplate(w, r, "personas.html", data)
}

// ===================== REGISTRAR PRÉSTAMO =====================
func PrestamoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "prestamos.html", nil)
		return
	}

	usuario := r.FormValue("usuario")
	libro := r.FormValue("libro")
	fecha := r.FormValue("fecha")

	if usuario == "" || libro == "" || fecha == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	doc := Prestamo{
		Usuario: usuario,
		Libro:   libro,
		Fecha:   fecha,
	}

	_, _, err := FirestoreClient.Collection("prestamos").Add(context.Background(), doc)
	if err != nil {
		http.Error(w, "Error al guardar el préstamo", http.StatusInternalServerError)
		log.Println("Error al guardar préstamo:", err)
		return
	}

	log.Println("✅ Préstamo registrado:", usuario, "→", libro)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ===================== TEMPLATE WRAPPER =====================
func renderTemplate(w http.ResponseWriter, r *http.Request, archivo string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+archivo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error cargando plantilla:", archivo, err)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}
