import { useLogout } from "../features/auth/hooks";

export function Sidebar() {
  const logout = useLogout();

  return (
    <aside className="fixed left-0 top-0 h-screen w-64 border-r border-zinc-800 bg-zinc-900 p-4">
      <div className="mb-6">
        <h2 className="text-lg font-semibold">Trading</h2>
      </div>
      <div className="mt-auto pt-4">
        <button
          className="w-full cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:cursor-default disabled:opacity-60"
          type="button"
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
        >
          {logout.isPending ? "Signing out..." : "Logout"}
        </button>
        <button
          className="w-full cursor-pointer rounded-md bg-zinc-800 px-4 py-2.5 text-base text-gray-400 hover:text-white margin-y-2"
          type="button"
          onClick={() => (window.location.hash = "#/blotter")}
        >
          Blotter
        </button>
      </div>
    </aside>
  );
}
