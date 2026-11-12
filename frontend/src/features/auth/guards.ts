import { LoaderFunctionArgs, redirect } from "react-router";
import { queryClient } from "../../lib/queryClient";
import { ApiError } from "../../lib/http";
import { authQueries } from "./queries";

const whoami = () =>
  queryClient.query({ ...authQueries.whoami, staleTime: "static" });

export async function requireAuth({ request }: LoaderFunctionArgs) {
  try {
    const user = await whoami();
    if (user) return user;
  } catch (e) {
    if (!(e instanceof ApiError && e.status === 401)) throw e;
  }
  const { pathname, search } = new URL(request.url);
  throw redirect(`/login?redirectTo=${encodeURIComponent(pathname + search)}`);
}

export async function redirectIfAuthed() {
  const user = await whoami().catch(() => null);
  if (user) throw redirect("/");
  return null;
}
