// Package repositorio es la ÚNICA capa que habla con PostgreSQL (GORM).
// Si mañana cambia la base de datos, solo se toca este archivo.
package repositorio

import (
	"errors"

	"COMUNITARIO/internal/comentarios/servicio"
	"COMUNITARIO/internal/modelos"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Gorm struct{ DB *gorm.DB }

func Nuevo(db *gorm.DB) *Gorm { return &Gorm{DB: db} }

func traducir(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return modelos.ErrNoEncontrado
	}
	return err
}

func (g *Gorm) CrearComentario(c *modelos.Comentario) error { return g.DB.Create(c).Error }

func (g *Gorm) ListarComentarios(estado string, page, limit int) ([]modelos.Comentario, error) {
	q := g.DB.Order("id desc").Limit(limit).Offset((page - 1) * limit)
	if estado != "" {
		q = q.Where("estado = ?", estado) // parámetro seguro (evita inyección SQL)
	}
	var lista []modelos.Comentario
	err := q.Find(&lista).Error
	return lista, err
}

func (g *Gorm) ObtenerComentario(id uint) (*modelos.Comentario, error) {
	var c modelos.Comentario
	if err := g.DB.First(&c, id).Error; err != nil {
		return nil, traducir(err)
	}
	return &c, nil
}

func (g *Gorm) GuardarComentario(c *modelos.Comentario) error { return g.DB.Save(c).Error }

func (g *Gorm) BorrarComentario(id uint) error {
	res := g.DB.Delete(&modelos.Comentario{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return modelos.ErrNoEncontrado
	}
	return nil
}

func (g *Gorm) ObtenerUsuario(id uint) (*modelos.Usuario, error) {
	var u modelos.Usuario
	if err := g.DB.First(&u, id).Error; err != nil {
		return nil, traducir(err)
	}
	return &u, nil
}

func (g *Gorm) ObtenerNegocio(id uint) (*modelos.Negocio, error) {
	var n modelos.Negocio
	if err := g.DB.First(&n, id).Error; err != nil {
		return nil, traducir(err)
	}
	return &n, nil
}

// ListarUsuarios usa Preload: pocas consultas, sin el problema N+1.
func (g *Gorm) ListarUsuarios() ([]modelos.Usuario, error) {
	var us []modelos.Usuario
	err := g.DB.Preload("Comentarios").Preload("Medallas.Medalla").Find(&us).Error
	return us, err
}

func (g *Gorm) ExisteCategoria(id uint) (bool, error) {
	var n int64
	err := g.DB.Model(&modelos.Categoria{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

func (g *Gorm) SumarReputacion(usuarioID uint, puntos int) (int, error) {
	err := g.DB.Model(&modelos.Usuario{}).Where("id = ?", usuarioID).
		UpdateColumn("reputacion", gorm.Expr("reputacion + ?", puntos)).Error
	if err != nil {
		return 0, err
	}
	var u modelos.Usuario
	if err := g.DB.Select("reputacion").First(&u, usuarioID).Error; err != nil {
		return 0, traducir(err)
	}
	return u.Reputacion, nil
}

func (g *Gorm) RegistrarPunto(p *modelos.Punto) error { return g.DB.Create(p).Error }

func (g *Gorm) CatalogoMedallas() ([]modelos.Medalla, error) {
	var ms []modelos.Medalla
	err := g.DB.Order("puntos_minimos").Find(&ms).Error
	return ms, err
}

// OtorgarMedalla no repite: si ya la tiene, no hace nada.
func (g *Gorm) OtorgarMedalla(um *modelos.UsuarioMedalla) error {
	return g.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(um).Error
}

// EnTransaccion: todo lo que haga fn se guarda junto o no se guarda nada.
func (g *Gorm) EnTransaccion(fn func(servicio.Repositorio) error) error {
	return g.DB.Transaction(func(tx *gorm.DB) error { return fn(&Gorm{DB: tx}) })
}

// BuscarCuenta sirve al inicio de sesión (auth.CuentaRepo).
func (g *Gorm) BuscarCuenta(login string) (*modelos.Cuenta, error) {
	var c modelos.Cuenta
	if err := g.DB.Preload("Persona").Where("login = ?", login).First(&c).Error; err != nil {
		return nil, traducir(err)
	}
	return &c, nil
}
