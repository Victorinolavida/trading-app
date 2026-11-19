// components/Sidebar.tsx
import { Link, NavLink } from "react-router";
import {
  ClipboardList,
  LogOut,
  Search,
  X,
  type LucideIcon,
} from "lucide-react";
import { useLogout } from "../features/auth/hooks";

type NavItem = { to: string; label: string; icon: LucideIcon; end?: boolean };

const NAV_ITEMS: NavItem[] = [
  { to: "/", label: "Securities", icon: Search, end: true },
  { to: "/blotter", label: "Blotter", icon: ClipboardList },
];

type SidebarProps = {
  open: boolean;
  onClose: () => void;
};

export function Sidebar({ open, onClose }: SidebarProps) {
  const logout = useLogout();

  return (
    <>
      {/* Backdrop (mobile only) */}
      <div
        onClick={onClose}
        aria-hidden
        className={[
          "fixed inset-0 z-40 bg-overlay/60 transition-opacity md:hidden",
          open ? "opacity-100" : "pointer-events-none opacity-0",
        ].join(" ")}
      />

      <aside
        className={[
          "fixed left-0 top-0 z-50 flex h-dvh w-64 flex-col border-r border-border bg-surface-raised p-4",
          "transition-transform duration-200 ease-out",
          "md:translate-x-0",
          open ? "translate-x-0" : "-translate-x-full",
        ].join(" ")}
      >
        <div className="mb-6 flex items-center justify-between px-2">
          <div className="flex items-center gap-3">
            <Link to="/" className="flex gap-x-2 items-center">
              <img src="/favicon.svg" alt="" className="h-8 w-8" />
              <h2 className="text-lg font-semibold text-fg-strong">Trading</h2>
            </Link>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close menu"
            className="rounded-md p-1.5 text-fg-muted hover:bg-surface-hover hover:text-fg-strong md:hidden"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <nav aria-label="Main" className="flex flex-1 flex-col gap-1">
          {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                [
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-surface-hover text-fg-strong"
                    : "text-fg-muted hover:bg-surface-hover/60 hover:text-fg-strong",
                ].join(" ")
              }
            >
              <Icon className="h-4 w-4 shrink-0" aria-hidden />
              {label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-border pt-4">
          <button
            type="button"
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
            className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-md bg-surface-hover px-4 py-2 text-sm text-fg-soft transition-colors hover:bg-danger-strong hover:text-fg-strong disabled:cursor-default disabled:opacity-60"
          >
            <LogOut className="h-4 w-4" aria-hidden />
            {logout.isPending ? "Signing out..." : "Logout"}
          </button>
        </div>
      </aside>
    </>
  );
}
