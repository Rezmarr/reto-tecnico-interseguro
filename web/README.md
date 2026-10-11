# Frontend Angular 21

Interfaz para ingresar una matriz mediante una cuadrícula editable con selector de filas y columnas, consumir la API Go y visualizar Q, R y sus estadísticas. El frontend convierte las celdas a la matriz numérica que espera el contrato de la API, documentado en [../docs/CONTRATO_API.md](../docs/CONTRATO_API.md).

## Ejecución local

Requiere Node.js 22.12 o posterior.

```powershell
npm ci
npm start
```

La aplicación abre en `http://localhost:4200` y usa Go en `http://localhost:8080` y Node en `http://localhost:3000`. En desarrollo, Go y Node permiten por defecto el origen `http://localhost:4200`.

La URL de Node solo se usa para consultar `/health` y anticipar el arranque. Go sigue siendo quien solicita las estadísticas a Node.

Al iniciar una factorización consulto los endpoints de salud en paralelo y luego envío la matriz a Go. Si el inicio tarda más de unos segundos, muestro que los servicios están iniciando en Render y que una instancia gratuita puede tardar hasta un minuto después de inactividad. La API sigue siendo la autoridad para aplicar los límites de dimensiones.

## Pruebas

`npm test` utiliza el builder de pruebas de Angular 21 y Vitest en jsdom. Verifica inicio de sesión, persistencia/cierre de sesión, el envío de JWT en HTTP y la entrada, redimensionado y envío de la matriz desde la cuadrícula.

# Inicio de sesión

El frontend solicita un JWT a `POST /api/v1/auth/token` en la API Go y envía el token como Bearer al calcular. La sesión se conserva en `sessionStorage` durante la pestaña y se borra al cerrar sesión o detectar expiración.

## Separación de páginas

- `src/app/app.component`: coordina la sesión activa y selecciona qué página mostrar.
- `src/app/features/auth/`: captura credenciales y solicita el token.
- `src/app/features/calculation/`: administra la matriz, las conexiones, el cálculo y la presentación de resultados. Recibe el token y emite una solicitud de cierre de sesión.
- `src/app/shared/`: centraliza los contratos de datos y las llamadas HTTP compartidas.
