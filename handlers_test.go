package main

import (
	
	"net/http"
	"net/http/httptest"
	"testing"
)



// ======= TEST: Registro sin datos =========
func TestRegistrarHandlerEmptyForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/registrar", nil)
	rr := httptest.NewRecorder()

	RegistrarHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400 pero recibió %d", rr.Code)
	}
}

// ======= TEST: Vista de login =========
func TestLoginHandlerGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado status 200 en GET /login pero recibió %d", rr.Code)
	}
}

// ======= TEST: Mis préstamos sin sesión =========
func TestMisPrestamosHandlerWithoutCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/mis-prestamos", nil)
	rr := httptest.NewRecorder()

	MisPrestamosHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("esperado redirect a login pero recibió %d", rr.Code)
	}
}
