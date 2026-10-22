import { clearToken, getToken, setToken } from "./token";

const API_BASE = "/api/v0";

export interface AuthUser {
  user_id: string;
  first_name: string;
  last_name: string;
  token: string;
}

export interface User {
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly message: string,
  ) {
    super(message);
  }
}

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: unknown };
    if (typeof body.error === "string") return body.error;
  } catch {
    // ignore body parse failures
  }
  return res.statusText || "request failed";
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (res.status === 401) clearToken();

  if (!res.ok) throw new ApiError(res.status, await parseError(res));

  return (await res.json()) as T;
}

export async function login(email: string, password: string): Promise<AuthUser> {
  const user = await request<AuthUser>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  setToken(user.token);
  return user;
}

export async function register(
  email: string,
  firstName: string,
  lastName: string,
  password: string,
): Promise<AuthUser> {
  const user = await request<AuthUser>("/auth/register", {
    method: "POST",
    body: JSON.stringify({
      email,
      first_name: firstName,
      last_name: lastName,
      password,
    }),
  });
  setToken(user.token);
  return user;
}

export function whoami(): Promise<User> {
  return request<User>("/auth/whoami");
}

