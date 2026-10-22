import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiError, login, register, whoami, type AuthUser, type User } from "../lib/api";
import { clearToken, getToken } from "../lib/token";

export const userKey = ["user"];

export function useSession() {
  const token = getToken();

  return useQuery<User | null>({
    queryKey: userKey,
    queryFn: async () => {
      if (!getToken()) return null;
      try {
        return await whoami();
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) return null;
        throw error;
      }
    },
    enabled: Boolean(token),
    staleTime: 30_000,
    retry: (failureCount, error) =>
      !(error instanceof ApiError && error.status === 401) && failureCount < 2,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      login(email, password),
    onSuccess: (user: AuthUser) => {
      queryClient.setQueryData(userKey, {
        user_id: user.user_id,
        first_name: user.first_name,
        last_name: user.last_name,
        email: "",
      });
    },
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
    }) => register(email, firstName, lastName, password),
    onSuccess: (user: AuthUser) => {
      queryClient.setQueryData(userKey, {
        user_id: user.user_id,
        first_name: user.first_name,
        last_name: user.last_name,
        email: "",
      });
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return () => {
    clearToken();
    queryClient.setQueryData(userKey, null);
    queryClient.clear();
  };
}

