import { useCallback, useState, type FormEvent } from "react";
import { redirect, useNavigate, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { securityQueries } from "../features/securities/queries";
import { OrderBook } from "../components/OrderBook";
import { Card, CardTitle } from "../components/ui/card";
import { SecurityInfoCard } from "../components/SecurityInfoCard";
import { useOrdersStream } from "../features/orders/stream";
import { DepthChart } from "../components/DepthChart";
import { createOrder } from "../lib/api";
import { toast } from "sonner";
import { getErrorMessage } from "../lib/apiError";

function CreateOrderForm({ securityId }: { securityId: string }) {
  const [side, setSide] = useState<"bid" | "ask">("bid");
  const [orderType, setOrderType] = useState<"limit" | "market">("limit");
  const [price, setPrice] = useState("");
  const [yieldVal, setYieldVal] = useState("");
  const [quantity, setQuantity] = useState("");
  const [creating, setCreating] = useState(false);
  const [createMsg, setCreateMsg] = useState<string | null>(null);

  const handleCreateOrder = async (e: FormEvent) => {
    e.preventDefault();
    if (!securityId) return;
    setCreating(true);
    setCreateMsg(null);
    try {
      const res = await createOrder({
        security_id: securityId,
        direction: side,
        order_type: orderType,
        price: price || "0",
        yield: yieldVal || "0",
        quantity: Number(quantity),
      });
      setCreateMsg(
        `Order created: group ${res.order_group_id}, order ${res.order_id}, status ${res.status}`,
      );
    } catch (err) {
      setCreateMsg(getErrorMessage(err) ?? "Failed to create order");
    } finally {
      setCreating(false);
    }
  };

  return (
    <Card>
      <CardTitle className="mb-3">Create Order</CardTitle>
      <form className="grid grid-cols-2 gap-3" onSubmit={handleCreateOrder}>
        <div>
          <label className="mb-1 block text-xs text-gray-400" htmlFor="side">
            Side
          </label>
          <select
            id="side"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            value={side}
            onChange={(e) => setSide(e.target.value as "bid" | "ask")}
          >
            <option value="bid">Bid</option>
            <option value="ask">Ask</option>
          </select>
        </div>
        <div>
          <label
            className="mb-1 block text-xs text-gray-400"
            htmlFor="orderType"
          >
            Type
          </label>
          <select
            id="orderType"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            value={orderType}
            onChange={(e) => setOrderType(e.target.value as "limit" | "market")}
          >
            <option value="limit">Limit</option>
            <option value="market">Market</option>
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400" htmlFor="price">
            Price
          </label>
          <input
            id="price"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="0"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400" htmlFor="yield">
            Yield
          </label>
          <input
            id="yield"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            value={yieldVal}
            onChange={(e) => setYieldVal(e.target.value)}
            placeholder="0"
          />
        </div>
        <div className="col-span-2">
          <label
            className="mb-1 block text-xs text-gray-400"
            htmlFor="quantity"
          >
            Quantity
          </label>
          <input
            id="quantity"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            required
            placeholder="e.g. 100"
          />
        </div>
        <div className="col-span-2">
          <button
            className="w-full cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:opacity-60"
            type="submit"
            disabled={creating}
          >
            {creating ? "Creating..." : "Create Order"}
          </button>
        </div>
      </form>
      {createMsg && <p className="mt-2 text-sm text-gray-300">{createMsg}</p>}
    </Card>
  );
}

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
