import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authApi } from "./api";
import { AuthUser } from "./types";
import { toast } from "sonner";
import { getErrorMessage } from "../../lib/apiError";

export const authKeys = Object.freeze({
  whoami: ["auth", "whoami"],
  user: ["auth", "user"],
});

export const useWhoami = () =>
  useQuery({
    queryKey: authKeys.whoami,
    queryFn: () => authApi.whoami(),
    retry: false, // 401 means logged out, don't retry
    staleTime: 5 * 60_000,
  });

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      authApi.login(email, password),
    onSuccess: (user: AuthUser) => {
      queryClient.setQueryData(authKeys.user, {
        user_id: user.user_id,
        first_name: user.first_name,
        last_name: user.last_name,
        email: "",
      });
    },
    onError: (error) => toast.error(getErrorMessage(error)),
  });
}

export function useRegister() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      email,
      firstName,
      lastName,
      password,
    }: {
      email: string;
      firstName: string;
      lastName: string;
      password: string;
    }) => authApi.register(email, firstName, lastName, password),
    onSuccess: (user: AuthUser) => {
      queryClient.setQueryData(authKeys.user, {
        user_id: user.user_id,
        first_name: user.first_name,
        last_name: user.last_name,
        email: "",
      });
    },
    onError: (error) => toast.error(getErrorMessage(error)),
  });
}
