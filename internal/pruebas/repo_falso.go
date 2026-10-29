// Package pruebas: repositorio en memoria para probar sin PostgreSQL.
package pruebas

import (
	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"
)

type Falso struct {
	Comentarios map[uint]*modelos.Comentario
	Usuarios    map[uint]*modelos.Usuario
	Negocios    map[uint]*modelos.Negocio
	Cuentas     map[string]*modelos.Cuenta
	Categorias  map[uint]bool
	Puntos      []modelos.Punto
	Catalogo    []modelos.Medalla
	Otorgadas   []modelos.UsuarioMedalla
	proximo     uint
}

// Nuevo deja datos de partida parecidos a la semilla.
func Nuevo(claveHash func(string) string) *Falso {
	f := &Falso{
		Comentarios: map[uint]*modelos.Comentario{}, Usuarios: map[uint]*modelos.Usuario{},
		Negocios: map[uint]*modelos.Negocio{}, Cuentas: map[string]*modelos.Cuenta{},
		Categorias: map[uint]bool{1: true}, proximo: 100,
		Catalogo: []modelos.Medalla{{ID: 1, Nombre: "Novato", PuntosMinimos: 0}, {ID: 2, Nombre: "Ayudante", PuntosMinimos: 25}, {ID: 3, Nombre: "Experto", PuntosMinimos: 50}},
	}
	f.Negocios[1] = &modelos.Negocio{ID: 1, Nombre: "Tienda", EstadoPago: modelos.PagoActiva}
	f.Usuarios[1] = &modelos.Usuario{ID: 1, NegocioID: 1, Nombre: "Jose", Rol: modelos.RolMiembro}
	f.Usuarios[2] = &modelos.Usuario{ID: 2, NegocioID: 1, Nombre: "Luis", Rol: modelos.RolModerador}
	f.Cuentas["jose"] = &modelos.Cuenta{ID: 1, Login: "jose", ClaveHash: claveHash("jose-2026"), UsuarioID: 1, Persona: *f.Usuarios[1]}
	f.Cuentas["luis"] = &modelos.Cuenta{ID: 2, Login: "luis", ClaveHash: claveHash("luis-2026"), UsuarioID: 2, Persona: *f.Usuarios[2]}
	f.Comentarios[1] = &modelos.Comentario{ID: 1, UsuarioID: 1, Texto: "ya destacado", Estado: modelos.EstadoDestacado}
	f.Comentarios[2] = &modelos.Comentario{ID: 2, UsuarioID: 1, Texto: "pendiente", Estado: modelos.EstadoPendiente}
	return f
}

func (f *Falso) CrearComentario(c *modelos.Comentario) error {
	f.proximo++
	c.ID = f.proximo
	cp := *c
	f.Comentarios[c.ID] = &cp
	return nil
}
func (f *Falso) ListarComentarios(estado string, page, limit int) ([]modelos.Comentario, error) {
	var out []modelos.Comentario
	for _, c := range f.Comentarios {
		if estado == "" || c.Estado == estado {
			out = append(out, *c)
		}
	}
	return out, nil
}
func (f *Falso) ObtenerComentario(id uint) (*modelos.Comentario, error) {
	c, ok := f.Comentarios[id]
	if !ok {
		return nil, modelos.ErrNoEncontrado
	}
	cp := *c
	return &cp, nil
}
func (f *Falso) GuardarComentario(c *modelos.Comentario) error {
	cp := *c
	f.Comentarios[c.ID] = &cp
	return nil
}
func (f *Falso) BorrarComentario(id uint) error {
	if _, ok := f.Comentarios[id]; !ok {
		return modelos.ErrNoEncontrado
	}
	delete(f.Comentarios, id)
	return nil
}
func (f *Falso) ObtenerUsuario(id uint) (*modelos.Usuario, error) {
	u, ok := f.Usuarios[id]
	if !ok {
		return nil, modelos.ErrNoEncontrado
	}
	return u, nil
}
func (f *Falso) ObtenerNegocio(id uint) (*modelos.Negocio, error) {
	n, ok := f.Negocios[id]
	if !ok {
		return nil, modelos.ErrNoEncontrado
	}
	return n, nil
}
func (f *Falso) ListarUsuarios() ([]modelos.Usuario, error) {
	var out []modelos.Usuario
	for _, u := range f.Usuarios {
		out = append(out, *u)
	}
	return out, nil
}
func (f *Falso) ExisteCategoria(id uint) (bool, error) { return f.Categorias[id], nil }
func (f *Falso) SumarReputacion(id uint, p int) (int, error) {
	f.Usuarios[id].Reputacion += p
	return f.Usuarios[id].Reputacion, nil
}
func (f *Falso) RegistrarPunto(p *modelos.Punto) error        { f.Puntos = append(f.Puntos, *p); return nil }
func (f *Falso) CatalogoMedallas() ([]modelos.Medalla, error) { return f.Catalogo, nil }
func (f *Falso) OtorgarMedalla(um *modelos.UsuarioMedalla) error {
	for _, o := range f.Otorgadas {
		if o.UsuarioID == um.UsuarioID && o.MedallaID == um.MedallaID {
			return nil
		}
	}
	f.Otorgadas = append(f.Otorgadas, *um)
	return nil
}
func (f *Falso) EnTransaccion(fn func(servicio.Repositorio) error) error { return fn(f) }
func (f *Falso) BuscarCuenta(login string) (*modelos.Cuenta, error) {
	c, ok := f.Cuentas[login]
	if !ok {
		return nil, modelos.ErrNoEncontrado
	}
	return c, nil
}
