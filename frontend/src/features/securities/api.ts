import { ApiClient } from "../../lib/http";
import { SearchParams, SecuritiesResponse, Security } from "./types";

const api = new ApiClient();

export const securitiesApi = {
  getSecuritiesActive: () =>
    api.get<SecuritiesResponse>("/v0/securities/active"),
  getSecuritiesByParams: (params: SearchParams) => {
    const query = new URLSearchParams();
    if (params.id) query.set("id", params.id);
    if (params.isin) query.set("isin", params.isin);
    if (params.serie) query.set("serie", params.serie);
    if (params.security_type) query.set("security_type", params.security_type);
    if (params.name) query.set("name", params.name);
    const q = params.q ?? params.query;
    if (q) query.set("q", q);
    const qs = query.toString();
    return api.get<SecuritiesResponse>(`/v0/securities${qs ? `?${qs}` : ""}`);
  },
  getSecurityByParams: (params: SearchParams) => {
    const query = new URLSearchParams();
    if (params.id) query.set("id", params.id);
    if (params.isin) query.set("isin", params.isin);
    if (params.serie) query.set("serie", params.serie);
    if (params.security_type) query.set("security_type", params.security_type);
    if (params.name) query.set("name", params.name);
    const q = params.q ?? params.query;
    if (q) query.set("q", q);
    const qs = query.toString();
    return api.get<Security>(`v0/securities/search${qs ? `?${qs}` : ""}`);
  },
};
