package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Libro struct {
	ID          int
 nombre      string
cedula       string
	Año         int
	Descripcion string
	ImagenURL   string
}

var libros = []Libro{
	{
		ID: 1,
	 nombre: "Cien Años de Soledad",
	cedula: "Gabriel García Márquez",
		Año: 1967,
		Descripcion: "Una novela sobre la familia Buendía.",
		ImagenURL: "/static/images.jpeg",
	},
	{
		ID: 2,
	 nombre: "Don Quijote de la Mancha",
	cedula: "Miguel de Cervantes",
		Año: 1605,
		Descripcion: "Una historia de aventuras y locura.",
		ImagenURL: "/static/Don_Quijote_de_la_Mancha-Cervantes_Miguel-lg.png",
	},
	{
		ID: 3,
	 nombre: "La sombra del viento",
	cedula: "Carlos Ruiz Zafón",
		Año: 2001,
		Descripcion: "Misterio y literatura en la Barcelona de posguerra.",
		ImagenURL: "/static/47856_portada___201609051317.jpg",
	},
}

type DatosPagina struct {
	Libros  []Libro
	Detalle *Libro
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

	data := DatosPagina{
		Libros:  libros,
		Detalle: libroSeleccionado,
	}

	tmpl, err := template.ParseFiles("base.html", "index.html")
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

func RegistrarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("base.html", "registrar.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.ExecuteTemplate(w, "base", nil)
		return
	}

	if r.Method == http.MethodPost {
	 nombre := r.FormValue( "nombre")
	cedula := r.FormValue("cedula")
		anio := r.FormValue("año")
	

		log.Println("Usuario registrado:")
		log.Println("Nombre:", nombre)
		log.Println("Cedula:",cedula)
		log.Println("fecha de nacimiento:", anio)
		

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func main() {
	// Manejar archivos estáticos (como imágenes locales)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Rutas principales
	http.HandleFunc("/", Index)
	http.HandleFunc("/registrar", RegistrarHandler)

	log.Println("Servidor corriendo en http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
