// Package semilla crea los datos iniciales (solo si la base está vacía).
package semilla

import (
	"time"

	"COMUNITARIO/internal/auth"
	"COMUNITARIO/internal/modelos"

	"gorm.io/gorm"
)

// Usuarios de ejemplo (contraseñas inventadas, solo para la demostración):
//
//	jose / jose-2026  → miembro
//	luis / luis-2026  → moderador
func Sembrar(db *gorm.DB) error {
	var n int64
	db.Model(&modelos.Negocio{}).Count(&n)
	if n > 0 {
		return nil
	}
	ahora := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		neg := modelos.Negocio{Nombre: "Tienda Manta Tech", RUC: "1391234567001", EstadoPago: modelos.PagoActiva, Creado: ahora}
		if err := tx.Create(&neg).Error; err != nil {
			return err
		}
		jose := modelos.Usuario{NegocioID: neg.ID, Nombre: "Jose Fuentes", Correo: "jose@mantatech.ec", Rol: modelos.RolMiembro, Reputacion: 10, Creado: ahora}
		luis := modelos.Usuario{NegocioID: neg.ID, Nombre: "Luis Anchundia", Rol: modelos.RolModerador, Correo: "luis@mantatech.ec", Reputacion: 0, Creado: ahora}
		if err := tx.Create(&[]*modelos.Usuario{&jose, &luis}).Error; err != nil {
			return err
		}
		for _, c := range []struct {
			login, clave string
			usuario      uint
		}{{"jose", "jose-2026", jose.ID}, {"luis", "luis-2026", luis.ID}} {
			h, err := auth.HashClave(c.clave)
			if err != nil {
				return err
			}
			if err := tx.Create(&modelos.Cuenta{Login: c.login, ClaveHash: h, UsuarioID: c.usuario}).Error; err != nil {
				return err
			}
		}
		cat := []modelos.Categoria{
			{Nombre: "Dudas técnicas", Descripcion: "Problemas con equipos y programas"},
			{Nombre: "Pagos y facturas", Descripcion: "Transferencias, RUC y facturación"},
			{Nombre: "Uso de la app", Descripcion: "Cómo usar la plataforma"},
		}
		if err := tx.Create(&cat).Error; err != nil {
			return err
		}
		med := []modelos.Medalla{
			{Nombre: "Novato", PuntosMinimos: 0, Descripcion: "Primer paso en la comunidad"},
			{Nombre: "Ayudante", PuntosMinimos: 25, Descripcion: "Ayuda con frecuencia"},
			{Nombre: "Experto", PuntosMinimos: 50, Descripcion: "Referente de la comunidad"},
		}
		if err := tx.Create(&med).Error; err != nil {
			return err
		}
		// Comentario 1: ya destacado (sirve para la petición 6: transición prohibida).
		c1 := modelos.Comentario{UsuarioID: jose.ID, CategoriaID: &cat[0].ID, Texto: "Para reiniciar el router espera 30 segundos antes de encenderlo.", Estado: modelos.EstadoDestacado, ModeradorID: &luis.ID, Creado: ahora, Actualizado: ahora}
		c2 := modelos.Comentario{UsuarioID: jose.ID, CategoriaID: &cat[1].ID, Texto: "La transferencia se verifica en 24 horas hábiles.", Estado: modelos.EstadoAprobado, ModeradorID: &luis.ID, Creado: ahora, Actualizado: ahora}
		c3 := modelos.Comentario{UsuarioID: luis.ID, CategoriaID: &cat[2].ID, Texto: "Puedes filtrar los comentarios por estado.", Estado: modelos.EstadoPendiente, Creado: ahora, Actualizado: ahora}
		if err := tx.Create(&[]*modelos.Comentario{&c1, &c2, &c3}).Error; err != nil {
			return err
		}
		if err := tx.Create(&modelos.Punto{UsuarioID: jose.ID, ComentarioID: &c1.ID, Cantidad: 10, Motivo: "comentario destacado", Creado: ahora}).Error; err != nil {
			return err
		}
		return tx.Create(&modelos.UsuarioMedalla{UsuarioID: jose.ID, MedallaID: med[0].ID, ComentarioID: &c1.ID, Otorgada: ahora}).Error
	})
}
