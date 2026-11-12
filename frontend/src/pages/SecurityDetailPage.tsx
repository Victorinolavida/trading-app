import { useCallback } from "react";
import { redirect, useNavigate, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { securityQueries } from "../features/securities/queries";
import { OrderBook } from "../components/OrderBook";
import { SecurityInfoCard } from "../components/SecurityInfoCard";
import { useOrdersStream } from "../features/orders/stream";
import { DepthChart } from "../components/DepthChart";
import { toast } from "sonner";
import { getErrorMessage } from "../lib/apiError";
import { CreateOrderForm } from "../components/CreateOrderForm";

export default function SecurityDetailPage() {
  const { securityId } = useParams();
  const navigate = useNavigate();
  const {
    data: security,
    isLoading,
    error,
  } = useQuery(securityQueries.detail({ id: securityId }));
  const { data } = useOrdersStream(securityId ?? "");
  const handleGoBack = useCallback(() => {
    navigate(-1);
  }, [navigate]);

  if (error) {
    toast.error(getErrorMessage(error));
    redirect("/");
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen overflow-x-hidden">
        <main className="flex-1 min-h-screen overflow-x-hidden p-8">
          <button
            className="mb-4 cursor-pointer text-sm text-blue-400"
            onClick={handleGoBack}
          >
            ← Back
          </button>
          <p className="text-gray-400">Loading...</p>
        </main>
      </div>
    );
  }

  if (!security) {
    return (
      <div className="flex min-h-screen overflow-x-hidden">
        <main className="flex-1 min-h-screen overflow-x-hidden p-8">
          <button
            className="mb-4 cursor-pointer text-sm text-blue-400"
            onClick={handleGoBack}
          >
            ← Back
          </button>
          <p className="text-gray-400">Security not found.</p>
        </main>
      </div>
    );
  }

  const asks = data?.asks ?? [];
  const bids = data?.bids ?? [];

  return (
    <div className="flex min-h-screen overflow-x-hidden">
      <main className="flex-1 min-h-screen overflow-x-hidden p-8">
        <button
          className="mb-4 cursor-pointer text-sm text-blue-400"
          onClick={handleGoBack}
        >
          ← Back
        </button>

        <div className="grid gap-6 lg:grid-cols-2">
          <SecurityInfoCard security={security} />

          <CreateOrderForm securityId={security.id} />

          <DepthChart asks={asks} bids={bids} height={300} />

          <OrderBook book={{ asks, bids }} />
        </div>
      </main>
    </div>
  );
}
