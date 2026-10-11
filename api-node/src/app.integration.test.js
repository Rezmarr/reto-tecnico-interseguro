import assert from "node:assert/strict";
import { once } from "node:events";
import test from "node:test";
import { SignJWT } from "jose";
import { createApp } from "./app.js";

const secret = Buffer.alloc(32, 7);
const config = {
  port: 0,
  jwtSecret: secret,
  jwtIssuer: "integration-test",
  serviceApiKey: "integration-service-key-with-32-bytes-minimum",
  diagonalTolerance: 1e-10,
  maxMatrixRows: 10,
  maxMatrixColumns: 10,
  maxMatrixElements: 100,
  bodyLimit: "1mb",
  frontendOrigins: ["http://localhost:4200"],
};

async function createToken() {
  return new SignJWT({})
    .setProtectedHeader({ alg: "HS256" })
    .setIssuer(config.jwtIssuer)
    .setSubject("demo-user")
    .setIssuedAt()
    .setExpirationTime("5m")
    .sign(secret);
}

test("salud pública y estadísticas protegidas por JWT de usuario y credencial de servicio", async (t) => {
  const server = createApp(config).listen(0, "127.0.0.1");
  await once(server, "listening");
  t.after(() => new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve())));
  const base = `http://127.0.0.1:${server.address().port}`;

  const health = await fetch(`${base}/health`);
  assert.equal(health.status, 200);

  const payload = { matrices: { q: [[1, 0], [0, 1]], r: [[1, 2, 3], [0, 4, 5]] } };
  const endpoint = `${base}/api/v1/statistics`;
  const unauthenticated = await fetch(endpoint, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) });
  assert.equal(unauthenticated.status, 401);

  const token = await createToken();
  const missingServiceKey = await fetch(endpoint, {
    method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: JSON.stringify(payload),
  });
  assert.equal(missingServiceKey.status, 401);

  const invalidToken = await fetch(endpoint, {
    method: "POST", headers: { "Content-Type": "application/json", Authorization: "Bearer invalid", "X-Service-Token": config.serviceApiKey }, body: JSON.stringify(payload),
  });
  assert.equal(invalidToken.status, 401);

  const expiredToken = await new SignJWT({})
    .setProtectedHeader({ alg: "HS256" })
    .setIssuer(config.jwtIssuer)
    .setSubject("demo-user")
    .setIssuedAt(Math.floor(Date.now() / 1000) - 120)
    .setExpirationTime(Math.floor(Date.now() / 1000) - 60)
    .sign(secret);
  const expired = await fetch(endpoint, {
    method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${expiredToken}`, "X-Service-Token": config.serviceApiKey }, body: JSON.stringify(payload),
  });
  assert.equal(expired.status, 401);

  const accepted = await fetch(endpoint, {
    method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, "X-Service-Token": config.serviceApiKey }, body: JSON.stringify(payload),
  });
  assert.equal(accepted.status, 200);
  const statistics = await accepted.json();
  assert.equal(statistics.count, 10);
  assert.equal(statistics.sum, 17);
  assert.equal(statistics.diagonal.r.applicable, false);

  const invalidMatrix = await fetch(endpoint, {
    method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, "X-Service-Token": config.serviceApiKey },
    body: JSON.stringify({ matrices: { q: [[1, 2]], r: [[1]] } }),
  });
  assert.equal(invalidMatrix.status, 400);
});
