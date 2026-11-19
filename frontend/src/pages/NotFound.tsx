import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4">
      <h1 className="ui-title">404</h1>
      <p className="text-fg-muted">This page doesn't exist.</p>
      <Link to="/" className="ui-link">
        Go home
      </Link>
    </div>
  );
}
