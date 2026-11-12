import axios, {
  AxiosError,
  type AxiosAdapter,
  type AxiosInstance,
  type AxiosRequestConfig,
} from "axios";

import { CSRF_COOKIE_NAME } from "./csrf";

export class ApiError extends Error {
  constructor(
    public readonly status: number, // 0 = no response (network error, timeout)
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface ApiClientOptions {
  adapter?: AxiosAdapter; // injectable for tests
  timeoutMs?: number;
  csrfHeaderName?: string;
}

export class ApiClient {
  private readonly http: AxiosInstance;

  constructor(opts: ApiClientOptions = {}) {
    this.http = axios.create({
      baseURL: "/api",
      withCredentials: true, // send the httpOnly session cookie
      timeout: opts.timeoutMs ?? 10_000, // 10s
      adapter: opts.adapter,
      xsrfCookieName: CSRF_COOKIE_NAME,
      xsrfHeaderName: opts.csrfHeaderName ?? "X-CSRF-Token",
    });

    // axios's built-in xsrf handling echoes the csrf cookie back in the
    // X-CSRF-Token header on same-origin requests; the cookie is not HttpOnly
    // on purpose so it can be read here.
    this.http.interceptors.response.use(
      (res) => res,
      (err: unknown) => {
        const apiErr = this.toApiError(err);
        return Promise.reject(apiErr);
      },
    );
  }

  get<T>(path: string, config?: AxiosRequestConfig) {
    return this.request<T>({ ...config, method: "GET", url: path });
  }

  post<T>(path: string, body?: unknown, config?: AxiosRequestConfig) {
    return this.request<T>({
      ...config,
      method: "POST",
      url: path,
      data: body,
    });
  }

  put<T>(path: string, body?: unknown, config?: AxiosRequestConfig) {
    return this.request<T>({ ...config, method: "PUT", url: path, data: body });
  }

  patch<T>(path: string, body?: unknown, config?: AxiosRequestConfig) {
    return this.request<T>({
      ...config,
      method: "PATCH",
      url: path,
      data: body,
    });
  }

  delete<T = void>(path: string, config?: AxiosRequestConfig) {
    return this.request<T>({ ...config, method: "DELETE", url: path });
  }

  private async request<T>(config: AxiosRequestConfig): Promise<T> {
    return this.http.request<T>(config).then((res) => {
      if (res.status === 204) return undefined as T;
      return res.data;
    });
  }

  private toApiError(err: unknown): unknown {
    if (axios.isCancel(err)) return err; // let aborts pass through untouched
    if (!(err instanceof AxiosError)) {
      return err;
    }

    if (!err.response) {
      return new ApiError(
        0,
        err.code === "ECONNABORTED" ? "request timed out" : "network error",
      );
    }

    const { status, data, statusText } = err.response;
    const message =
      data !== null &&
      typeof data === "object" &&
      "error" in data &&
      typeof data.error === "string"
        ? data.error
        : statusText || `request failed (${status})`;

    return new ApiError(status, message);
  }
}

export type RequestOptions = { signal?: AbortSignal };
