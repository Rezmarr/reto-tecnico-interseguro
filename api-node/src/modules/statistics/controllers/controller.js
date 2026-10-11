import { statisticsResponse } from "../contracts/response.js";
import { calculateStatistics } from "../logic/statistics.js";
import { validateStatisticsRequest } from "../validators/matrices.js";

export function createStatisticsController(config) {
  return (req, res, next) => {
    try {
      const matrices = validateStatisticsRequest(req.body, config);
      return res.json(statisticsResponse(calculateStatistics(matrices, config.diagonalTolerance)));
    } catch (error) {
      return next(error);
    }
  };
}
