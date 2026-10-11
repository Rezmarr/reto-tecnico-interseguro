import { fileURLToPath } from "node:url";
import { createApp } from "./app.js";
import { loadConfig } from "./config.js";

try {
  try {
    process.loadEnvFile(fileURLToPath(new URL("../.env", import.meta.url)));
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
  const config = loadConfig();
  const app = createApp(config);
  app.listen(config.port, "0.0.0.0", () => console.log(`La API de Node está escuchando en el puerto ${config.port}`));
} catch (error) {
  console.error(`No se pudo iniciar la API de Node: ${error.message}`);
  process.exitCode = 1;
}
