package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

type Persona struct {
	Nombre     string `firestore:"nombre"`
	Cedula     string `firestore:"cedula"`
	Ano        string `firestore:"ano"`
	Contrasena string `firestore:"contrasena"`
	Rol        string `firestore:"rol"`
}

type Prestamo struct {
	Usuario string `firestore:"usuario"`
	Libro   string `firestore:"libro"`
	Fecha   string `firestore:"fecha"`
}
type PrestamoListado struct {
	ID     string
	Libro  string
	Fecha  string
}

// MisPrestamosHandler maneja las solicitudes HTTP para mostrar los préstamos asociados a un usuario.
// Extrae la información del usuario desde la solicitud y responde con los datos correspondientes.
// Parámetros:
//   - w: http.ResponseWriter para enviar la respuesta HTTP.
//   - r: *http.Request que contiene los detalles de la solicitud HTTP.
func MisPrestamosHandler(w http.ResponseWriter, r *http.Request) {
	usuario := ""
	rol := ""
	if c, err := r.Cookie("usuario"); err == nil {
		usuario = c.Value
	}
	if c, err := r.Cookie("rol"); err == nil {
		rol = c.Value
	}
	if usuario == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx := context.Background()
	iter := FirestoreClient.Collection("prestamos").Where("usuario", "==", usuario).Documents(ctx)
	var prestamos []PrestamoListado
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		data := doc.Data()
		libro, _ := data["libro"].(string)
		fecha, _ := data["fecha"].(string)
		prestamos = append(prestamos, PrestamoListado{
			ID:    doc.Ref.ID,
			Libro: libro,
			Fecha: fecha,
		})
	}
	tmpl, err := template.New("mis_prestamos.html").Funcs(template.FuncMap{"safe": func(s string) template.HTML { return template.HTML(s) }}).ParseFiles("templates/base.html", "templates/mis_prestamos.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Prestamos []PrestamoListado
		Año       int
		Usuario   string
		Rol       string
	}{
		Prestamos: prestamos,
		Año:       time.Now().Year(),
		Usuario:   usuario,
		Rol:       rol,
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func EditarPrestamoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	fecha := r.FormValue("fecha")
	if id == "" || fecha == "" {
		http.Error(w, "Datos incompletos", http.StatusBadRequest)
		return
	}
	_, err := FirestoreClient.Collection("prestamos").Doc(id).Update(r.Context(), []firestore.Update{
		{Path: "fecha", Value: fecha},
	})
	if err != nil {
		http.Error(w, "Error actualizando préstamo", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/mis-prestamos", http.StatusSeeOther)
}

func DevolverPrestamoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	_, err := FirestoreClient.Collection("prestamos").Doc(id).Delete(r.Context())
	if err != nil {
		http.Error(w, "Error eliminando préstamo", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/mis-prestamos", http.StatusSeeOther)
}



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
		"rol":        "usuario",
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/base.html", "templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.ExecuteTemplate(w, "base", nil)
		return
	}

	if r.Method == http.MethodPost {
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

		doc, err := iter.Next()
		if err != nil {
			http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
			return
		}

		rol := "usuario"
		if rdoc := doc.Data()["rol"]; rdoc != nil {
			if val, ok := rdoc.(string); ok {
				rol = val
			}
		}

		http.SetCookie(w, &http.Cookie{Name: "usuario", Value: nombre, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "rol", Value: rol, Path: "/"})

		log.Println("✅ Sesión iniciada:", nombre, "| Rol:", rol)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "usuario", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "rol", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
	rol := ""
	if c, err := r.Cookie("usuario"); err == nil {
		usuario = c.Value
	}
	if c, err := r.Cookie("rol"); err == nil {
		rol = c.Value
	}

	data := struct {
		Personas []Persona
		Año      int
		Usuario  string
		Rol      string
	}{
		Personas: personas,
		Año:      time.Now().Year(),
		Usuario:  usuario,
		Rol:      rol,
	}

	renderTemplate(w, r, "personas.html", data)
}

func PrestamoHandler(w http.ResponseWriter, r *http.Request) {
	usuario := ""
	rol := ""
	if c, err := r.Cookie("usuario"); err == nil {
		usuario = c.Value
	}
	if c, err := r.Cookie("rol"); err == nil {
		rol = c.Value
	}
	if usuario == "" {
		http.Error(w, "⚠️ Debes iniciar sesión para registrar un préstamo", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		ctx := context.Background()
		iter := FirestoreClient.Collection("libro").Documents(ctx)

		var libros []Libro
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}
			var libro Libro
			doc.DataTo(&libro)
			libros = append(libros, libro)
		}

		data := struct {
			Libros  []Libro
			Usuario string
			Rol     string
			Año     int
		}{
			Libros:  libros,
			Usuario: usuario,
			Rol:     rol,
			Año:     time.Now().Year(),
		}

		renderTemplate(w, r, "prestamos.html", data)
		return
	}

	if r.Method == http.MethodPost {
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
			http.Error(w, "Error al registrar el préstamo", http.StatusInternalServerError)
			log.Println("Error al guardar préstamo:", err)
			return
		}

		log.Println("✅ Préstamo registrado:", usuario, "→", libro)
		http.Redirect(w, r, "/mis-prestamos", http.StatusSeeOther)
		return
	}

	// Método no soportado
	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}


func RegistrarLibroHandler(w http.ResponseWriter, r *http.Request) {
	usuario := ""
	rol := ""
	if c, err := r.Cookie("usuario"); err == nil {
		usuario = c.Value
	}
	if c, err := r.Cookie("rol"); err == nil {
		rol = c.Value
	}

	if r.Method == http.MethodGet {
		data := struct {
			Usuario string
			Rol     string
			Año     int
		}{
			Usuario: usuario,
			Rol:     rol,
			Año:     time.Now().Year(),
		}

		renderTemplate(w, r, "registrar_libro.html", data)
		return
	}

	// POST lógico aquí...
}


func renderTemplate(w http.ResponseWriter, r *http.Request, archivo string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+archivo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error cargando plantilla:", archivo, err)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}
