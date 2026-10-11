# API Go

Servicio HTTP Fiber para recibir matrices, calcular su factorización QR con Gonum/LAPACK y consultar las estadísticas a la API Node. El contrato entre servicios está documentado en [../docs/CONTRATO_API.md](../docs/CONTRATO_API.md).

## Organización por módulos

```text
main.go                  Composición e inicio del proceso
config/                  Configuración compartida del servicio
routes/                  Ensambla módulos y middleware HTTP
modules/
  factorization/
    contracts/            DTOs propios de factorización y estadísticas
    controllers/           Adaptador HTTP y orquestación del caso de uso
    validators/            Validación del payload y de las matrices
    logic/                 Caso de uso QR encapsulando Gonum/LAPACK
  health/
    contracts/             Respuesta del endpoint de salud
    controllers/            Endpoint de salud
integrations/nodeapi/      Adaptador HTTP hacia la API Node
  contracts/               DTOs del protocolo de estadísticas con Node
shared/
  contracts/               Sobre de error común
  http/                     Request IDs y manejo de errores Fiber
```

Agrupo por capacidad de negocio para que cada módulo sea dueño de sus contratos, entrada HTTP y validaciones. `health` no contiene validadores ni lógica vacíos, porque su endpoint no recibe datos ni ejecuta un caso de uso. El cliente Node y los DTOs exactos de su protocolo viven en `integrations/nodeapi`. El módulo de factorización mantiene su propio DTO de respuesta pública; el controlador traduce entre ambos contratos en el límite de integración para evitar que el modelo HTTP de Node se filtre a la API Go. La validación de matrices es exclusiva de factorización. La dependencia Gonum/LAPACK queda encapsulada en el servicio del módulo; el controlador no conoce sus detalles de almacenamiento ni de extracción de factores. `routes` solo conecta dependencias y registra endpoints.

Uso la rutina QR de LAPACK incluida en Gonum para admitir matrices altas, cuadradas y anchas (`filas < columnas`). En todos los casos se devuelve `Q` cuadrada de tamaño `m × m` y `R` de tamaño `m × n`, de modo que `A = Q·R`. En matrices anchas extraigo los reflectores almacenados por QR y Gonum genera la `Q` completa. Conservo una comprobación de resultados finitos para traducir desbordamientos numéricos a un error HTTP controlado.

## Ejecución local

Requiere Go 1.26 o posterior y la API Node ejecutándose en `http://localhost:3000` (`cd ../api-node; npm ci; npm start`).
Al ejecutar la API desde `api-go`, Go carga automáticamente `api-go/.env`; las variables ya definidas en el sistema tienen prioridad. El archivo `.env` no debe versionarse.

```powershell
go run .
```

## Pruebas

Ejecuta las pruebas unitarias de QR y validación, además de la integración HTTP Fiber→Node simulado:

```powershell
go test ./...
```

| Variable              | Predeterminado                                | Descripción                                                                                                              |
| --------------------- | --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `PORT`                | `8080`                                        | Puerto HTTP                                                                                                              |
| `NODE_API_URL`        | `http://localhost:3000`                       | URL base de Express                                                                                                      |
| `NODE_API_TIMEOUT`    | `90s`                                         | Tiempo límite de la integración Node                                                                                     |
| `MAX_MATRIX_ROWS`     | `100`                                         | Máximo de filas                                                                                                          |
| `MAX_MATRIX_COLUMNS`  | `100`                                         | Máximo de columnas                                                                                                       |
| `MAX_MATRIX_ELEMENTS` | `10000`                                       | Máximo total de elementos por matriz                                                                                     |
| `FRONTEND_ORIGINS`    | `http://localhost:4200,http://127.0.0.1:4200` | Orígenes CORS autorizados, separados por coma                                                                            |
| `JWT_SECRET`          | Obligatoria                                   | Clave aleatoria codificada en Base64 estándar; al decodificarla debe tener al menos 32 bytes. Compartida entre Go y Node |
| `JWT_ISSUER`          | `reto-tecnico`                                | Emisor que validan ambas API                                                                                             |
| `JWT_TTL`             | `15m`                                         | Vigencia del token, máximo 24 horas                                                                                      |
| `DEMO_USERNAME`       | Obligatoria                                   | Usuario único de demostración                                                                                            |
| `DEMO_PASSWORD`       | Obligatoria                                   | Contraseña de demostración                                                                                               |
| `SERVICE_API_KEY`     | Obligatoria                                   | Clave aleatoria independiente de al menos 32 bytes, compartida entre Go y Node                                           |

Las peticiones se limitan a 1 MiB. El tamaño de matriz está acotado independientemente por filas, columnas y cantidad total de elementos. Configuro `FRONTEND_ORIGINS` con los orígenes CORS permitidos y lo ajusto al dominio publicado del frontend.

`POST /api/v1/auth/token` intercambia las credenciales de demostración por un JWT HS256. Envío ese token como Bearer a la ruta de factorización; Go lo valida y lo reenvía a Node junto con `X-Service-Token`.
