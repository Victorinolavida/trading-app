import { useEffect, useState } from "react";
import { Outlet, useLocation } from "react-router";
import { Menu } from "lucide-react";
import { Sidebar } from "./components/Sidebar";

export default function App() {
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
        <button
          type="button"
          onClick={() => setOpen(true)}
          aria-label="Open menu"
          aria-expanded={open}
          className="rounded-md p-1.5 text-fg-soft hover:bg-surface-hover hover:text-fg-strong"
        >
          <Menu className="h-5 w-5" />
        </button>
        <img src="/favicon.svg" alt="" className="h-6 w-6" />
        <span className="font-semibold text-fg-strong">Trading</span>
      </header>

      <Sidebar open={open} onClose={() => setOpen(false)} />

      {/* min-w-0 lets wide children (tables) scroll inside their own container */}
      <main className="min-w-0 p-4 md:ml-64 md:p-8">
        <Outlet />
      </main>
    </div>
  );
}
