export class NumericalError extends Error {
  constructor() {
    super("Los valores de entrada exceden el rango numérico soportado para calcular estadísticas.");
    this.status = 422;
    this.code = "NUMERICAL_FAILURE";
    this.publicMessage = this.message;
  }
}

function diagonalCheck(matrix, tolerance) {
  if (matrix.length !== matrix[0].length) return { applicable: false, isDiagonal: null };
  for (let row = 0; row < matrix.length; row += 1) {
    for (let column = 0; column < matrix.length; column += 1) {
      if (row !== column && Math.abs(matrix[row][column]) > tolerance) {
        return { applicable: true, isDiagonal: false };
      }
    }
  }
  return { applicable: true, isDiagonal: true };
}

export function calculateStatistics(matrices, tolerance) {
  let minimum = Infinity;
  let maximum = -Infinity;
  let sum = 0;
  let count = 0;
  for (const matrix of Object.values(matrices)) {
    for (const row of matrix) {
      for (const value of row) {
        minimum = Math.min(minimum, value);
        maximum = Math.max(maximum, value);
        sum += value;
        count += 1;
      }
    }
  }
  if (!Number.isFinite(sum) || count === 0) throw new NumericalError();
  const average = sum / count;
  if (!Number.isFinite(average)) throw new NumericalError();
  const q = diagonalCheck(matrices.q, tolerance);
  const r = diagonalCheck(matrices.r, tolerance);
  return {
    minimum, maximum, sum, average, count,
    diagonal: { q, r, anyMatrixDiagonal: (q.applicable && q.isDiagonal) || (r.applicable && r.isDiagonal) },
    tolerance,
  };
}
