import { useOnlineStatus } from "../lib/useOnlineStatus";
import { useState } from "react";

export function OfflineBanner() {
  const isOnline = useOnlineStatus();
  const [showBanner, setShowBanner] = useState(false);

  useEffect(() => {
    setShowBanner(!isOnline);
  }, [isOnline]);

  if (isOnline) return null;

  return (
    <div
      role="alert"
      className="fixed top-0 left-0 right-0 z-50 bg-danger-surface/80 border-danger-border/60 border-y p-4 text-danger-soft text-sm"
    >
      <div className="flex items-center gap-3">
        <span className="iconify" icon="mdi:alert-octagon" className="text-danger-soft" /></span>
        <span>You are offline. Some features may not work properly.</span>
      </div>
    </div>
  );
}