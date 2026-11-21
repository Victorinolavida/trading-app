import { Button } from "../components/ui/button";
import { useCallback } from "react";
import { redirect, useNavigate, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { securityQueries } from "../features/securities/queries";
import { OrderBook } from "../components/OrderBook";
import { SecurityInfoCard } from "../components/SecurityInfoCard";
import { useOrdersStream } from "../features/orders/stream";
import { toast } from "sonner";
import { getErrorMessage } from "../lib/apiError";
import { CreateOrderForm } from "../components/CreateOrderForm";
import { DepthChart } from "../components/DepthChart";

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

  const asks = data?.asks ?? [];
  const bids = data?.bids ?? [];

  return (
    <div className="mx-auto flex w-full max-w-[110rem] flex-col gap-4">
      <Button
        variant="link"
        size="inline"
        className="self-start text-sm"
        onClick={handleGoBack}
      >
        ← Back
      </Button>

      {isLoading || !security ? (
        <p className="ui-panel p-4 text-sm text-fg-muted">
          {isLoading ? "Loading..." : "Security not found."}
        </p>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          <SecurityInfoCard security={security} />
          <CreateOrderForm key={security.id} securityId={security.id} />
          <div className="grid min-w-0 gap-4 lg:col-span-2">
            <OrderBook book={{ asks, bids }} />
            <DepthChart asks={asks} bids={bids} />
          </div>
        </div>
      )}
    </div>
  );
}
