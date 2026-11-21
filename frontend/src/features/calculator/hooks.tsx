import { useQuery } from "@tanstack/react-query";
import { calculatorQueries } from "./queries";

export const useCalculatePrice = (params: {
  securityId: string;
  settlementDate: string;
  yield: string;
  enabled: boolean
}) => useQuery(calculatorQueries.price(params));

export const useCalculateYield = (params: {
  securityId: string;
  settlementDate: string;
  cleanPrice: string;
  enabled: boolean
}) => useQuery(calculatorQueries.yield(params));
