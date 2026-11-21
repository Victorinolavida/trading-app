import { ApiClient, type RequestOptions } from "../../lib/http";
import type {
  PriceCalculationRequest,
  PriceResult,
  YieldCalculationRequest,
} from "./types";

const api = new ApiClient();

export const calculatorApi = {
  calculatePrice: (request: PriceCalculationRequest, options: RequestOptions = {}) =>
    api.post<PriceResult>("/v0/calculator/price", request, options),
  calculateYield: (request: YieldCalculationRequest, options: RequestOptions = {}) =>
    api.post<number>("/v0/calculator/yield", request, options),
};
