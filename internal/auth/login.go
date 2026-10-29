package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"COMUNITARIO/internal/modelos"
)

// CuentaRepo: lo que el login necesita de la base de datos.
type CuentaRepo interface {
	BuscarCuenta(login string) (*modelos.Cuenta, error) // con Persona cargada
}

type Manejador struct {
	Cuentas CuentaRepo
	Secreto []byte
	Vida    time.Duration
}

// POST /login → { "token": "..." }
func (m *Manejador) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if in.Usuario == "" || in.Clave == "" {
		http.Error(w, "usuario y clave son obligatorios", http.StatusUnprocessableEntity)
		return
	}
	cuenta, err := m.Cuentas.BuscarCuenta(in.Usuario)
	// Mismo mensaje si no existe o si la clave falla: no se revela cuál.
	if err != nil || !ClaveCorrecta(in.Clave, cuenta.ClaveHash) {
		http.Error(w, "Usuario o clave incorrectos", http.StatusUnauthorized)
		return
	}
	token, err := Firmar(Claims{
		CuentaID: cuenta.ID, UsuarioID: cuenta.UsuarioID, Login: cuenta.Login,
		Rol: cuenta.Persona.Rol, Expira: time.Now().Add(m.Vida).Unix(),
	}, m.Secreto)
	if err != nil {
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
