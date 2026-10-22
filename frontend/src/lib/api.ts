import { getCsrfToken, needsCsrfHeader } from "./csrf";

const API_BASE = "/api/v0";

export interface AuthUser {
  user_id: string;
  first_name: string;
  last_name: string;
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

  if (needsCsrfHeader(init.method ?? "GET")) {
    const csrfToken = getCsrfToken();
    if (csrfToken) headers.set("X-CSRF-Token", csrfToken);
  }

  // The session lives in an httpOnly cookie, so the browser has to send it.
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  if (!res.ok) throw new ApiError(res.status, await parseError(res));

  if (res.status === 204) return undefined as T;

  return (await res.json()) as T;
}

export async function login(email: string, password: string): Promise<AuthUser> {
  return request<AuthUser>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function register(
  email: string,
  firstName: string,
  lastName: string,
  password: string,
): Promise<AuthUser> {
  return request<AuthUser>("/auth/register", {
    method: "POST",
    body: JSON.stringify({
      email,
      first_name: firstName,
      last_name: lastName,
      password,
    }),
  });
}

export function whoami(): Promise<User> {
  return request<User>("/auth/whoami");
}

export function logout(): Promise<void> {
  return request<void>("/auth/logout", { method: "POST" });
}
