import { keepPreviousData, queryOptions } from "@tanstack/react-query";
import { SearchParams } from "./types";

import { securitiesApi } from "./api";

export const securityKeys = Object.freeze({
  all: ["securities"] as const,
  lists: () => [...securityKeys.all, "list"] as const,
  active: () => [...securityKeys.lists(), "active"] as const,
  search: (params: SearchParams) =>
    [...securityKeys.lists(), "search", normalizeParams(params)] as const,

  details: () => [...securityKeys.all, "detail"] as const,
  detail: (params: SearchParams) =>
    [...securityKeys.details(), normalizeParams(params)] as const,
});

function normalizeParams(params: SearchParams): SearchParams {
  return Object.fromEntries(
    Object.entries(params)
      .map(([k, v]) => [k, typeof v === "string" ? v.trim() : v])
      .filter(([, v]) => v !== undefined && v !== null && v !== ""),
  ) as SearchParams;
}
export const securityQueries = {
  active: () =>
    queryOptions({
      queryKey: securityKeys.active(),
      queryFn: ({ signal }) => securitiesApi.getSecuritiesActive({ signal }),
      staleTime: 30_000,
    }),

  search: (params: SearchParams = {}) => {
    const clean = normalizeParams(params);
    return queryOptions({
      queryKey: securityKeys.search(clean),
      queryFn: ({ signal }) =>
        securitiesApi.getSecuritiesByParams(clean, { signal }),
      staleTime: 10_000,
      enabled: Object.entries(clean).length > 0,
      placeholderData: keepPreviousData,
    });
  },

  detail: (params: SearchParams) => {
    const clean = normalizeParams(params);
    return queryOptions({
      queryKey: securityKeys.detail(clean),
      queryFn: ({ signal }) =>
        securitiesApi.getSecurityByParams(clean, { signal }),
      enabled: Object.entries(clean).length > 0,
      staleTime: 60_000,
    });
  },
};
