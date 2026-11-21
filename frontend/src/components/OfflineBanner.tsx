import { OctagonAlert } from "lucide-react";
import { useOnlineStatus } from "../hooks/useOnlineStatus";

export function OfflineBanner() {
  const isOnline = useOnlineStatus();
  if (isOnline) return null;

  return (
    <div
      role="alert"
      className="fixed top-0 left-0 right-0 z-50 bg-danger-surface/80 border-danger-border/60 border-y p-4 text-danger-soft text-sm"
    >
      <div className="flex items-center gap-3">
        <OctagonAlert
          className="size-5 shrink-0 text-danger-soft"
          aria-hidden="true"
        />
        <span>You are offline. Some features may not work properly.</span>
      </div>
    </div>
  );
}
