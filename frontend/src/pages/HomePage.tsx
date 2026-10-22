import type { User } from "../lib/api";

export function HomePage({ user, onLogout }: { user: User; onLogout: () => void }) {
  return (
    <div className="w-full max-w-sm rounded-lg border border-zinc-800 bg-zinc-900 p-8">
      <h1 className="mb-5 text-2xl font-semibold">Welcome</h1>
      <dl className="mb-2 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1">
        <dt className="text-sm text-gray-400">User ID</dt>
        <dd className="text-sm break-all">{user.user_id}</dd>
        <dt className="text-sm text-gray-400">First Name</dt>
        <dd className="text-sm break-all">{user.first_name}</dd>
        <dt className="text-sm text-gray-400">Last Name</dt>
        <dd className="text-sm break-all">{user.last_name}</dd>
        <dt className="text-sm text-gray-400">Email</dt>
        <dd className="text-sm break-all">{user.email}</dd>
      </dl>
      <button
        className="mt-4 cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:cursor-default disabled:opacity-60"
        type="button"
        onClick={onLogout}
      >
        Sign out
      </button>
    </div>
  );
}

