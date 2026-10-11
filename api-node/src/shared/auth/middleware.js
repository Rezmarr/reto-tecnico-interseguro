import { timingSafeEqual } from "node:crypto";
import { jwtVerify } from "jose";

const unauthorized = (message = "Se requiere autenticación válida.") => {
  const error = new Error(message);
  error.status = 401;
  error.code = "UNAUTHORIZED";
  error.publicMessage = message;
  return error;
};

function bearer(req) {
  const value = req.get("Authorization") ?? "";
  const match = /^Bearer ([^\s]+)$/.exec(value);
  if (!match) throw unauthorized();
  return match[1];
}

export function createAuthentication(config) {
  const key = config.jwtSecret;
  return async (req, _res, next) => {
    try {
      const token = bearer(req);
      await jwtVerify(token, key, {
        algorithms: ["HS256"], issuer: config.jwtIssuer,
        requiredClaims: ["sub", "iat", "exp"],
      });
      const serviceToken = req.get("X-Service-Token") ?? "";
      const expected = Buffer.from(config.serviceApiKey);
      const actual = Buffer.from(serviceToken);
      if (actual.length !== expected.length || !timingSafeEqual(actual, expected)) throw unauthorized();
      return next();
    } catch {
      return next(unauthorized());
    }
  };
}
