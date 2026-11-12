import { Link, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import {
  DirectionBadge,
  OrderTypeBadge,
  StatusBadge,
} from "../components/Badge";
import { Card, DetailItem, Section } from "../components/ui/card";
import { useCancelOrder } from "../features/orders/hooks";
import { OrderQueries } from "../features/orders/query";
import { getErrorMessage } from "../lib/apiError";

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "medium",
});

const priceFormatter = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 4,
});

const yieldFormatter = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 3,
});

const quantityFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 0,
});

const nominalFormatter = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const fxFormatter = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 4,
});

const percentFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 2,
});

function formatDecimal(value: string, formatter: Intl.NumberFormat): string {
  if (value === "") return "—";
  const numeric = Number(value);
  return Number.isFinite(numeric) ? formatter.format(numeric) : value;
}

function formatDateTime(value: string): string {
  if (value === "") return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : dateTimeFormatter.format(date);
}

function formatPercentage(value: string): string {
  if (value === "") return "—";
  const numeric = Number(value);
  return Number.isFinite(numeric)
    ? `${percentFormatter.format(numeric)}%`
    : value;
}

export default function OrderDetailsPage() {
  const { orderId } = useParams();
  const orderQuery = useQuery(OrderQueries.detail(orderId ?? ""));
  const cancelOrder = useCancelOrder();

  if (!orderId) {
    return <OrderMessage message="Order not found." />;
  }

  if (orderQuery.isPending) {
    return <OrderMessage message="Loading order…" />;
  }

  if (orderQuery.isError) {
    return (
      <OrderMessage
        message={getErrorMessage(orderQuery.error) ?? "Could not load order."}
      />
    );
  }

  const order = orderQuery.data;
  if (!order) {
    return <OrderMessage message="Order not found." />;
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-4">
      <Link
        to="/blotter"
        className="text-sm text-blue-400 hover:text-blue-300 hover:underline"
      >
        ← Back to blotter
      </Link>

      <header className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-semibold text-zinc-100">
            Order details
          </h1>
        </div>

        {order.status === "active" && (
          <button
            type="button"
            disabled={cancelOrder.isPending}
            onClick={() => cancelOrder.mutate(order.id)}
            className="rounded border border-red-900/70 bg-red-950/40 px-3 py-1.5 text-xs font-medium uppercase tracking-wider text-red-300 hover:border-red-700 hover:bg-red-900/50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {cancelOrder.isPending ? "Cancelling..." : "Cancel order"}
          </button>
        )}
      </header>

      <Card className="mb-0">
        <Section title="Identification">
          <DetailItem label="Order ID" mono className="sm:col-span-2">
            {order.id}
          </DetailItem>
          <DetailItem label="Security ID" mono>
            <Link
              to={`/securities/${order.security_id}`}
              className="text-blue-400 hover:text-blue-300 hover:underline"
            >
              {order.security_id}
            </Link>
          </DetailItem>
          <DetailItem label="Status">
            <StatusBadge status={order.status} />
          </DetailItem>
        </Section>

        <Section title="Execution">
          <DetailItem label="Direction">
            <DirectionBadge direction={order.direction} />
          </DetailItem>
          <DetailItem label="Order type">
            <OrderTypeBadge orderType={order.order_type} />
          </DetailItem>
          {order.order_type !== "market" && (
            <>
              <DetailItem label="Price">
                {formatDecimal(order.price, priceFormatter)}
              </DetailItem>
              <DetailItem label="Yield">
                {formatDecimal(order.yield, yieldFormatter)}
              </DetailItem>
            </>
          )}
          <DetailItem label="Quantity">
            {quantityFormatter.format(order.quantity)}
          </DetailItem>
          <DetailItem label="Nominal value">
            {formatDecimal(order.nominal_value, nominalFormatter)}
          </DetailItem>
          <DetailItem label="FX">
            {formatDecimal(order.fx, fxFormatter)}
          </DetailItem>
        </Section>

        <Section title="Allocation">
          <DetailItem label="Allocated quantity">
            {quantityFormatter.format(order.allocated_quantity)}
          </DetailItem>
          <DetailItem label="Allocation">
            {formatPercentage(order.allocation_ptc)}
          </DetailItem>
          <DetailItem label="Allocation status">
            <span className="capitalize">
              {order.allocation_status.replace(/_/g, " ")}
            </span>
          </DetailItem>
        </Section>

        <Section title="Dates">
          <DetailItem label="Created">
            {formatDateTime(order.created_at)}
          </DetailItem>
          <DetailItem label="Updated">
            {formatDateTime(order.updated_at)}
          </DetailItem>
          <DetailItem label="Expires">
            {formatDateTime(order.expires_at)}
          </DetailItem>
        </Section>
      </Card>
    </div>
  );
}

function OrderMessage({ message }: { message: string }) {
  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-4">
      <Link
        to="/blotter"
        className="text-sm text-blue-400 hover:text-blue-300 hover:underline"
      >
        ← Back to blotter
      </Link>
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/70 p-6 text-sm text-zinc-500">
        {message}
      </div>
    </div>
  );
}
