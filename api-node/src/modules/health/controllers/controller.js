import { healthResponse } from "../contracts/response.js";

export function getHealth(_req, res) {
  return res.json(healthResponse());
}
