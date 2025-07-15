package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type Libro struct {
	ID          string `firestore:"-"`
	Nombre      string `firestore:"nombre"`
	Autor       string `firestore:"autor"`
	Descripcion string `firestore:"descripcion"`
	ImagenURL   string `firestore:"imagen"`
	Ano         int    `firestore:"ano"`
	Disponible  bool   `firestore:"disponible"`
}

type DatosPagina struct {
	Libros  []Libro
	Detalle *Libro
	Año     int
	Usuario string
	Rol     string
}

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	iter := FirestoreClient.Collection("libro").Documents(ctx)

	var libros []Libro
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		data := doc.Data()

		nombre, _ := data["nombre"].(string)
		autor, _ := data["autor"].(string)
		descripcion, _ := data["descripcion"].(string)
		imagen, _ := data["imagen"].(string)
		anoFloat, _ := data["ano"].(int64)
		ano := int(anoFloat)
		disponible, _ := data["disponible"].(bool)

		libros = append(libros, Libro{
			ID:          doc.Ref.ID,
			Nombre:      nombre,
			Autor:       autor,
			Descripcion: descripcion,
			Ano:         ano,
			ImagenURL:   imagen,
			Disponible:  disponible,
		})
	}

	var libroSeleccionado *Libro
	id := r.URL.Query().Get("id")
	for _, libro := range libros {
		if libro.ID == id {
			libroSeleccionado = &libro
			break
		}
	}

	usuario := ""
	rol := ""
	if cookie, err := r.Cookie("usuario"); err == nil {
		usuario = cookie.Value
	}
	if cookie, err := r.Cookie("rol"); err == nil {
		rol = cookie.Value
	}

	data := DatosPagina{
		Libros:  libros,
		Detalle: libroSeleccionado,
		Año:     time.Now().Year(),
		Usuario: usuario,
		Rol:     rol,
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error en plantilla:", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error al ejecutar template:", err)
	}
}

func main() {
	InitFirebase()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", Index)
	http.HandleFunc("/registrar", RegistrarHandler)
	http.HandleFunc("/personas", PersonasHandler)
	http.HandleFunc("/prestamos", PrestamoHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/registrar-libro", RegistrarLibroHandler)
	http.HandleFunc("/mis-prestamos", MisPrestamosHandler)
	http.HandleFunc("/editar-prestamo", EditarPrestamoHandler)
	http.HandleFunc("/devolver-prestamo", DevolverPrestamoHandler)
	http.HandleFunc("/buscar", BuscarHandler) // Descomenta si existe

	// 🚩 Usa el puerto dinámico de Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // fallback local
	}

	log.Println("Servidor corriendo en puerto:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
