import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="text-fg-muted">This page doesn't exist.</p>
      <Link to="/" className="text-accent-soft hover:underline">
        Go home
      </Link>
    </div>
  );
}
