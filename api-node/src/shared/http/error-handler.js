export function createErrorHandler() {
  return (error, req, res, _next) => {
    let status = Number.isInteger(error.status) ? error.status : 500;
    let code = error.code ?? "INTERNAL_ERROR";
    let message = error.publicMessage ?? "Ocurrió un error interno.";
    if (error.type === "entity.too.large") {
      status = 413;
      code = "PAYLOAD_TOO_LARGE";
      message = "El cuerpo de la petición excede el límite permitido.";
    } else if (error.type === "entity.parse.failed" || error instanceof SyntaxError) {
      status = 400;
      code = "INVALID_PAYLOAD";
      message = "El cuerpo debe contener JSON válido.";
    }
    if (status >= 500) console.error("Unhandled request error", { requestId: req.requestId, error });
    return res.status(status).json({ error: { code, message, requestId: req.requestId } });
  };
}
