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

export interface Security {
  id: string;
  isin: string;
  name: string;
  security_type: string;
  serie: string;
  issue_date: string;
  maturity_date: string;
  nominal_value: number;
  currency: string;
  fx: number;
  enable: boolean;
}

export interface SecuritiesResponse {
  securities: Security[];
}

export function getSecuritiesActive(): Promise<SecuritiesResponse> {
  return request<SecuritiesResponse>("/securities/active");
}

export function getSecuritiesByParams(params: {
  id?: string;
  isin?: string;
  serie?: string;
  security_type?: string;
  name?: string;
  q?: string;
  query?: string;
} = {}): Promise<SecuritiesResponse> {
  const query = new URLSearchParams();
  if (params.id) query.set("id", params.id);
  if (params.isin) query.set("isin", params.isin);
  if (params.serie) query.set("serie", params.serie);
  if (params.security_type) query.set("security_type", params.security_type);
  if (params.name) query.set("name", params.name);
  const q = params.q ?? params.query;
  if (q) query.set("q", q);
  const qs = query.toString();
  return request<SecuritiesResponse>(`/securities${qs ? `?${qs}` : ""}`);
}

export function getSecurityByParams(params: {
  id?: string;
  isin?: string;
  serie?: string;
  security_type?: string;
  name?: string;
  q?: string;
  query?: string;
} = {}): Promise<Security> {
  const query = new URLSearchParams();
  if (params.id) query.set("id", params.id);
  if (params.isin) query.set("isin", params.isin);
  if (params.serie) query.set("serie", params.serie);
  if (params.security_type) query.set("security_type", params.security_type);
  if (params.name) query.set("name", params.name);
  const q = params.q ?? params.query;
  if (q) query.set("q", q);
  const qs = query.toString();
  return request<Security>(`/securities/search${qs ? `?${qs}` : ""}`);
}

export interface CreateOrderRequest {
  trader_id?: string;
  security_id: string;
  direction: "bid" | "ask";
  order_type: "limit" | "market";
  price?: string;
  yield?: string;
  quantity: string;
}

export interface CreateOrderResponse {
  order_group_id: string;
  order_id: string;
  status: string;
}

export interface CancelOrderRequest {
  order_group_id: string;
  trader_id?: string;
}

export interface CancelOrderResponse {
  order_group_id: string;
  status: string;
}

export interface Order {
  id: string;
  order_group_id: string;
  trader_id: string;
  security_id: string;
  direction: string;
  order_type: string;
  price: string;
  yield: string;
  quantity: string;
  nominal_value: string;
  fx: string;
  status: string;
}

export interface OrderGroupDetail {
  id: string;
  public_id: string;
  trader_id: string;
  security_id: string;
  direction: string;
  order_type: string;
  allocation_pc: string;
  allocation_status: string;
  order_group_status: string;
  weighted_price: string;
  weighted_yield: string;
  weighted_fx: string;
  orders: Order[];
}

export function createOrder(req: CreateOrderRequest): Promise<CreateOrderResponse> {
  return request<CreateOrderResponse>("/orders", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function cancelOrder(req: CancelOrderRequest): Promise<CancelOrderResponse> {
  return request<CancelOrderResponse>("/orders/cancel", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function getOrderGroup(orderGroupId: string): Promise<OrderGroupDetail> {
  return request<OrderGroupDetail>(`/orders/${orderGroupId}`);
}
