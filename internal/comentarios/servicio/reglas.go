// Reglas puras del negocio. Este archivo NO conoce HTTP ni la base de datos.
package servicio

import "COMUNITARIO/internal/modelos"

// PuntosDestacado: lo que gana el autor cuando su comentario se destaca.
const PuntosDestacado = 10

// LimiteMaximo: datos móviles limitados, nunca más de 20 por página.
const LimiteMaximo = 20

var estadosValidos = map[string]bool{
	modelos.EstadoPendiente: true,
	modelos.EstadoAprobado:  true,
	modelos.EstadoDestacado: true,
}

// transiciones: solo hacia adelante. De destacado no se sale.
var transiciones = map[string]string{
	modelos.EstadoPendiente: modelos.EstadoAprobado,
	modelos.EstadoAprobado:  modelos.EstadoDestacado,
}

// EstadoValido dice si el texto es uno de los tres estados.
func EstadoValido(e string) bool { return estadosValidos[e] }

// TransicionPermitida: pendiente→aprobado y aprobado→destacado. Nada más.
func TransicionPermitida(de, a string) bool { return transiciones[de] == a && a != "" }

// MedallaMayor devuelve la medalla de más puntos que ya alcanzó la reputación.
func MedallaMayor(reputacion int, catalogo []modelos.Medalla) *modelos.Medalla {
	var mejor *modelos.Medalla
	for i := range catalogo {
		m := &catalogo[i]
		if reputacion >= m.PuntosMinimos && (mejor == nil || m.PuntosMinimos > mejor.PuntosMinimos) {
			mejor = m
		}
	}
	return mejor
}
