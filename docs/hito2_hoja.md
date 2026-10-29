# Hito 2 · Hoja del hito — AyudaComún

Pareja: Robinson Benitez · Alexander Lazaro — Aplicaciones Web II, paralelo A

## 1 · Usuarios de ejemplo

| Usuario | Rol | Contraseña de ejemplo |
|---------|-----|-----------------------|
| `jose` | miembro | `jose-2026` |
| `luis` | moderador | `luis-2026` |

Son usuarios de prueba, con contraseñas inventadas para la demostración. Nunca una contraseña real.

## 2 · Las seis peticiones

Petición 0 (preparación): `POST /comentarios` con el token de Jose crea un comentario nuevo (queda `pendiente`). Su id se pone en las peticiones 4 y 5.

| # | Petición | Con qué credencial | Debe responder |
|---|----------|--------------------|----------------|
| 1 | `GET /comentarios` | Ninguna | 401 |
| 2 | `POST /login` con `jose` | Ninguna | 200 y el token |
| 3 | `GET /comentarios` | Token de Jose | 200 |
| 4 | `PATCH /comentarios/{id nuevo}/estado` → `aprobado` | Token de Jose (miembro) | 403 |
| 5 | `PATCH /comentarios/{id nuevo}/estado` → `aprobado` | Token de Luis (moderador) | 200 |
| 6 | `PATCH /comentarios/1/estado` → `pendiente`, estando `destacado` | Token de Luis (moderador) | 422 |

## 3 · Matriz ruta × rol

| Ruta | Sin sesión | Miembro | Moderador |
|------|------------|---------|-----------|
| `POST /login` | Sí | Sí | Sí |
| `GET /comentarios` | No (401) | Sí | Sí |
| `POST /comentarios` | No (401) | Sí | Sí |
| `GET /comentarios/{id}` | No (401) | Sí | Sí |
| `PATCH /comentarios/{id}/estado` | No (401) | No (403) | Sí |
| `DELETE /comentarios/{id}` | No (401) | No (403) | Sí |
| `GET /usuarios` | No (401) | Sí | Sí |

## 4 · Recuadro de cumplimiento

☑ Los structs de todas las entidades están en el código y sus tablas existen · ☑ El validador termina sin errores ni avisos · ☑ Las seis peticiones están en la colección, en orden · ☑ Cada una responde lo que dice esta hoja · ☑ `go test ./...` en verde, con una prueba de integración · ☑ El código de la entidad principal está en capas y la estructura está explicada en `docs/decisiones.md` · ☑ Hay una captura de respaldo de cada paso