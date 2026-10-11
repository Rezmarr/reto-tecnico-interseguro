const ALLOWED_METHODS = "GET,POST,OPTIONS";
const ALLOWED_HEADERS = "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Service-Token";

export function corsPolicy(origins) {
  const allowed = new Set(origins);
  return (req, res, next) => {
    const origin = req.get("Origin");
    if (origin && allowed.has(origin)) {
      res.set("Access-Control-Allow-Origin", origin);
      res.set("Vary", "Origin");
      res.set("Access-Control-Allow-Methods", ALLOWED_METHODS);
      res.set("Access-Control-Allow-Headers", ALLOWED_HEADERS);
    }
    if (req.method === "OPTIONS") return res.sendStatus(204);
    return next();
  };
}
