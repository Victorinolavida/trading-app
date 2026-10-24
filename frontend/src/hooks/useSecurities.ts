import { useQuery } from "@tanstack/react-query";
import { getSecuritiesActive, getSecuritiesByParams, getSecurityByParams, type Security } from "../lib/api";
import { useDebounce } from "./useDebounce";

export function useActiveSecurities() {
  return useQuery({
    queryKey: ["securities", "active"],
    queryFn: () => getSecuritiesActive(),
  });
}

export function useSearchSecurities(searchTerm: string, delay = 300) {
  const debounced = useDebounce(searchTerm.trim(), delay);
  return useQuery({
    queryKey: ["securities", "search", debounced],
    queryFn: async () => {
      if (debounced.length === 0) return { securities: [] };
      try {
        const res = await getSecuritiesByParams({ q: debounced });
        if (res.securities?.length) return res;
      } catch {}
      try {
        return await getSecuritiesByParams({ isin: debounced });
      } catch {
        return { securities: [] };
      }
    },
    enabled: debounced.length > 0,
    staleTime: 10_000,
  });
}

export function useSearchSecurityByQuery(searchTerm: string, delay = 300) {
  const debounced = useDebounce(searchTerm.trim(), delay);
  return useQuery({
    queryKey: ["securities", "search-single", debounced],
    queryFn: async () => {
      if (debounced.length === 0) return null;
      try {
        const res = await getSecuritiesByParams({ q: debounced });
        if (res.securities && res.securities.length > 0) {
          return res.securities[0] as Security;
        }
      } catch {}
      try {
        return await getSecurityByParams({ isin: debounced });
      } catch {
        return null;
      }
    },
    enabled: debounced.length > 0,
    staleTime: 10_000,
  });
}
