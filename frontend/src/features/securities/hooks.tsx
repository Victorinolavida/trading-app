import { useQuery } from "@tanstack/react-query";
import { securitiesApi } from "./api";

export const securitykeys = Object.freeze({
  active: ["securities", "active"],
  search: ["securities", "search"],
});

export function useActiveSecurities() {
  return useQuery({
    queryKey: ["securities", "active"],
    queryFn: () => securitiesApi.getSecuritiesActive(),
  });
}

export function useSearchSecuritiesByTerm(searchTerm: string) {
  const cleanTerm = searchTerm.trim();
  return useQuery({
    queryKey: ["securities", "search", searchTerm],
    queryFn: () => securitiesApi.getSecuritiesByParams({ q: cleanTerm }),
    enabled: cleanTerm.length > 0,
    staleTime: 10_000,
  });
}