# API Node.js

Servicio Express que calcula estadísticas agregadas y diagonalidad para Q y R recibidas desde Go. El contrato HTTP está documentado en [../docs/CONTRATO_API.md](../docs/CONTRATO_API.md).

## Arquitectura modular

```text
src/
  main.js, app.js, config.js       Arranque, composición y configuración
  modules/
    health/                        Endpoint de salud
      contracts/ controllers/ routes/
    statistics/                    Validación y caso de uso estadístico
      contracts/ controllers/ logic/ routes/ validators/
  shared/http/                     Request ID y errores HTTP comunes
```

Mantengo cada funcionalidad dentro de su módulo. `statistics` valida el payload, calcula agregados en un recorrido y presenta su contrato de salida; `health` solo responde con el estado básico. `app.js` registra módulos y middleware, mientras que `main.js` carga la configuración y levanta el servidor.

## Ejecución

Requiere Node.js 22.12 o posterior.
Node carga `api-node/.env` relativo a `src/main.js`, incluso si se inicia desde la raíz con `npm --prefix api-node start`. Las variables ya definidas en el sistema tienen prioridad. El archivo `.env` no debe versionarse.

```powershell
npm ci
npm start
```

## Pruebas

`npm test` ejecuta con el runner nativo de Node las pruebas unitarias de estadísticas/validación y la integración Express con JWT, clave de servicio y matriz inválida.

| Variable | Predeterminado | Descripción |
|---|---:|---|
| `PORT` | `3000` | Puerto HTTP |
| `DIAGONAL_TOLERANCE` | `1e-10` | Tolerancia absoluta para diagonalidad |
| `MAX_MATRIX_ROWS` | `100` | Máximo de filas por matriz |
| `MAX_MATRIX_COLUMNS` | `100` | Máximo de columnas por matriz |
| `MAX_MATRIX_ELEMENTS` | `20000` | Máximo conjunto de elementos Q y R |
| `FRONTEND_ORIGINS` | `http://localhost:4200,http://127.0.0.1:4200` | Orígenes CORS autorizados, separados por coma |
| `JWT_SECRET` | Obligatoria | El mismo secreto en Base64 estándar que usa Go; debe decodificar a al menos 32 bytes |
| `JWT_ISSUER` | `reto-tecnico` | Emisor esperado en el token; debe coincidir con Go |
| `SERVICE_API_KEY` | Obligatoria | La misma clave interna aleatoria de al menos 32 bytes que usa Go |

`GET /health` informa disponibilidad. `POST /api/v1/statistics` recibe `{ "matrices": { "q": [[...]], "r": [[...]] } }`. El cuerpo se limita a 1 MiB. El mínimo, máximo, suma y promedio se calculan sobre ambas matrices; la diagonalidad se evalúa por matriz y solo aplica a formas cuadradas.

La ruta de estadísticas requiere un JWT HS256 con issuer, sujeto y expiración válidos, además de `X-Service-Token`. La clave de servicio es distinta del secreto JWT. `GET /health` permanece pública para los chequeos de disponibilidad.
