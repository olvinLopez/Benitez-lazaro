// Package servicio aplica las reglas del negocio sobre los comentarios.
// Habla con la base de datos solo a través de la interfaz Repositorio.
package servicio

import (
	"errors"
	"strings"
	"time"

	"COMUNITARIO/internal/modelos"
)

// Errores del negocio. La capa HTTP decide qué código HTTP les corresponde.
var (
	ErrParametros      = errors.New("page y limit deben ser números válidos (limit máximo 20)")
	ErrEstadoFiltro    = errors.New("estado no permitido")
	ErrTextoVacio      = errors.New("el texto es obligatorio")
	ErrTextoLargo      = errors.New("el texto no puede pasar de 1000 caracteres")
	ErrUsuarioInvalido = errors.New("el usuario no existe")
	ErrCategoria       = errors.New("la categoría no existe")
	ErrSuscripcion     = errors.New("la suscripción del negocio no está activa")
	ErrEstadoInvalido  = errors.New("estado no permitido")
	ErrTransicion      = errors.New("transición de estado no permitida")
)

// Repositorio es lo que el servicio necesita de la base de datos.
type Repositorio interface {
	CrearComentario(c *modelos.Comentario) error
	ListarComentarios(estado string, page, limit int) ([]modelos.Comentario, error)
	ObtenerComentario(id uint) (*modelos.Comentario, error)
	GuardarComentario(c *modelos.Comentario) error
	BorrarComentario(id uint) error
	ObtenerUsuario(id uint) (*modelos.Usuario, error)
	ObtenerNegocio(id uint) (*modelos.Negocio, error)
	ListarUsuarios() ([]modelos.Usuario, error)
	ExisteCategoria(id uint) (bool, error)
	SumarReputacion(usuarioID uint, puntos int) (int, error)
	RegistrarPunto(p *modelos.Punto) error
	CatalogoMedallas() ([]modelos.Medalla, error)
	OtorgarMedalla(um *modelos.UsuarioMedalla) error
	EnTransaccion(fn func(Repositorio) error) error
}

type Servicio struct {
	Repo  Repositorio
	Ahora func() time.Time
}

func Nuevo(r Repositorio) *Servicio { return &Servicio{Repo: r, Ahora: time.Now} }

// Crear: el autor sale del token (no del cuerpo). Todo comentario nace pendiente.
func (s *Servicio) Crear(usuarioID uint, texto string, categoriaID *uint) (*modelos.Comentario, error) {
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return nil, ErrTextoVacio
	}
	if len([]rune(texto)) > 1000 {
		return nil, ErrTextoLargo
	}
	usr, err := s.Repo.ObtenerUsuario(usuarioID)
	if err != nil {
		return nil, ErrUsuarioInvalido
	}
	neg, err := s.Repo.ObtenerNegocio(usr.NegocioID)
	if err != nil || neg.EstadoPago != modelos.PagoActiva {
		return nil, ErrSuscripcion
	}
	if categoriaID != nil {
		ok, err := s.Repo.ExisteCategoria(*categoriaID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrCategoria
		}
	}
	ahora := s.Ahora()
	c := &modelos.Comentario{
		UsuarioID: usuarioID, CategoriaID: categoriaID, Texto: texto,
		Estado: modelos.EstadoPendiente, Creado: ahora, Actualizado: ahora,
	}
	if err := s.Repo.CrearComentario(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Listar con filtro seguro y paginado.
func (s *Servicio) Listar(estado string, page, limit int) ([]modelos.Comentario, error) {
	if estado != "" && !EstadoValido(estado) {
		return nil, ErrEstadoFiltro
	}
	if page < 1 || limit < 1 || limit > LimiteMaximo {
		return nil, ErrParametros
	}
	return s.Repo.ListarComentarios(estado, page, limit)
}

func (s *Servicio) Obtener(id uint) (*modelos.Comentario, error) {
	return s.Repo.ObtenerComentario(id)
}

func (s *Servicio) ListarUsuarios() ([]modelos.Usuario, error) {
	return s.Repo.ListarUsuarios()
}

func (s *Servicio) Borrar(id uint) error { return s.Repo.BorrarComentario(id) }

// CambiarEstado aplica la máquina de estados. Quién puede llamarla (moderador)
// lo decide el middleware de roles, no este archivo.
// Al destacar: suma puntos, actualiza reputación y otorga la medalla que corresponda,
// todo en UNA transacción.
func (s *Servicio) CambiarEstado(moderadorID, id uint, nuevo string) (*modelos.Comentario, error) {
	if !EstadoValido(nuevo) {
		return nil, ErrEstadoInvalido
	}
	c, err := s.Repo.ObtenerComentario(id)
	if err != nil {
		return nil, err
	}
	if !TransicionPermitida(c.Estado, nuevo) {
		return nil, ErrTransicion
	}

	err = s.Repo.EnTransaccion(func(tx Repositorio) error {
		c.Estado = nuevo
		c.ModeradorID = &moderadorID
		c.Actualizado = s.Ahora()
		if err := tx.GuardarComentario(c); err != nil {
			return err
		}
		if nuevo != modelos.EstadoDestacado {
			return nil
		}
		cid := c.ID
		if err := tx.RegistrarPunto(&modelos.Punto{
			UsuarioID: c.UsuarioID, ComentarioID: &cid, Cantidad: PuntosDestacado,
			Motivo: "comentario destacado", Creado: s.Ahora(),
		}); err != nil {
			return err
		}
		rep, err := tx.SumarReputacion(c.UsuarioID, PuntosDestacado)
		if err != nil {
			return err
		}
		catalogo, err := tx.CatalogoMedallas()
		if err != nil {
			return err
		}
		if m := MedallaMayor(rep, catalogo); m != nil {
			return tx.OtorgarMedalla(&modelos.UsuarioMedalla{
				UsuarioID: c.UsuarioID, MedallaID: m.ID, ComentarioID: &cid, Otorgada: s.Ahora(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
