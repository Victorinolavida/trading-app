import { Button } from "./ui/button";
import { useEffect } from "react";
import { useRouteError, isRouteErrorResponse } from "react-router";

export default function ErrorBoundary() {
  const error = useRouteError();

  useEffect(() => {
    // send to your error tracker here (Sentry, etc.)
    console.error("Route error:", error);
  }, [error]);

  const isResponse = isRouteErrorResponse(error);

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="ui-panel w-full max-w-md p-6 text-center">
        <div className="text-6xl mb-4">⚠️</div>
        <h2 className="ui-title mb-2">
          {isResponse
            ? `${error.status} ${error.statusText}`
            : "Unexpected Application Error!"}
        </h2>
        <p className="text-fg-muted mb-6">
          Something went wrong on our end. Our team has been notified, and we're
          working to fix it.
        </p>
        <Button onClick={() => window.location.reload()} variant="primary">
          Retry
        </Button>
      </div>
    </div>
  );
}
