import express from "express";
import { createErrorHandler } from "./shared/http/error-handler.js";
import { requestId } from "./shared/http/request-id.js";
import { corsPolicy } from "./shared/http/cors.js";
import { registerHealthRoutes } from "./modules/health/routes.js";
import { registerStatisticsRoutes } from "./modules/statistics/routes.js";

export function createApp(config) {
  const app = express();
  app.disable("x-powered-by");
  app.use(requestId);
  app.use(corsPolicy(config.frontendOrigins));
  app.use(express.json({ limit: config.bodyLimit, strict: true }));
  registerHealthRoutes(app);
  registerStatisticsRoutes(app, config);
  app.use((req, _res, next) => {
    const error = new Error(`Ruta no encontrada: ${req.method} ${req.path}`);
    error.status = 404;
    error.code = "NOT_FOUND";
    next(error);
  });
  app.use(createErrorHandler());
  return app;
}
