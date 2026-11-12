import { Outlet } from "react-router";

export default function AuthLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md rounded-xl p-8 shadow-lg">
        <Outlet />
      </div>
    </div>
  );
}
