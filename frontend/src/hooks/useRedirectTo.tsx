import { useNavigate, useSearchParams } from "react-router";

function safePath(path: string | null, fallback: string) {
  if (
    !path ||
    !path.startsWith("/") ||
    path.startsWith("//") ||
    path.startsWith("/\\")
  ) {
    return fallback;
  }
  return path;
}

export function useRedirectTo(fallback = "/") {
  const [params] = useSearchParams();
  const navigate = useNavigate();

  const redirectTo = safePath(params.get("redirectTo"), fallback);

  return {
    redirectTo,
    goToRedirect: () => navigate(redirectTo, { replace: true }),
  };
}
