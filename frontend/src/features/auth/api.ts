import { AuthUser } from "../../lib/api";
import { ApiClient } from "../../lib/http";
import { User } from "./types";

const api = new ApiClient();

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
  logout: () => api.post<void>("/v0/auth/logout"),
};
