import { Button } from "./ui/button";
// components/Sidebar.tsx
import { Link, NavLink } from "react-router";
import {
  ClipboardList,
  Bell,
  LogOut,
  Search,
  X,
  type LucideIcon,
} from "lucide-react";
import { useLogout } from "../features/auth/hooks";
import { useQuery } from "@tanstack/react-query";
import {
  notificationKeys,
  notificationsApi,
} from "../features/notifications/api";

type NavItem = { to: string; label: string; icon: LucideIcon; end?: boolean };

const NAV_ITEMS: NavItem[] = [
  { to: "/", label: "Securities", icon: Search, end: true },
  { to: "/blotter", label: "Blotter", icon: ClipboardList },
  { to: "/notifications", label: "Notifications", icon: Bell },
];

type SidebarProps = {
  open: boolean;
  onClose: () => void;
};

export function Sidebar({ open, onClose }: SidebarProps) {
  const logout = useLogout();
  const unread = useQuery({
    queryKey: notificationKeys.count,
    queryFn: ({ signal }) => notificationsApi.unreadCount(signal),
    refetchInterval: 60_000,
  });

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
        <div className="mb-8 flex items-center justify-between border-b border-fg-strong px-2 pb-5">
          <div className="flex items-center gap-3">
            <Link to="/" className="flex gap-x-2 items-center">
              <span aria-hidden="true" className="h-8 w-1.5 bg-accent" />
              <h2 className="text-xl font-semibold tracking-tight text-fg-strong">
                Trading
              </h2>
            </Link>
          </div>
          <Button
            onClick={onClose}
            aria-label="Close menu"
            variant="ghost"
            size="icon"
            className="md:hidden"
          >
            <X className="h-5 w-5" />
          </Button>
        </div>

        <nav aria-label="Main" className="flex flex-1 flex-col gap-1">
          {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                [
                  "flex items-center gap-3 border-l-2 px-3 py-3 text-sm font-medium transition-colors",
                  isActive
                    ? "border-accent bg-surface text-accent"
                    : "border-transparent text-fg-muted hover:bg-surface hover:text-fg-strong",
                ].join(" ")
              }
            >
              <Icon className="h-4 w-4 shrink-0" aria-hidden />
              {label}
              {to === "/notifications" && (unread.data?.count ?? 0) > 0 && (
                <span
                  className="ml-auto min-w-6 bg-accent px-1.5 py-0.5 text-center font-mono text-xs font-semibold text-surface"
                  aria-label={`${unread.data?.count} unread notifications`}
                >
                  {(unread.data?.count ?? 0) > 99 ? "99+" : unread.data?.count}
                </span>
              )}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-border pt-4">
          <Button
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
            className="w-full enabled:hover:border-danger-border enabled:hover:bg-danger-surface enabled:hover:text-danger-soft"
          >
            <LogOut className="h-4 w-4" aria-hidden />
            {logout.isPending ? "Signing out..." : "Logout"}
          </Button>
        </div>
      </aside>
    </>
  );
}
