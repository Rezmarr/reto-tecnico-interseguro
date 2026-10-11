# Reto técnico

Solución de factorización QR con Go/Fiber, Node.js/Express y Angular 21. El contrato está en [docs/CONTRATO_API.md](docs/CONTRATO_API.md). Las API están organizadas por módulos; Go usa Gonum/LAPACK para QR. JWT, pruebas automatizadas y Docker Compose están implementados.

## Levantar todo con Docker

Requisitos: Docker Desktop o Docker Engine con Compose.

1. Se debe definir el `.env` y cambiar los valores de `JWT_SECRET`, `SERVICE_API_KEY`, `DEMO_USERNAME` y `DEMO_PASSWORD`. Genera secretos independientes y aleatorios; `JWT_SECRET` debe ser Base64 estándar de al menos 32 bytes decodificados.
2. Ejecuta `docker compose up --build` desde la raíz.
3. Abre [http://localhost:8088](http://localhost:8088) e inicia sesión con `DEMO_USERNAME` y `DEMO_PASSWORD` del `.env`.

Compose construye las tres imágenes, espera a que Node y Go pasen sus health checks, y publica el frontend. Nginx enruta `/go` hacia Go y `/node` hacia Node por la red privada de Compose. Go también está disponible directamente en `http://localhost:8080`; Node no publica un puerto al host. Detén los servicios con `docker compose down`.

El `.env` raíz solo configura el arranque local de Compose y está excluido de Git y de los contextos de build.

## Ejecutar las pruebas

- Go: `cd api-go; go test ./...`
- Node: `cd api-node; npm test`
- Angular: `cd web; npm ci; npm test`
