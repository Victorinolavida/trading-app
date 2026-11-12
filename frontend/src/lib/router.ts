import { createBrowserRouter } from "react-router";
import AuthLayout from "../layout/auth";
import { LoginPage } from "../pages/LoginPage";
import { RegisterPage } from "../pages/RegisterPage";
import App from "../App";
import HomePage from "../pages/HomePage";
import { redirectIfAuthed, requireAuth } from "../features/auth/guards";

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
]);
