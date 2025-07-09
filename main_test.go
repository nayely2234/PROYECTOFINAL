package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// 🔹 PRUEBA: Usuario NO logueado → PrestamoHandler devuelve 401
func TestPrestamoHandler_SinSesion(t *testing.T) {
	req := httptest.NewRequest("GET", "/prestamos", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(PrestamoHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Esperaba 401, obtuve %d", rr.Code)
	}
}

// 🔹 PRUEBA: Usuario NO logueado → MisPrestamosHandler redirige a login
func TestMisPrestamosHandler_SinSesion(t *testing.T) {
	req := httptest.NewRequest("GET", "/mis-prestamos", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(MisPrestamosHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Esperaba redirect 303, obtuve %d", rr.Code)
	}
}

// 🔹 PRUEBA: Usuario NO logueado → RegistrarLibroHandler GET carga OK
func TestRegistrarLibroHandler_Get(t *testing.T) {
	req := httptest.NewRequest("GET", "/registrar-libro", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(RegistrarLibroHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Esperaba 200 GET registrar libro, obtuve %d", rr.Code)
	}
}

// 🔹 PRUEBA: LoginHandler GET carga OK
func TestLoginHandler_Get(t *testing.T) {
	req := httptest.NewRequest("GET", "/login", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(LoginHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Esperaba 200 GET login, obtuve %d", rr.Code)
	}
}

// 🔹 PRUEBA: Simulación flujo POST registrar libro → campos vacíos
func TestRegistrarLibroHandler_Post_Incompleto(t *testing.T) {
	form := url.Values{}
	form.Add("nombre", "")
	form.Add("autor", "")
	form.Add("descripcion", "")
	form.Add("imagen", "")
	form.Add("ano", "")

	req := httptest.NewRequest("POST", "/registrar-libro", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(RegistrarLibroHandler)
	handler.ServeHTTP(rr, req)

	// Podrías esperar redirect si campos inválidos → en tu handler real manejas redirect con error param
	if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
		t.Errorf("Esperaba 303 o 400, obtuve %d", rr.Code)
	}
}
