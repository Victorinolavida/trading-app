import { useState } from "react";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { HomePage } from "./pages/HomePage";
import { useLogout, useSession } from "./hooks/useAuth";

type AuthView = "login" | "register";

export default function App() {
  const [view, setView] = useState<AuthView>("login");
  const { data: user, isLoading, isError } = useSession();
  const logout = useLogout();

  if (isLoading) {
    return <div className="flex min-h-screen items-center justify-center p-4">Loading session...</div>;
  }

  if (isError) {
    return <div className="flex min-h-screen items-center justify-center p-4 text-sm text-red-400">Could not reach the gateway.</div>;
  }

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        {view === "login" ? (
          <LoginPage onSwitch={() => setView("register")} />
        ) : (
          <RegisterPage onSwitch={() => setView("login")} />
        )}
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <HomePage user={user} onLogout={() => logout.mutate()} />
    </div>
  );
}

