import { createBrowserRouter, redirect } from "react-router";
import AuthLayout from "../layout/auth";
import { LoginPage } from "../pages/LoginPage";
import { RegisterPage } from "../pages/RegisterPage";
import App from "../App";
import HomePage from "../pages/HomePage";
import { redirectIfAuthed, requireAuth } from "../features/auth/guards";
import NotFoundPage from "../pages/NotFound";
import SecurityDetailPage from "../pages/SecurityDetailPage";
import BlotterRoute from "../pages/blotter";
import { toast } from "sonner";
import { queryClient } from "./queryClient";
import { securityQueries } from "../features/securities/queries";

export function navigate(path: string) {
  window.history.pushState({}, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export const router = createBrowserRouter([
  {
    path: "/",
    Component: App,
    id: "root",
    loader: requireAuth,
    children: [
      {
        index: true,
        Component: HomePage,
      },
      {
        path: "/blotter",
        Component: BlotterRoute,
      },
      {
        Component: SecurityDetailPage,
        path: "/securities/:securityId",
        loader: async ({ params }) => {
          const securityId = params.securityId;
          if (!securityId || securityId.length <= 0) {
            toast.error("security not found");
            throw redirect("/");
          }
          const security = await queryClient.query({
            ...securityQueries.detail({ id: securityId }),
            staleTime: "static",
          });
          if (!security) {
            toast.error("security not found");
            throw redirect("/");
          }
          return security;
        },
      },
    ],
  },
  {
    Component: AuthLayout,
    loader: redirectIfAuthed,
    children: [
      {
        path: "/login",
        Component: LoginPage,
      },
      {
        path: "/register",
        Component: RegisterPage,
      },
    ],
  },
  { path: "*", Component: NotFoundPage },
]);
