import { ApiError } from "./http";

export function getErrorMessage(error: unknown): string | null {
  if (!error) return null;

  if (error instanceof ApiError) {
    return error.status === 0
      ? "Can't reach the server. Check your connection."
      : error.message;
  }

  return "Something went wrong. Please try again.";
}