# Hito 2 · Pasos para ejecutar y probar

1. En `.env` agreguen la línea `JWT_SECRET=` con un secreto inventado (por ejemplo `un-secreto-largo-solo-mio`). No lo suban al repositorio. En `.env.example` ya queda `JWT_SECRET=cambiar_este_secreto`.
2. Borren la carpeta vieja `internal/comentarios` y copien las carpetas nuevas de `internal/`, el `main.go`, `redocly.yaml` y la carpeta `docs/`.
3. Dependencias: no hay nada nuevo (el token y la clave usan solo la librería estándar de Go). Ejecuten `go mod tidy` por si acaso.
4. Arranquen con tablas nuevas: `go run . -reset` (borra y crea las 8 tablas y los datos de ejemplo).
5. Pruebas: `go test ./...` (no necesita PostgreSQL).
6. Validar el contrato: `npx @redocly/cli lint docs/openapi.yaml --format=stylish`.
7. Bruno: abrir la carpeta `docs/hito2_bruno`, elegir el entorno `local`. Antes de cada ensayo ejecutar la petición 0, copiar el id que sale en la consola y pegarlo en las peticiones 4 y 5 (reemplazando `ID_NUEVO`).
8. Capturas en `docs/hito2_capturas/`: una por cada petición (0 a 6), una del validador y una de `go test ./...`.
