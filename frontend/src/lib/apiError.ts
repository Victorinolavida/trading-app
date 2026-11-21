import { ApiError } from "./http";

export function getErrorMessage(error: unknown): string | null {
  if (!error) return null;

  if (error instanceof ApiError) {
    if (error.status === 0) {
      return "Can't reach the server. Check your connection.";
    }
    if (error.status === 500) {
      return "Internal server error. Please try again later or contact support.";
    }
    return error.message;
  }

  return "Something went wrong. Please try again.";
}