# Plan de implementación — Reto técnico IS

## Estado y alcance

Este archivo contiene el plan y el registro de decisiones de implementación. Ya implementé las fases 1 a 3, el frontend Angular 21, JWT, pruebas automatizadas y Docker Compose. El despliegue en Render queda como trabajo posterior.

Voy a entregar dos API REST —Go/Fiber y Node.js/Express—, una interfaz Angular, protección JWT, pruebas unitarias/de integración y contenedores Docker para desplegar una demo en Render. Implementé primero las dos API y el frontend, siguiendo el orden por etapas que acordamos.

## Lectura del reto y decisión sobre la discrepancia

El PDF describe en una sección que Go debe rotar la matriz y pasarla a Node, pero en la funcionalidad requerida pide explícitamente que Go calcule la factorización QR. También utiliza la palabra “rotación” para la transformación. Tomaré QR como resultado funcional principal. Usaré la biblioteca Gonum/LAPACK para la descomposición numérica y no haré una segunda rotación geométrica independiente de la matriz, porque el reto no define su dirección, eje ni resultado esperado.

La API Go devolverá las matrices Q y R. Node calculará estadísticas sobre el conjunto de valores de ambas matrices y comprobará por separado si Q y R son diagonales. La comprobación será literalmente de diagonalidad, no de triangularidad: R suele ser triangular superior, pero eso no la convierte en diagonal.

## Arquitectura y flujo

```text
Navegador
   │ HTTPS + Bearer JWT
   ▼
Frontend estático (Angular 21)
   │ POST /api/v1/factorizations
   ▼
API Go (Fiber) ──HTTP interno/autenticado──▶ API Node (Express)
   │                                         │
   └──────── respuesta Q, R y estadísticas ◀┘
```

En producción cada servicio tendrá su URL HTTPS y secretos propios inyectados como variables de entorno. Go será el punto de entrada: validará la matriz, calculará QR y solicitará a Node las estadísticas; responderá al cliente con el resultado combinado. Node también expondrá su endpoint de estadísticas para documentar y verificar su responsabilidad de forma independiente.

## Contratos HTTP propuestos

### API Go

- `GET /health`: estado del proceso y disponibilidad básica, sin revelar secretos.
- `POST /api/v1/factorizations`: recibe `{ "matrix": [[1, 2], [3, 4], [5, 6]] }`.
- Éxito: `{ "q": [[...]], "r": [[...]], "statistics": { ... } }`.
- Errores con JSON uniforme y códigos HTTP adecuados: `400` entrada inválida, `401` token inválido/ausente, `502` dependencia Node no disponible, `500` error interno inesperado.

### API Node

- `GET /health`: estado básico.
- `POST /api/v1/statistics`: recibe `{ "matrices": { "q": [[...]], "r": [[...]] } }` y devuelve mínimo, máximo, suma, promedio, cantidad de valores, tolerancia usada y banderas de diagonalidad por matriz.

El contrato utilizará `float64`/número JSON. El promedio se definirá sobre todos los elementos de Q y R, contando ambos conjuntos completos. La salida identificará qué matriz produjo el mínimo y el máximo si el formato sigue claro, para que el resultado sea útil y comprobable.

## Fases de implementación

### Fase 1 — Base del repositorio y decisiones compartidas

1. Crear carpetas independientes para `api-go`, `api-node` y `web`, además de `docs` y archivos de desarrollo raíz.
2. Definir nombres de rutas, estructuras JSON, convención de errores, puertos y variables de entorno antes de implementar la comunicación.
3. Añadir ejemplos de solicitudes/respuestas y límites iniciales para tamaño de matriz en la documentación.

**Justificación:** separar servicios evita mezclar toolchains, mantiene despliegues independientes y hace que la estructura sea fácil de explicar durante una entrevista. Fijar primero el contrato reduce retrabajo entre Go, Node y el frontend.

### Fase 2 — API Go y factorización QR

1. Construir el servidor con Fiber y el endpoint de factorización.
2. Validar matriz no vacía, filas rectangulares, elementos finitos y máximo de filas/columnas o elementos configurable.
3. Calcular QR con Gonum/LAPACK sobre `float64`. Para una matriz `m × n`, producir Q de `m × m` y R de `m × n`; soportar matrices anchas (`m < n`) usando la rutina LAPACK rectangular y generar Q completa desde sus reflectores.
4. Evitar división por cero al procesar pares de valores nulos y normalizar residuos numéricos pequeños donde corresponda, documentando la tolerancia.
5. Mandar Q y R a Node con timeout, propagación controlada de cancelación y número limitado de intentos (por defecto, sin reintento automático para no duplicar carga). Mapear el fallo remoto a `502`.

**Justificación:** Gonum/LAPACK aporta una implementación numérica consolidada, reduce el algoritmo QR propio que habría que mantener y permite factorizar matrices rectangulares. Los límites previenen peticiones que consuman memoria o CPU sin control. El timeout evita que la API Go quede bloqueada indefinidamente si Node no responde.

### Fase 3 — API Node y estadísticas (implementada)

1. Crear el servicio Express con validación del esquema de matrices recibidas.
2. Recorrer Q y R una vez para acumular mínimo, máximo, suma y conteo; calcular promedio al final. Complejidad lineal en el número total de elementos y memoria auxiliar constante.
3. Comprobar diagonalidad únicamente en matrices cuadradas, considerando no diagonales las rectangulares. Comparar elementos fuera de la diagonal con `abs(x) > tolerance`.
4. Responder con resultados JSON estables y errores claros.

**Justificación:** el recorrido único es simple y eficiente. La tolerancia de coma flotante permite que el criterio de diagonalidad sea significativo tras operaciones numéricas. Separar las banderas para Q y R evita ambigüedad en “si alguna matriz es diagonal”.

### Fase 4 — Comunicación, configuración y contenedores

1. Configuré Go con URL de Node, timeout, puertos y límites mediante entorno; Node recibe puerto, secreto JWT y tolerancias mediante entorno.
2. Añadí Dockerfiles para los tres componentes y `.dockerignore`: Go compila en una etapa separada, Node instala dependencias de producción y se ejecuta sin privilegios, y Angular se compila para servirse con Nginx.
3. Añadí `compose.yaml` con la red privada, health checks, orden de arranque y enrutamiento Nginx (`/go` y `/node`). Go llama a Node por nombre de servicio; Node no publica un puerto al host.
4. Añadí `.env.example` sin secretos reales, excluí `.env` de Git e indiqué cómo generar claves locales.

**Justificación:** multi-stage reduce el tamaño de Go y del frontend al excluir toolchains del runtime; Node instala únicamente dependencias de producción. Compose reproduce la topología de servicios y usa nombres DNS internos. El proxy de Nginx evita configuración CORS en navegador para el uso local y la configuración por entorno mantiene secretos fuera de la imagen.

### Fase 5 — Documentación del sistema

Crear `docs/SISTEMA.md` en **primera persona**, con voz de autoría personal (“Diseñé…”, “Elegí…”, “Validé…”). Documentaré arquitectura, requisitos, estructura del repositorio, contratos, algoritmo QR y complejidad, validaciones, errores, configuración, Docker, despliegue, seguridad, frontend, pruebas, limitaciones y decisiones con sus alternativas y razones. El `README.md` quedará como guía breve de arranque y enlazará a ese documento.

**Justificación:** separar una guía corta de una explicación técnica completa facilita tanto ejecutar la solución como defenderla. La primera persona atiende el estilo solicitado para la documentación del sistema.

### Fase 6 — Despliegue gratuito para demostración

Mi opción inicial será **Render**: alojaré Go y Node como dos servicios web Docker gratuitos, y el frontend como sitio estático gratuito. Es una ruta directa para una demo sin administrar máquinas virtuales y admite servicios Docker. Configuraré URLs públicas por variables de entorno y limitaré CORS al origen publicado del frontend. Como el plan gratuito no ofrece tráfico privado entre servicios, Go llamará la URL HTTPS de Node con autenticación de servicio.

Dejaré documentada la limitación: Render suspende servicios web gratuitos tras 15 minutos sin tráfico, la reactivación puede tardar alrededor de un minuto y el sistema de archivos es efímero. Es aceptable para una evaluación/demo, no para producción. Antes del despliegue revisaré límites vigentes, y no afirmaré costo cero garantizado: exceder cuotas o cambiar condiciones puede generar cargos o suspensión. No se guardarán datos, por lo que no se necesita almacenamiento persistente.

**Justificación:** priorizo facilidad de despliegue y costo nulo dentro de las condiciones publicadas para este reto. Cloud Run es alternativa si se prefiere cobro por uso y escala a cero, aunque requiere configurar Google Cloud y puede generar cargos fuera del nivel gratuito; por eso no lo tomaré como ruta inicial.

### Fase 7 — Mejoras finales: frontend, JWT y pruebas

Implementé estos componentes tras estabilizar el contrato y las dos API:

1. **Frontend:** implementé la aplicación Angular 21 con cuadrícula de matriz, validación inmediata, visualización de Q, R y estadísticas, configuración de URLs y estados de arranque en frío. Separé autenticación y cálculo en funcionalidades propias. El sitio será un Render Static Site, así que no se suspende como los servicios backend.
2. **Estado de arranque en frío:** al iniciar un cálculo, mostrar inmediatamente progreso visible y consultar los endpoints `/health` de Go y Node en paralelo para despertarlos. La interfaz identificará cada paso (“Conectando con los servicios”, “Iniciando el servicio de cálculo”, “Iniciando el servicio de estadísticas”, “Calculando”) y al completarse las comprobaciones enviará el cálculo. Si una llamada tarda, mostraré un aviso honesto: “El servicio gratuito puede estar iniciando; esto puede tardar alrededor de un minuto”. No afirmaré que Render está dormido ni que el servicio está cargando cuando el cliente solo observa latencia; la interfaz conoce el estado de la petición, no el estado interno de Render. Incluiré opción de cancelar el proceso desde el navegador, manejo de timeout/errores y acción de reintentar.
3. **JWT (implementado):** Go emite tokens HS256 de 15 minutos tras validar credenciales de demostración configuradas por entorno. Go y Node verifican algoritmo, firma, expiración e issuer; las operaciones requieren Bearer y la integración Go→Node además requiere `X-Service-Token` con clave independiente. Angular conserva el token en `sessionStorage` y limpia la sesión al cerrar o expirar.
4. **Pruebas Go (implementadas):** unitarias de factorización y validación; verifican dimensiones, ortogonalidad `QᵀQ ≈ I` y reconstrucción `QR ≈ A` en matrices cuadradas y rectangulares, además de casos de una fila/columna e inválidos. La integración Fiber emite/verifica JWT y llama a un Node simulado que comprueba cabeceras y formas de datos.
5. **Pruebas Node (implementadas):** unitarias de estadísticas, diagonalidad y entradas inválidas; integración Express comprueba health público, JWT ausente/incorrecto/expirado, clave de servicio, respuesta válida y matriz inválida.
6. **Pruebas frontend (implementadas):** Angular/Vitest comprueba emisión del token, navegación entre autenticación y cálculo, cierre de sesión, redimensionado y validación/envío de la cuadrícula. Ejecuto `go test ./...`, `npm test` en `api-node` y `npm test` en `web`; no dependen de servicios cloud.

**Justificación:** proteger con JWT las dos API y distinguir token de usuario de credencial de servicio evita confiar en una llamada interna por el mero hecho de ser HTTP. El sondeo concurrente de `/health` permite mostrar qué backend está tardando y empieza a despertarlos antes del cálculo; como el estado exacto de Render no se expone al navegador, los mensajes describirán el progreso observable sin atribuir con certeza la demora al cold start. Angular 21 cumple la versión solicitada. Las propiedades matemáticas de QR verifican resultados numéricos de manera más sólida que comparar únicamente una salida fija.

## Decisiones técnicas resumidas

| Decisión | Motivo | Consecuencia/limitación |
|---|---|---|
| Gonum/LAPACK para QR | Usa una implementación numérica mantenida y admite matrices rectangulares | Añade una dependencia; mantengo validación de entradas y comprobación de resultados finitos |
| Q completa `m×m`, R `m×n` | Contrato único válido para matrices altas y anchas | Q puede ser grande; limitar dimensiones |
| Estadísticas conjuntas de Q y R | Operación adicional aplicada a todas las matrices devueltas | La cantidad de valores incluye ambas matrices y se documentará explícitamente |
| Recorrido lineal único en Node | O(N) tiempo, O(1) memoria auxiliar | Promedio debe evitar entradas vacías (rechazadas en validación) |
| HTTP/JSON entre servicios | Interoperable, fácil de probar y desplegar | Necesita timeout y manejo explícito de fallos |
| Docker multi-stage + Compose | Entorno reproducible y despliegue uniforme | Requiere Docker para el arranque local completo |
| Angular 21, al final | Versión solicitada; permite separar componentes, servicios HTTP y estados de carga | El sitio estático se mantiene disponible, pero las API pueden estar despertando |
| Health checks visibles, al final | Despertar Go y Node y explicar demoras de arranque con progreso útil | No se puede distinguir con certeza cold start de red lenta o servicio degradado |
| JWT HS256 al final | Implementación concisa sin añadir un proveedor de identidad | Requiere proteger secretos y rotarlos; no equivale a SSO |
| Render gratuito como primera ruta | Menor esfuerzo inicial y despliegue directo de contenedores/estático | Suspensión por inactividad, cold start y cuotas; no producción |

## Criterios de cierre

Consideraré la implementación terminada cuando las dos API cumplan los contratos, Go reconstruya la matriz a partir de QR dentro de tolerancia, Node entregue las cinco estadísticas requeridas, Compose arranque el sistema completo, JWT proteja las operaciones, el frontend complete el flujo y las pruebas automatizadas cubran resultados, validación, autenticación y fallos entre servicios. La documentación final explicará cómo reproducir cada paso y estará redactada en primera persona.

## Referencias de despliegue consultadas

- [Render: despliegue gratuito](https://render.com/docs/free) — servicios Docker/web y static sites gratuitos; suspensión tras 15 minutos de inactividad y reactivación aproximada de un minuto. El frontend estático no tiene ese spin down.
- [Angular: versiones soportadas](https://angular.dev/reference/releases) y [compatibilidad](https://angular.dev/reference/versions) — Angular 21 se encuentra en LTS; fijaré una versión de Node.js compatible durante la implementación.
- [Render: Docker](https://render.com/docs/docker) — construcción y despliegue desde Dockerfile.
- [Google Cloud Run: precios](https://cloud.google.com/run/pricing) — alternativa basada en consumo con cuota gratuita, pero cobro por encima de ella y posibles cargos de red.
