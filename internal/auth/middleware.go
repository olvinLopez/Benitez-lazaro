package auth

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type ctxKey struct{}

// ClaimsDe devuelve quién es el usuario de la petición (lo puso Autenticar).
func ClaimsDe(r *http.Request) *Claims {
	c, _ := r.Context().Value(ctxKey{}).(*Claims)
	return c
}

// Autenticar: AQUÍ se decide que una ruta pide token. Sin token válido → 401.
func Autenticar(secreto []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(h, "Bearer ")
			if !ok || token == "" {
				http.Error(w, "Falta el token (Authorization: Bearer ...)", http.StatusUnauthorized)
				return
			}
			c, err := Verificar(token, secreto, time.Now())
			if err != nil {
				http.Error(w, "Token inválido o vencido", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, c)))
		})
	}
}

// RequiereRol: AQUÍ se decide qué rol puede entrar. Otro rol → 403.
func RequiereRol(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := ClaimsDe(r)
			if c == nil {
				http.Error(w, "No autenticado", http.StatusUnauthorized)
				return
			}
			for _, rol := range roles {
				if c.Rol == rol {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Tu rol no tiene permiso para esta acción", http.StatusForbidden)
		})
	}
}
