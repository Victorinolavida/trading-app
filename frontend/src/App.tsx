import { Button } from "./components/ui/button";
import { useEffect, useState } from "react";
import { Outlet, useLocation } from "react-router";
import { Menu } from "lucide-react";
import { Sidebar } from "./components/Sidebar";
import { OfflineBanner } from "./components/OfflineBanner";
import { useNotificationStream } from "./features/notifications/useNotificationStream";

export default function App() {
  useNotificationStream();
  const [open, setOpen] = useState(false);
  const { pathname } = useLocation();

  // Close the drawer after navigating
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  return (
    <div className="min-h-dvh bg-surface text-fg-strong">
      {/* Mobile top bar */}
      <header className="sticky top-0 z-30 flex items-center gap-3 border-b border-border bg-surface-raised px-4 py-3 md:hidden">
        <Button
          onClick={() => setOpen(true)}
          aria-label="Open menu"
          aria-expanded={open}
          variant="ghost"
          size="icon"
        >
          <Menu className="h-5 w-5" />
        </Button>
        <span aria-hidden="true" className="h-6 w-1.5 bg-accent" />
        <span className="font-semibold text-fg-strong">Trading</span>
      </header>

      <OfflineBanner />

      <Sidebar open={open} onClose={() => setOpen(false)} />

      {/* min-w-0 lets wide children (tables) scroll inside their own container */}
      <main className="min-w-0 p-4 md:ml-64 md:p-10">
        <Outlet />
      </main>
    </div>
  );
}
