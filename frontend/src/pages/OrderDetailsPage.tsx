import { Link, useParams } from "react-router";

export default function OrderDetailsPage() {
  const { orderId } = useParams();

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <Link
        to="/blotter"
        className="text-sm text-blue-400 hover:text-blue-300 hover:underline"
      >
        ← Back to blotter
      </Link>

      <header>
        <h1 className="text-2xl font-semibold text-zinc-100">Order details</h1>
        <p className="mt-1 text-sm text-zinc-500">
          Order{" "}
          <span className="font-mono text-zinc-300">
            {orderId ?? "unknown"}
          </span>
        </p>
      </header>

      <div className="rounded-lg border border-zinc-800 bg-zinc-900/70 p-6 text-sm text-zinc-500">
        Order details are not implemented yet.
      </div>
    </div>
  );
}
