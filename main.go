
// ======= main.go =======
package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Libro struct {
	ID          int
	Nombre      string
	Autor       string
	Ano         int
	Descripcion string
	ImagenURL   string
}

var libros = []Libro{
	{ID: 1, Nombre: "Cien Años de Soledad", Autor: "Gabriel García Márquez", Ano: 1967, Descripcion: "Una novela sobre la familia Buendía.", ImagenURL: "/static/images.jpeg"},
	{ID: 2, Nombre: "Don Quijote de la Mancha", Autor: "Miguel de Cervantes", Ano: 1605, Descripcion: "Una historia de aventuras y locura.", ImagenURL: "/static/Don_Quijote_de_la_Mancha-Cervantes_Miguel-lg.png"},
	{ID: 3, Nombre: "La sombra del viento", Autor: "Carlos Ruiz Zafón", Ano: 2001, Descripcion: "Misterio y literatura en la Barcelona de posguerra.", ImagenURL: "/static/47856_portada___201609051317.jpg"},
}

type DatosPagina struct {
	Libros  []Libro
	Detalle *Libro
	Año     int
	Usuario string
}

func Index(w http.ResponseWriter, r *http.Request) {
	var libroSeleccionado *Libro
	idStr := r.URL.Query().Get("id")
	if idStr != "" {
		if id, err := strconv.Atoi(idStr); err == nil {
			for _, libro := range libros {
				if libro.ID == id {
					libroSeleccionado = &libro
					break
				}
			}
		}
	}

	usuario := ""
	if cookie, err := r.Cookie("usuario"); err == nil {
		usuario = cookie.Value
	}

	data := DatosPagina{
		Libros:  libros,
		Detalle: libroSeleccionado,
		Año:     time.Now().Year(),
		Usuario: usuario,
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

	log.Println("✅ Conexión con Firebase Firestore exitosa")
	log.Println("Servidor corriendo en http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
