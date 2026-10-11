export class InputError extends Error {
  constructor(code, message) {
    super(message);
    this.status = 400;
    this.code = code;
    this.publicMessage = message;
  }
}

function validateMatrix(matrix, name, config) {
  if (!Array.isArray(matrix) || matrix.length === 0) {
    throw new InputError("INVALID_MATRIX", `${name} debe ser una matriz rectangular no vacía.`);
  }
  if (matrix.length > config.maxMatrixRows) {
    throw new InputError("INVALID_MATRIX", `${name} supera el máximo de ${config.maxMatrixRows} filas.`);
  }
  const columns = Array.isArray(matrix[0]) ? matrix[0].length : -1;
  if (columns <= 0 || columns > config.maxMatrixColumns) {
    throw new InputError("INVALID_MATRIX", `${name} tiene una cantidad de columnas no permitida.`);
  }
  if (matrix.length > Math.floor(config.maxMatrixElements / columns)) {
    throw new InputError("INVALID_MATRIX", `${name} supera el máximo de ${config.maxMatrixElements} elementos.`);
  }
  for (let row = 0; row < matrix.length; row += 1) {
    if (!Array.isArray(matrix[row]) || matrix[row].length !== columns) {
      throw new InputError("INVALID_MATRIX", `${name} debe ser rectangular.`);
    }
    for (let column = 0; column < columns; column += 1) {
      if (typeof matrix[row][column] !== "number" || !Number.isFinite(matrix[row][column])) {
        throw new InputError("INVALID_MATRIX", `El elemento de fila ${row} y columna ${column} de ${name} debe ser un número finito.`);
      }
    }
  }
  return { rows: matrix.length, columns };
}

export function validateStatisticsRequest(body, config) {
  if (body === null || typeof body !== "object" || Array.isArray(body) || !body.matrices ||
      typeof body.matrices !== "object" || Array.isArray(body.matrices)) {
    throw new InputError("INVALID_PAYLOAD", "El cuerpo debe contener un objeto matrices con q y r.");
  }
  const keys = Object.keys(body.matrices).sort();
  if (keys.length !== 2 || keys[0] !== "q" || keys[1] !== "r") {
    throw new InputError("INVALID_PAYLOAD", "matrices debe contener únicamente q y r.");
  }
  const qShape = validateMatrix(body.matrices.q, "q", config);
  const rShape = validateMatrix(body.matrices.r, "r", config);
  const total = qShape.rows * qShape.columns + rShape.rows * rShape.columns;
  if (total > config.maxMatrixElements) {
    throw new InputError("INVALID_MATRIX", `q y r superan el máximo conjunto de ${config.maxMatrixElements} elementos.`);
  }
  if (qShape.rows !== qShape.columns) throw new InputError("INVALID_MATRIX", "q debe ser cuadrada.");
  return body.matrices;
}
