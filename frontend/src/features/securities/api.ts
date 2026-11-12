import { ApiClient, RequestOptions } from "../../lib/http";
import { SearchParams, SecuritiesResponse, Security } from "./types";

const api = new ApiClient();


const SEARCH_KEYS = ["id", "isin", "serie", "security_type", "name"] as const;

function buildSearchQuery(params: SearchParams): string {
  const query = new URLSearchParams();

  for (const key of SEARCH_KEYS) {
    const value = params[key]?.trim();
    if (value) query.set(key, value);
  }

  // `||` (not `??`) so an empty q falls back to query
  const q = (params.q || params.query)?.trim();
  if (q) query.set("q", q);

  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

export const securitiesApi = {
  getSecuritiesActive: ({ signal }: RequestOptions = {}) =>
    api.get<SecuritiesResponse>("/v0/securities/active", { signal }),

  getSecuritiesByParams: (
    params: SearchParams,
    { signal }: RequestOptions = {},
  ) =>
    api.get<SecuritiesResponse>(`/v0/securities${buildSearchQuery(params)}`, {
      signal,
    }),

  getSecurityByParams: (
    params: SearchParams,
    { signal }: RequestOptions = {},
  ) =>
    api.get<Security>(`/v0/securities/search${buildSearchQuery(params)}`, {
      signal,
    }),
};
