package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const iteraciones = 120000

// HashClave guarda la clave como "pbkdf2$iteraciones$sal$hash" (nunca en claro).
func HashClave(clave string) (string, error) {
	sal := make([]byte, 16)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	h, err := pbkdf2.Key(sha256.New, clave, sal, iteraciones, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2$%d$%s$%s", iteraciones, enc.EncodeToString(sal), base64.RawStdEncoding.EncodeToString(h)), nil
}

// ClaveCorrecta compara en tiempo constante.
func ClaveCorrecta(clave, guardada string) bool {
	p := strings.Split(guardada, "$")
	if len(p) != 4 || p[0] != "pbkdf2" {
		return false
	}
	it, err := strconv.Atoi(p[1])
	if err != nil {
		return false
	}
	sal, err := enc.DecodeString(p[2])
	if err != nil {
		return false
	}
	esperado, err := base64.RawStdEncoding.DecodeString(p[3])
	if err != nil {
		return false
	}
	h, err := pbkdf2.Key(sha256.New, clave, sal, it, len(esperado))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(h, esperado) == 1
}
