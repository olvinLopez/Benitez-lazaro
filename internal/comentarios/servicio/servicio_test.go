package servicio_test

import (
	"errors"
	"testing"

	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"
	"COMUNITARIO/internal/pruebas"
)

func nuevo() (*servicio.Servicio, *pruebas.Falso) {
	f := pruebas.Nuevo(func(s string) string { return s })
	return servicio.Nuevo(f), f
}

func TestTransiciones(t *testing.T) {
	casos := []struct {
		de, a string
		ok    bool
	}{
		{modelos.EstadoPendiente, modelos.EstadoAprobado, true},
		{modelos.EstadoAprobado, modelos.EstadoDestacado, true},
		{modelos.EstadoPendiente, modelos.EstadoDestacado, false},
		{modelos.EstadoDestacado, modelos.EstadoPendiente, false},
		{modelos.EstadoDestacado, modelos.EstadoAprobado, false},
		{modelos.EstadoAprobado, modelos.EstadoPendiente, false},
	}
	for _, c := range casos {
		if got := servicio.TransicionPermitida(c.de, c.a); got != c.ok {
			t.Errorf("%s→%s: esperado %v, salió %v", c.de, c.a, c.ok, got)
		}
	}
}

func TestMedallaMayor(t *testing.T) {
	cat := []modelos.Medalla{{Nombre: "Novato"}, {Nombre: "Ayudante", PuntosMinimos: 25}, {Nombre: "Experto", PuntosMinimos: 50}}
	for rep, esperado := range map[int]string{0: "Novato", 24: "Novato", 25: "Ayudante", 50: "Experto", 99: "Experto"} {
		if m := servicio.MedallaMayor(rep, cat); m == nil || m.Nombre != esperado {
			t.Errorf("rep %d: esperado %s", rep, esperado)
		}
	}
}

func TestTransicionProhibidaDa422(t *testing.T) {
	s, _ := nuevo()
	_, err := s.CambiarEstado(2, 1, modelos.EstadoPendiente) // comentario 1 está destacado
	if !errors.Is(err, servicio.ErrTransicion) {
		t.Fatalf("esperaba ErrTransicion, salió %v", err)
	}
}

func TestDestacarSumaPuntosYOtorgaMedalla(t *testing.T) {
	s, f := nuevo()
	if _, err := s.CambiarEstado(2, 2, modelos.EstadoAprobado); err != nil {
		t.Fatal(err)
	}
	c, err := s.CambiarEstado(2, 2, modelos.EstadoDestacado)
	if err != nil {
		t.Fatal(err)
	}
	if c.ModeradorID == nil || *c.ModeradorID != 2 {
		t.Error("debe guardar quién moderó")
	}
	if f.Usuarios[1].Reputacion != servicio.PuntosDestacado || len(f.Puntos) != 1 {
		t.Errorf("reputación %d, puntos %d", f.Usuarios[1].Reputacion, len(f.Puntos))
	}
	if len(f.Otorgadas) != 1 || f.Otorgadas[0].MedallaID != 1 {
		t.Errorf("debía otorgar Novato: %+v", f.Otorgadas)
	}
}

func TestCrearValidaciones(t *testing.T) {
	s, f := nuevo()
	if _, err := s.Crear(1, "   ", nil); !errors.Is(err, servicio.ErrTextoVacio) {
		t.Errorf("texto vacío: %v", err)
	}
	if _, err := s.Crear(99, "hola", nil); !errors.Is(err, servicio.ErrUsuarioInvalido) {
		t.Errorf("usuario inexistente: %v", err)
	}
	cat := uint(77)
	if _, err := s.Crear(1, "hola", &cat); !errors.Is(err, servicio.ErrCategoria) {
		t.Errorf("categoría inexistente: %v", err)
	}
	f.Negocios[1].EstadoPago = modelos.PagoVencida
	if _, err := s.Crear(1, "hola", nil); !errors.Is(err, servicio.ErrSuscripcion) {
		t.Errorf("negocio vencido: %v", err)
	}
	f.Negocios[1].EstadoPago = modelos.PagoActiva
	c, err := s.Crear(1, "hola", nil)
	if err != nil || c.Estado != modelos.EstadoPendiente {
		t.Errorf("debía nacer pendiente: %v %+v", err, c)
	}
}

func TestListarParametros(t *testing.T) {
	s, _ := nuevo()
	if _, err := s.Listar("", 1, 21); !errors.Is(err, servicio.ErrParametros) {
		t.Error("limit 21 debe fallar")
	}
	if _, err := s.Listar("raro", 1, 10); !errors.Is(err, servicio.ErrEstadoFiltro) {
		t.Error("estado raro debe fallar")
	}
}
