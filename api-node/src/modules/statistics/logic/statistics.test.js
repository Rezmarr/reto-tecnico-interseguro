import assert from "node:assert/strict";
import test from "node:test";
import { calculateStatistics, NumericalError } from "./statistics.js";

test("calcula agregados conjuntos y aplicabilidad de diagonalidad", () => {
  const result = calculateStatistics({
    q: [[1, 0], [0, 1]],
    r: [[1, 2, 3], [0, 4, 5]],
  }, 1e-10);
  assert.deepEqual(
    { minimum: result.minimum, maximum: result.maximum, sum: result.sum, average: result.average, count: result.count },
    { minimum: 0, maximum: 5, sum: 17, average: 1.7, count: 10 },
  );
  assert.deepEqual(result.diagonal, {
    q: { applicable: true, isDiagonal: true },
    r: { applicable: false, isDiagonal: null },
    anyMatrixDiagonal: true,
  });
});

test("usa la tolerancia configurada al comprobar matrices diagonales", () => {
  const result = calculateStatistics({ q: [[1, 0.001], [0, 1]], r: [[1, 0], [0, 1]] }, 0.01);
  assert.equal(result.diagonal.q.isDiagonal, true);
  assert.equal(result.diagonal.r.isDiagonal, true);
  assert.equal(result.diagonal.anyMatrixDiagonal, true);
});

test("rechaza sumas agregadas no finitas", () => {
  assert.throws(() => calculateStatistics({ q: [[Number.MAX_VALUE]], r: [[Number.MAX_VALUE]] }, 0), NumericalError);
});
