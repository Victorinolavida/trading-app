import type { User } from "../lib/api";
import { SecuritySearch } from "../components/SecuritySearch";
import { Sidebar } from "../components/Sidebar";

export function HomePage({ user, onLogout }: { user: User; onLogout: () => void }) {
  return (
    <div className="flex min-h-screen overflow-x-hidden">
      <Sidebar />
      <main className="ml-64 flex-1 min-h-screen overflow-x-hidden p-8">
        <div className="mx-auto w-full max-w-md">
          <h1 className="mb-6 text-2xl font-semibold">Securities</h1>
          <SecuritySearch />
        </div>
      </main>
    </div>
  );
}

