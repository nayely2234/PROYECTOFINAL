package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
	ID    string
	Libro string
	Fecha string
}

func InicioHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	iter := FirestoreClient.Collection("libro").Documents(ctx)

	var libros []Libro
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var libro Libro
		if err := doc.DataTo(&libro); err != nil {
			log.Println("Error map Libro:", err)
			continue
		}

		libros = append(libros, libro)
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

	renderTemplate(w, r, "index.html", data)
}

// Libro representa un libro en la colección de Firestore
func MisPrestamosHandler(w http.ResponseWriter, r *http.Request) {
	usuario := ""
	rol := ""

	// ✅ Lógica básica para obtener cookies
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

	// ✅ Trae SOLO préstamos del usuario logueado
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

	data := struct {
		Prestamos []PrestamoListado
		Año       int
		Usuario   string
		Rol       string
		Query     map[string]string
	}{
		Prestamos: prestamos,
		Año:       time.Now().Year(),
		Usuario:   usuario,
		Rol:       rol,
		Query: map[string]string{
			"error":   r.URL.Query().Get("error"),
			"success": r.URL.Query().Get("success"),
		},
	}

	tmpl, err := template.New("mis_prestamos.html").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).ParseFiles("templates/base.html", "templates/mis_prestamos.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Redirect(w, r, "/mis-prestamos?error=incompleto", http.StatusSeeOther)
		return
	}

	// ✅ Verifica límite de 15 días máximo
	fechaNueva, err := time.Parse("2006-01-02", fecha)
	if err != nil {
		http.Redirect(w, r, "/mis-prestamos?error=fecha", http.StatusSeeOther)
		return
	}

	if fechaNueva.After(time.Now().AddDate(0, 0, 15)) {
		http.Redirect(w, r, "/mis-prestamos?error=max_15_dias", http.StatusSeeOther)
		return
	}

	// ✅ Trae el préstamo actual para saber el libro
	docSnap, err := FirestoreClient.Collection("prestamos").Doc(id).Get(r.Context())
	if err != nil {
		http.Redirect(w, r, "/mis-prestamos?error=no_encontrado", http.StatusSeeOther)
		return
	}

	libro := docSnap.Data()["libro"].(string)

	// ✅ Verifica si YA existe otro préstamo con el MISMO libro y fecha
	iter := FirestoreClient.Collection("prestamos").
		Where("libro", "==", libro).
		Where("fecha", "==", fecha).
		Documents(r.Context())

	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		if doc.Ref.ID != id {
			// Existe otro préstamo con la misma fecha
			http.Redirect(w, r, "/mis-prestamos?error=ocupado", http.StatusSeeOther)
			return
		}
	}

	// ✅ Si todo bien, actualiza fecha
	_, err = FirestoreClient.Collection("prestamos").Doc(id).Update(r.Context(), []firestore.Update{
		{Path: "fecha", Value: fecha},
	})

	if err != nil {
		http.Redirect(w, r, "/mis-prestamos?error=update_fail", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/mis-prestamos?success=editado", http.StatusSeeOther)
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

		// ✅ SOLO libros DISPONIBLES
		iter := FirestoreClient.Collection("libro").Where("disponible", "==", true).Documents(ctx)

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
			Query   map[string]string
		}{
			Libros:  libros,
			Usuario: usuario,
			Rol:     rol,
			Año:     time.Now().Year(),
			Query: map[string]string{
				"error": r.URL.Query().Get("error"),
			},
		}

		renderTemplate(w, r, "prestamos.html", data)
		return
	}

	if r.Method == http.MethodPost {
		libro := r.FormValue("libro")
		fecha := r.FormValue("fecha")

		if usuario == "" || libro == "" || fecha == "" {
			http.Redirect(w, r, "/prestamos?error=incompleto", http.StatusSeeOther)
			return
		}

		ctx := context.Background()

		// ✅ Verifica que el libro siga disponible
		iter := FirestoreClient.Collection("libro").
			Where("nombre", "==", libro).
			Where("disponible", "==", true).
			Documents(ctx)

		docSnap, err := iter.Next()
		if err != nil {
			// No hay libro disponible → ya está prestado
			http.Redirect(w, r, "/prestamos?error=no_disponible", http.StatusSeeOther)
			return
		}

		// ✅ Marca como NO disponible
		_, err = docSnap.Ref.Update(ctx, []firestore.Update{
			{Path: "disponible", Value: false},
		})
		if err != nil {
			http.Error(w, "Error al actualizar disponibilidad", http.StatusInternalServerError)
			log.Println("❌ Error al actualizar disponibilidad:", err)
			return
		}

		// ✅ Crea el préstamo
		prestamo := Prestamo{
			Usuario: usuario,
			Libro:   libro,
			Fecha:   fecha,
		}

		_, _, err = FirestoreClient.Collection("prestamos").Add(ctx, prestamo)
		if err != nil {
			http.Error(w, "Error al registrar el préstamo", http.StatusInternalServerError)
			log.Println("Error al guardar préstamo:", err)
			return
		}

		log.Println("✅ Préstamo registrado:", usuario, "→", libro)
		http.Redirect(w, r, "/mis-prestamos?success=1", http.StatusSeeOther)
		return
	}

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
			Query   map[string]string
		}{
			Usuario: usuario,
			Rol:     rol,
			Año:     time.Now().Year(),
			Query: map[string]string{
				"error":   r.URL.Query().Get("error"),
				"success": r.URL.Query().Get("success"),
			},
		}
		renderTemplate(w, r, "registrar_libro.html", data)
		return
	}

	if r.Method == http.MethodPost {
		nombre := r.FormValue("nombre")
		autor := r.FormValue("autor")
		descripcion := r.FormValue("descripcion")
		imagen := r.FormValue("imagen")
		anoStr := r.FormValue("ano")

		ano, err := strconv.Atoi(anoStr)
		if err != nil || ano < 0 {
			http.Redirect(w, r, "/registrar-libro?error=ano", http.StatusSeeOther)
			return
		}

		// ✅ Verifica si YA existe
		iter := FirestoreClient.Collection("libro").Where("nombre", "==", nombre).Documents(r.Context())
		_, err = iter.Next()
		if err == nil {
			// Existe → redirige con error
			http.Redirect(w, r, "/registrar-libro?error=existente", http.StatusSeeOther)
			return
		}

		doc := map[string]interface{}{
			"nombre":      nombre,
			"autor":       autor,
			"descripcion": descripcion,
			"ano":         ano,
			"imagen":      imagen,
			"disponible":  true,
		}

		_, _, err = FirestoreClient.Collection("libro").Add(r.Context(), doc)
		if err != nil {
			http.Error(w, "Error al registrar libro", http.StatusInternalServerError)
			log.Println("Error Firestore libro:", err)
			return
		}

		log.Println("✅ Libro registrado:", nombre)
		http.Redirect(w, r, "/registrar-libro?success=1", http.StatusSeeOther)
	}
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

	// ✅ Obtiene el préstamo
	doc, err := FirestoreClient.Collection("prestamos").Doc(id).Get(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo préstamo", http.StatusInternalServerError)
		return
	}

	libroNombre, ok := doc.Data()["libro"].(string)
	if !ok || libroNombre == "" {
		http.Error(w, "Datos del libro inválidos", http.StatusBadRequest)
		return
	}

	// ✅ Elimina el préstamo
	_, err = FirestoreClient.Collection("prestamos").Doc(id).Delete(r.Context())
	if err != nil {
		http.Error(w, "Error eliminando préstamo", http.StatusInternalServerError)
		return
	}

	// ✅ Marca el libro como disponible otra vez
	iter := FirestoreClient.Collection("libro").Where("nombre", "==", libroNombre).Documents(r.Context())
	for {
		libroDoc, err := iter.Next()
		if err != nil {
			break
		}
		_, err = libroDoc.Ref.Update(r.Context(), []firestore.Update{
			{Path: "disponible", Value: true},
		})
		if err != nil {
			log.Println("❌ Error al actualizar disponibilidad:", err)
		}
	}

	http.Redirect(w, r, "/mis-prestamos", http.StatusSeeOther)
}

func renderTemplate(w http.ResponseWriter, _ *http.Request, archivo string, data interface{}) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+archivo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error cargando plantilla:", archivo, err)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}
func BuscarHandler(w http.ResponseWriter, r *http.Request) {
	usuario := ""
	rol := ""
	if c, err := r.Cookie("usuario"); err == nil {
		usuario = c.Value
	}
	if c, err := r.Cookie("rol"); err == nil {
		rol = c.Value
	}

	// Obtener query de búsqueda
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx := context.Background()

	// Busca libros cuyo nombre contenga la consulta (solo exacto o prefijo)
	iter := FirestoreClient.Collection("libro").Where("nombre", ">=", query).Where("nombre", "<=", query+"\uf8ff").Documents(ctx)

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

	// Datos para la plantilla
	data := struct {
		Query   string
		Libros  []Libro
		Usuario string
		Rol     string
		Año     int
	}{
		Query:   query,
		Libros:  libros,
		Usuario: usuario,
		Rol:     rol,
		Año:     time.Now().Year(),
	}

	renderTemplate(w, r, "resultados.html", data)
}
