import { useQuery } from "@tanstack/react-query";
import { SearchParams } from "./types";
import { securityQueries } from "./queries";

export const useActiveSecurities = () => useQuery(securityQueries.active());
export const useSearchSecuritiesByParams = (params: SearchParams = {}) =>
  useQuery(securityQueries.search(params));
export const useSecurityDetail = (params: SearchParams) =>
  useQuery(securityQueries.detail(params));
