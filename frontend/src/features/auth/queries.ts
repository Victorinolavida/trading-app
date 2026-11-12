import { queryOptions } from "@tanstack/react-query";
import { authApi } from "./api";

export const authKeys = Object.freeze({
  user: ["auth", "user"],
});

export const authQueries = {
  whoami: queryOptions({
    queryKey: authKeys.user,
    queryFn: () => authApi.whoami(),
    retry: false, // 401 means logged out, don't retry
    staleTime: 5 * 60_000,
  }),
};
