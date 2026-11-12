import { LoaderFunctionArgs, redirect } from "react-router";
import { queryClient } from "../../lib/queryClient";
import { ApiError } from "../../lib/http";
import { authKeys } from "./hooks";
import { authApi } from "./api";

export const whoamIQuery = {
  queryKey: authKeys.whoami,
  queryFn: authApi.whoami,
  retry: false,
};

const whoami = () => queryClient.query({ ...whoamIQuery, staleTime: "static" });

export async function requireAuth({ request }: LoaderFunctionArgs) {
  try {
    return await whoami();
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      const { pathname, search } = new URL(request.url);
      throw redirect(
        `/login?redirectTo=${encodeURIComponent(pathname + search)}`,
      );
    }
    throw e;
  }
}

export async function redirectIfAuthed() {
  try {
    await whoami();
  } catch {
    return null;
  }
  throw redirect("/");
}
