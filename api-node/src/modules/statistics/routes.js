import { createStatisticsController } from "./controllers/controller.js";
import { createAuthentication } from "../../shared/auth/middleware.js";

export function registerStatisticsRoutes(app, config) {
  app.post("/api/v1/statistics", createAuthentication(config), createStatisticsController(config));
}
