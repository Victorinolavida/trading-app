import { AuthUser } from "../../lib/api";
import { getCsrfToken, needsCsrfHeader } from "../../lib/csrf";
import { ApiClient, ApiClientOptions } from "../../lib/http";
import { User } from "./types";

const baseUrl = "api/v0";

const api = new ApiClient({
  baseUrl: baseUrl,
  getCsrfToken: getCsrfToken,
  needsCsrfHeader: needsCsrfHeader,
} as ApiClientOptions);

export const authApi = {
  whoami: () => api.get<User>("/v0/auth/whoami"),
  login: (email: string, password: string) =>
    api.post<AuthUser>("/v0/auth/login", { email, password }),
  register: (
    email: string,
    firstName: string,
    lastName: string,
    password: string,
  ) =>
    api.post<AuthUser>("/v0/auth/register", {
      email,
      first_name: firstName,
      last_name: lastName,
      password,
    }),
};
