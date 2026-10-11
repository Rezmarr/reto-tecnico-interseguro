import { getHealth } from "./controllers/controller.js";

export function registerHealthRoutes(app) {
  app.get("/health", getHealth);
}
