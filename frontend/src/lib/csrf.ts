const CSRF_COOKIE_NAME = import.meta.env.VITE_CSRF_COOKIE_NAME ?? "csrf_token";

const UNSAFE_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);

/**
 * Reads the csrf cookie the gateway set. The cookie is not HttpOnly on
 * purpose: the frontend has to echo it back in the X-CSRF-Token header.
 */
export function getCsrfToken(): string | null {
  const prefix = `${CSRF_COOKIE_NAME}=`;

  for (const part of document.cookie.split(";")) {
    const entry = part.trim();
    if (entry.startsWith(prefix)) {
      return decodeURIComponent(entry.slice(prefix.length));
    }
  }

  return null;
}

/**
 * The csrf token is only required on state changing requests.
 */
export function needsCsrfHeader(method: string): boolean {
  return UNSAFE_METHODS.has(method.toUpperCase());
}
