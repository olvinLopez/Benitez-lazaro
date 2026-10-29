// Package auth: inicio de sesión, token firmado y protección de rutas.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrTokenInvalido = errors.New("token inválido o vencido")

// Claims es lo que lleva el token por dentro.
type Claims struct {
	CuentaID  uint   `json:"cuenta_id"`
	UsuarioID uint   `json:"usuario_id"`
	Login     string `json:"login"`
	Rol       string `json:"rol"`
	Expira    int64  `json:"exp"`
}

var enc = base64.RawURLEncoding

const cabecera = `{"alg":"HS256","typ":"JWT"}`

func firma(datos string, secreto []byte) string {
	h := hmac.New(sha256.New, secreto)
	h.Write([]byte(datos))
	return enc.EncodeToString(h.Sum(nil))
}

// Firmar crea un JWT (HS256) con el secreto del servidor.
func Firmar(c Claims, secreto []byte) (string, error) {
	cuerpo, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	datos := enc.EncodeToString([]byte(cabecera)) + "." + enc.EncodeToString(cuerpo)
	return datos + "." + firma(datos, secreto), nil
}

// Verificar comprueba firma y vencimiento. Si alguien cambia el rol dentro
// del token, la firma ya no coincide y falla.
func Verificar(token string, secreto []byte, ahora time.Time) (*Claims, error) {
	p := strings.Split(token, ".")
	if len(p) != 3 {
		return nil, ErrTokenInvalido
	}
	esperada := firma(p[0]+"."+p[1], secreto)
	if subtle.ConstantTimeCompare([]byte(esperada), []byte(p[2])) != 1 {
		return nil, ErrTokenInvalido
	}
	cuerpo, err := enc.DecodeString(p[1])
	if err != nil {
		return nil, ErrTokenInvalido
	}
	var c Claims
	if err := json.Unmarshal(cuerpo, &c); err != nil {
		return nil, ErrTokenInvalido
	}
	if ahora.Unix() >= c.Expira {
		return nil, ErrTokenInvalido
	}
	return &c, nil
}
