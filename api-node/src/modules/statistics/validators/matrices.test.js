import assert from "node:assert/strict";
import test from "node:test";
import { InputError, validateStatisticsRequest } from "./matrices.js";

const config = { maxMatrixRows: 5, maxMatrixColumns: 5, maxMatrixElements: 20 };

test("accepts rectangular R and square Q within the combined element limit", () => {
  const matrices = { q: [[1, 0], [0, 1]], r: [[1, 2, 3], [0, 4, 5]] };
  assert.equal(validateStatisticsRequest({ matrices }, config), matrices);
});

test("rejects malformed contracts and invalid matrix shapes", () => {
  const cases = [
    [{}, "INVALID_PAYLOAD"],
    [{ matrices: { q: [[1]], r: [[1]], extra: [] } }, "INVALID_PAYLOAD"],
    [{ matrices: { q: [[1, 2]], r: [[1], [3]] } }, "INVALID_MATRIX"],
    [{ matrices: { q: [[1, 2]], r: [[1]] } }, "INVALID_MATRIX"],
    [{ matrices: { q: [[1]], r: [["2"]] } }, "INVALID_MATRIX"],
    [{ matrices: { q: [[1], [2], [3]], r: [[1]] } }, "INVALID_MATRIX"],
    [{ matrices: { q: [[1, 0, 0], [0, 1, 0], [0, 0, 1]], r: [[1, 2, 3, 4], [0, 1, 2, 3], [0, 0, 1, 2]] } }, "INVALID_MATRIX"],
  ];
  for (const [body, code] of cases) {
    assert.throws(() => validateStatisticsRequest(body, config), (error) => error instanceof InputError && error.code === code);
  }
});
