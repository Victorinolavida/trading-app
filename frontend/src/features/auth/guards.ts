import { LoaderFunctionArgs, redirect } from "react-router";
import { queryClient } from "../../lib/queryClient";
import { ApiError } from "../../lib/http";
import { authKeys } from "./hooks";
import { authApi } from "./api";

export const whoamIQuery = {
  queryKey: authKeys.user,
  queryFn: authApi.whoami,
  retry: false,
};

const whoami = () => queryClient.query({ ...whoamIQuery, staleTime: "static" });

export async function requireAuth({ request }: LoaderFunctionArgs) {
  try {
    const user = await whoami();
    console.log("requireAuth user:", user);

    if (user) return user;
  } catch (e) {
    if (!(e instanceof ApiError && e.status === 401)) throw e;
  }
  const { pathname, search } = new URL(request.url);
  throw redirect(`/login?redirectTo=${encodeURIComponent(pathname + search)}`);
}

export async function redirectIfAuthed() {
  const user = await whoami().catch(() => null);
  console.log("redirectIfAuthed user:", user);
  if (user) throw redirect("/");
  return null;
}
