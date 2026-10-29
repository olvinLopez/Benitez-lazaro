package api_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"COMUNITARIO/internal/api"
	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/comentarios/handler"
	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/pruebas"
	"time"
)

var secreto = []byte("secreto-de-prueba")

func servidor(t *testing.T) http.Handler {
	t.Helper()
	// Clave guardada con el mismo hash real que usa la semilla.
	hash := func(c string) string { h, _ := auth.HashClave(c); return h }
	f := pruebas.Nuevo(hash)
	h := &handler.Manejador{Servicio: servicio.Nuevo(f)}
	a := &auth.Manejador{Cuentas: f, Secreto: secreto, Vida: time.Hour}
	return api.Router(h, a, secreto)
}

func enviar(srv http.Handler, metodo, ruta, token, cuerpo string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	return w
}

func login(t *testing.T, srv http.Handler, usuario, clave string) string {
	t.Helper()
	w := enviar(srv, "POST", "/login", "", `{"usuario":"`+usuario+`","clave":"`+clave+`"}`)
	if w.Code != 200 {
		t.Fatalf("login %s: %d %s", usuario, w.Code, w.Body)
	}
	var r struct{ Token string }
	json.NewDecoder(w.Body).Decode(&r)
	return r.Token
}

// Recorre las seis peticiones del Hito 2 contra la API completa.
func TestLasSeisPeticiones(t *testing.T) {
	srv := servidor(t)

	// Petición 0: crear registro nuevo (queda pendiente).
	jose := login(t, srv, "jose", "jose-2026")
	w := enviar(srv, "POST", "/comentarios", jose, `{"texto":"nuevo para la demo"}`)
	if w.Code != 201 {
		t.Fatalf("petición 0: %d", w.Code)
	}
	var nuevo struct{ ID int }
	json.NewDecoder(w.Body).Decode(&nuevo)
	ruta := "/comentarios/" + itoa(nuevo.ID) + "/estado"

	// 1. Sin token → 401
	if c := enviar(srv, "GET", "/comentarios", "", "").Code; c != 401 {
		t.Errorf("petición 1: %d", c)
	}
	// 2. Login → 200 y token (login hecho arriba); 3. Con token → 200
	if c := enviar(srv, "GET", "/comentarios", jose, "").Code; c != 200 {
		t.Errorf("petición 3: %d", c)
	}
	// 4. Miembro intenta cambiar estado → 403
	if c := enviar(srv, "PATCH", ruta, jose, `{"estado":"aprobado"}`).Code; c != 403 {
		t.Errorf("petición 4: %d", c)
	}
	// 5. Moderador, mismo cambio → 200
	luis := login(t, srv, "luis", "luis-2026")
	if c := enviar(srv, "PATCH", ruta, luis, `{"estado":"aprobado"}`).Code; c != 200 {
		t.Errorf("petición 5: %d", c)
	}
	// 6. Transición prohibida (comentario 1 está destacado → pendiente) → 422
	if c := enviar(srv, "PATCH", "/comentarios/1/estado", luis, `{"estado":"pendiente"}`).Code; c != 422 {
		t.Errorf("petición 6: %d", c)
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

func TestLoginMalo(t *testing.T) {
	srv := servidor(t)
	if c := enviar(srv, "POST", "/login", "", `{"usuario":"jose","clave":"mala"}`).Code; c != 401 {
		t.Errorf("clave mala: %d", c)
	}
	if c := enviar(srv, "POST", "/login", "", `{"usuario":"","clave":""}`).Code; c != 422 {
		t.Errorf("vacío: %d", c)
	}
	if c := enviar(srv, "POST", "/login", "", `no es json`).Code; c != 400 {
		t.Errorf("json malo: %d", c)
	}
}

// Si alguien cambia el rol dentro del token, la firma falla → 401.
func TestTokenManipuladoSeRechaza(t *testing.T) {
	srv := servidor(t)
	jose := login(t, srv, "jose", "jose-2026")
	partes := strings.Split(jose, ".")
	falso := auth.Claims{CuentaID: 1, UsuarioID: 1, Login: "jose", Rol: "moderador", Expira: time.Now().Add(time.Hour).Unix()}
	cuerpo, _ := json.Marshal(falso)
	import64 := func(b []byte) string {
		return strings.TrimRight(strings.NewReplacer("+", "-", "/", "_").Replace(b64(b)), "=")
	}
	manipulado := partes[0] + "." + import64(cuerpo) + "." + partes[2]
	if c := enviar(srv, "PATCH", "/comentarios/2/estado", manipulado, `{"estado":"aprobado"}`).Code; c != 401 {
		t.Errorf("token manipulado: %d", c)
	}
}

func TestTokenVencido(t *testing.T) {
	srv := servidor(t)
	tok, _ := auth.Firmar(auth.Claims{UsuarioID: 1, Rol: "miembro", Expira: time.Now().Add(-time.Minute).Unix()}, secreto)
	if c := enviar(srv, "GET", "/comentarios", tok, "").Code; c != 401 {
		t.Errorf("vencido: %d", c)
	}
}

func TestRutasProtegidasPorRol(t *testing.T) {
	srv := servidor(t)
	jose := login(t, srv, "jose", "jose-2026")
	if c := enviar(srv, "DELETE", "/comentarios/2", jose, "").Code; c != 403 {
		t.Errorf("miembro borra: %d", c)
	}
	luis := login(t, srv, "luis", "luis-2026")
	if c := enviar(srv, "DELETE", "/comentarios/2", luis, "").Code; c != 204 {
		t.Errorf("moderador borra: %d", c)
	}
	if c := enviar(srv, "GET", "/comentarios/2", luis, "").Code; c != 404 {
		t.Errorf("borrado: %d", c)
	}
}

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
