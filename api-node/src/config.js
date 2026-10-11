const DEFAULT_PORT = 3000;
const DEFAULT_TOLERANCE = 1e-10;

function positiveInteger(name, fallback) {
  const value = process.env[name];
  if (value === undefined || value === "") return fallback;
  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed) || parsed <= 0) throw new Error(`${name} debe ser un entero positivo`);
  return parsed;
}

function nonNegativeFiniteNumber(name, fallback) {
  const value = process.env[name];
  if (value === undefined || value === "") return fallback;
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed < 0) throw new Error(`${name} debe ser un número finito no negativo`);
  return parsed;
}

export function loadConfig() {
  const port = positiveInteger("PORT", DEFAULT_PORT);
  if (port > 65535) throw new Error("PORT debe ser un entero entre 1 y 65535");
  let encodedJwtSecret = (process.env.JWT_SECRET ?? "").trim();
  if ((encodedJwtSecret.startsWith('"') && encodedJwtSecret.endsWith('"')) ||
      (encodedJwtSecret.startsWith("'") && encodedJwtSecret.endsWith("'"))) {
    encodedJwtSecret = encodedJwtSecret.slice(1, -1);
  }
  const jwtSecret = Buffer.from(encodedJwtSecret, "base64");
  const serviceApiKey = process.env.SERVICE_API_KEY ?? "";
  if (jwtSecret.toString("base64") !== encodedJwtSecret) {
    throw new Error("JWT_SECRET debe ser Base64 estándar válido; quita las comillas adicionales si las hay");
  }
  if (jwtSecret.length < 32) throw new Error(`JWT_SECRET debe decodificarse a por lo menos 32 bytes (actualmente tiene ${jwtSecret.length})`);
  if (Buffer.byteLength(serviceApiKey) < 32) throw new Error("SERVICE_API_KEY debe contener por lo menos 32 bytes");
  return Object.freeze({
    port,
    jwtSecret,
    jwtIssuer: process.env.JWT_ISSUER || "reto-tecnico",
    serviceApiKey,
    diagonalTolerance: nonNegativeFiniteNumber("DIAGONAL_TOLERANCE", DEFAULT_TOLERANCE),
    maxMatrixRows: positiveInteger("MAX_MATRIX_ROWS", 100),
    maxMatrixColumns: positiveInteger("MAX_MATRIX_COLUMNS", 100),
    maxMatrixElements: positiveInteger("MAX_MATRIX_ELEMENTS", 10000),
    bodyLimit: "1mb",
    frontendOrigins: (process.env.FRONTEND_ORIGINS ?? "http://localhost:4200,http://127.0.0.1:4200")
      .split(",").map((origin) => origin.trim()).filter(Boolean),
  });
}
