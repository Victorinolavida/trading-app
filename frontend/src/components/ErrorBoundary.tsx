import { useState, useEffect } from "react";

export function ErrorBoundary() {
  const [didCatch, setDidCatch] = useState(false);

  useEffect(() => {
    // Optionally log the error to an error tracking service
    if (didCatch) {
      // console.error("Error caught by ErrorBoundary");
    }
  }, [didCatch]);

  if (!didCatch) return null;

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-surface-raised p-8 rounded-lg shadow-lg max-w-md w-full text-center">
        <div className="text-6xl mb-4">⚠️</div>
        <h2 className="text-2xl font-semibold text-fg-strong mb-2">Unexpected Application Error!</h2>
        <p className="text-fg-muted mb-6">
          Something went wrong on our end. Our team has been notified, and we're working to fix it.
        </p>
        <button
          onClick={() => window.location.reload()}
          className="inline-flex items-center gap-2 px-4 py-2 border border-border rounded-md bg-primary text-primary-foreground hover:bg-primary/90"
        >
          Retry
        </button>
      </div>
    </div>
  );
}