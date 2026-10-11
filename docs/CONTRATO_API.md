# Contrato inicial de las API

## Propósito y flujo

Definí este contrato antes de implementar cada servicio para que Go, Node y el frontend compartan los mismos nombres, formas de datos y reglas de error. La API Go con Fiber será el punto de entrada. Recibirá la matriz A, calculará `A = Q × R`, solicitará estadísticas a la API Node con Express y devolverá Q, R y dichas estadísticas en una respuesta.

La palabra “rotación” del enunciado la interpreto en el contexto de la factorización QR; usaré Gonum/LAPACK para calcularla. Node calculará las estadísticas sobre todos los valores de Q y R, conjuntamente, e informará también si cada matriz es diagonal.

Los arreglos JSON representan matrices por filas (row-major). Índices en metadatos, si se añaden en otra fase, serán cero-based.

## Direcciones y rutas

| Servicio     | Puerto local | Endpoint                      | Uso                                               |
| ------------ | -----------: | ----------------------------- | ------------------------------------------------- |
| Go/Fiber     |       `8080` | `GET /health`                 | Estado básico de la instancia Go                  |
| Go/Fiber     |       `8080` | `POST /api/v1/auth/token`     | Intercambiar credenciales de demostración por JWT |
| Go/Fiber     |       `8080` | `POST /api/v1/factorizations` | Factorizar A y orquestar las estadísticas         |
| Node/Express |       `3000` | `GET /health`                 | Estado básico de la instancia Node                |
| Node/Express |       `3000` | `POST /api/v1/statistics`     | Calcular estadísticas de Q y R                    |

En despliegue, cada servicio escuchará el puerto indicado por `PORT` si la plataforma lo proporciona. Go obtendrá la URL base de Node desde `NODE_API_URL`; localmente podrá apuntar al nombre del servicio Compose y en Render a la URL HTTPS pública de Node. `/health` será un chequeo básico del proceso y no filtrará configuración ni secretos. Para que el frontend pueda iniciar ambos servicios dormidos, consultará ambas rutas de salud directamente.

## Entrada de factorización (Go)

`POST /api/v1/factorizations`

Requiere `Authorization: Bearer <JWT>` vigente.

### Obtener token

`POST /api/v1/auth/token` es pública. Recibe `{ "username": "...", "password": "..." }`; devuelve `accessToken`, `tokenType: "Bearer"`, `expiresIn` en segundos y `expiresAt` UTC. Credenciales incorrectas responden `401 UNAUTHORIZED` con mensaje genérico.

```json
{
  "matrix": [
    [1, 0],
    [0, 1]
  ]
}
```

`matrix` debe ser una matriz no vacía, rectangular y de números JSON finitos. No se aceptan filas vacías, filas de longitudes distintas, `null`, cadenas, `NaN` ni infinitos. Una entrada inválida produce HTTP `400`.

### Respuesta exitosa de Go

Para la matriz identidad, Q y R también son la identidad. El endpoint devuelve el resultado QR junto con las estadísticas producidas por Node:

```json
{
  "shape": { "rows": 2, "columns": 2 },
  "q": [
    [1, 0],
    [0, 1]
  ],
  "r": [
    [1, 0],
    [0, 1]
  ],
  "statistics": {
    "minimum": 0,
    "maximum": 1,
    "sum": 4,
    "average": 0.5,
    "count": 8,
    "diagonal": {
      "q": { "applicable": true, "isDiagonal": true },
      "r": { "applicable": true, "isDiagonal": true },
      "anyMatrixDiagonal": true
    },
    "tolerance": 1e-10
  }
}
```

Q será una matriz ortogonal completa `m × m`; R tendrá dimensiones `m × n`, para una entrada A de `m × n`. Esto define el resultado también para matrices anchas (`m < n`). Q y R son resultados de coma flotante; las igualdades matemáticas se entienden dentro de tolerancia numérica.

## Entrada de estadísticas (Node)

`POST /api/v1/statistics`

Esta integración interna requiere el mismo `Authorization: Bearer <JWT>` y `X-Service-Token: <SERVICE_API_KEY>`. Go valida primero el token y lo propaga; Node valida nuevamente firma HS256, issuer, sujeto y expiración. El secreto JWT y la clave de servicio se configuran por separado y nunca van al navegador. `GET /health` y la emisión de token permanecen públicas.

```json
{
  "matrices": {
    "q": [
      [1, 0],
      [0, 1]
    ],
    "r": [
      [1, 0],
      [0, 1]
    ]
  }
}
```

Node recorrerá los elementos de las dos matrices una vez. `minimum`, `maximum`, `sum` y `average` se calculan sobre Q y R conjuntamente; `count` incluye los valores de ambas. La diagonalidad se determina por separado para cada matriz. Una matriz cuadrada es diagonal cuando todos sus elementos fuera de la diagonal cumplen `abs(value) <= tolerance`. Una matriz rectangular tiene `applicable: false` e `isDiagonal: null`. `anyMatrixDiagonal` indica si al menos una comprobación aplicable resultó verdadera.

### Respuesta exitosa de Node

```json
{
  "minimum": 0,
  "maximum": 1,
  "sum": 4,
  "average": 0.5,
  "count": 8,
  "diagonal": {
    "q": { "applicable": true, "isDiagonal": true },
    "r": { "applicable": true, "isDiagonal": true },
    "anyMatrixDiagonal": true
  },
  "tolerance": 1e-10
}
```

La tolerancia inicial es `1e-10` y se configura mediante `DIAGONAL_TOLERANCE` en Node. La API también valida filas, columnas y total de elementos recibidos usando sus variables de entorno.

## Forma común de errores

Ambas API utilizarán este cuerpo JSON para errores conocidos:

```json
{
  "error": {
    "code": "INVALID_MATRIX",
    "message": "La matriz debe ser rectangular y no vacía.",
    "requestId": "identificador-de-solicitud"
  }
}
```

|  HTTP | Código previsto                      | Significado                                                                 |
| ----: | ------------------------------------ | --------------------------------------------------------------------------- |
| `400` | `INVALID_MATRIX` / `INVALID_PAYLOAD` | JSON válido pero estructura o valores no aceptables                         |
| `401` | `UNAUTHORIZED`                       | Credenciales o JWT ausentes, incorrectos o expirados                        |
| `413` | `PAYLOAD_TOO_LARGE`                  | El cuerpo excede el máximo acordado                                         |
| `500` | `INTERNAL_ERROR`                     | Error no esperado; sin detalles internos en el mensaje público              |
| `502` | `STATISTICS_SERVICE_UNAVAILABLE`     | Go no pudo completar la solicitud a Node                                    |
| `504` | `STATISTICS_SERVICE_TIMEOUT`         | Node no respondió dentro del timeout de Go                                  |
| `422` | `NUMERICAL_FAILURE`                  | Los valores finitos de entrada excedieron el rango representable durante QR |
| `422` | `NUMERICAL_FAILURE`                  | La suma estadística excedió el rango numérico finito                        |

Go asignará un `requestId` y lo propagará a Node; Node lo devolverá en errores. Respuestas exitosas podrán incluirlo en la cabecera `X-Request-ID`. La implementación podrá añadir un identificador si no se recibió uno válido.

## Límites iniciales

Para acotar consumo de memoria y CPU desde la primera versión, propongo estos máximos configurables:

- Matriz de entrada: hasta 100 filas, 100 columnas y 10 000 elementos.
- Petición JSON: hasta 1 MiB.
- Entrada de estadísticas: Q y R de hasta 100 filas/columnas cada una; el total de elementos de ambas no puede superar 20 000.
- No se aceptan matrices con valores no finitos. Las respuestas numéricas también deberán ser finitas y serializables en JSON.

El máximo de salida considera que Q es `m × m`; por ello el límite de filas también acota el crecimiento cuadrático de Q. Estos límites son decisiones iniciales para una demo, no requisitos del PDF, y se podrán ajustar en configuración sin cambiar el contrato.

## Autenticación y CORS

Las rutas `/health` permanecen públicas para que el navegador pueda comprobar disponibilidad y despertar las instancias en Render (ya que se desactivan después de 15 minutos de inactividad). Go emite tokens HS256 de 15 minutos tras validar credenciales de demo y protege las operaciones de factorización. Node vuelve a validar el JWT de usuario y exige una credencial de servicio aparte en las llamadas Go→Node. El frontend guarda el token solo en `sessionStorage` y elimina la sesión al expirar. CORS se limita al origen del sitio Angular publicado; la llamada Go→Node servidor a servidor no depende de CORS.

## Variables de entorno previstas

| Variable                         | Servicio  | Valor local inicial                           | Uso                                                                               |
| -------------------------------- | --------- | --------------------------------------------- | --------------------------------------------------------------------------------- |
| `PORT`                           | Go y Node | `8080` / `3000`                               | Puerto HTTP; prevalece el valor de la plataforma                                  |
| `NODE_API_URL`                   | Go        | `http://api-node:3000` en Compose             | URL base de Node                                                                  |
| `NODE_API_TIMEOUT`               | Go        | `90s`                                         | Tiempo máximo para la llamada a Node; debe permitir su arranque en frío en Render |
| `MAX_MATRIX_ROWS`                | Go        | `100`                                         | Máximo configurable de filas de la matriz original                                |
| `MAX_MATRIX_COLUMNS`             | Go        | `100`                                         | Máximo configurable de columnas de la matriz original                             |
| `DIAGONAL_TOLERANCE`             | Node      | `1e-10`                                       | Tolerancia absoluta de diagonalidad                                               |
| `MAX_MATRIX_ROWS`                | Node      | `100`                                         | Máximo de filas de cada matriz recibida                                           |
| `MAX_MATRIX_COLUMNS`             | Node      | `100`                                         | Máximo de columnas de cada matriz recibida                                        |
| `MAX_MATRIX_ELEMENTS`            | Node      | `20000`                                       | Máximo conjunto de elementos de Q y R                                             |
| `FRONTEND_ORIGINS`               | Go y Node | `http://localhost:4200,http://127.0.0.1:4200` | Orígenes web permitidos por CORS, separados por coma                              |
| `JWT_SECRET`                     | Go y Node | Obligatoria, idéntica en ambos servicios      | Clave aleatoria en Base64 estándar que decodifica a al menos 32 bytes para HS256  |
| `JWT_ISSUER`                     | Go y Node | `reto-tecnico`                                | Emisor del JWT, igual en ambas API                                                |
| `JWT_TTL`                        | Go        | `15m`                                         | Vigencia del token (hasta 24 h)                                                   |
| `DEMO_USERNAME`, `DEMO_PASSWORD` | Go        | Obligatoria                                   | Credenciales únicas de acceso a la demo                                           |
| `SERVICE_API_KEY`                | Go y Node | Obligatoria, idéntica en ambos servicios      | Secreto independiente de al menos 32 bytes para Go→Node                           |

Los secretos de JWT y de servicio se configuran fuera del repositorio y nunca se incluyen en archivos versionados.
