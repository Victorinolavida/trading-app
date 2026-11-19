import { useCallback, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
import { cn } from "../components/ui/card";
import { activeColumns, historyColumns } from "../features/orders/columns";
import { HISTORY_PAGE_SIZE } from "../features/orders/constants";
import { OrdersTable } from "../features/orders/OrdersTable";
import { orderKeys, OrderQueries } from "../features/orders/query";
import type { Order } from "../features/orders/types";
import { useOrdersStream } from "../features/orders/useOrdersStream";

const DEFAULT_SORTING: SortingState = [{ id: "created_at", desc: true }];
const EMPTY_ORDERS: Order[] = [];

type BlotterTab = "active" | "history";

export default function BlotterRoute() {
  const queryClient = useQueryClient();
  const ordersQuery = useQuery(OrderQueries.list());
  const [tab, setTab] = useState<BlotterTab>("active");

  const upsertOrder = useCallback(
    (incoming: Order) => {
      queryClient.setQueryData<Order[]>(orderKeys.list(), (current) => {
        const orders = current ?? EMPTY_ORDERS;
        const index = orders.findIndex((order) => order.id === incoming.id);
        if (index === -1) return [incoming, ...orders];

        const next = orders.slice();
        next[index] = incoming;
        return next;
      });
    },
    [queryClient],
  );

  const replaceOrders = useCallback(
    (orders: Order[]) => {
      queryClient.setQueryData<Order[]>(orderKeys.list(), orders);
    },
    [queryClient],
  );

  useOrdersStream({ onUpsert: upsertOrder, onSnapshot: replaceOrders });

  const orders = ordersQuery.data ?? EMPTY_ORDERS;

  const { activeOrders, historyOrders } = useMemo(() => {
    const active: Order[] = [];
    const history: Order[] = [];

    for (const order of orders) {
      if (order.status === "active") {
        active.push(order);
      } else {
        history.push(order);
      }
    }

    return { activeOrders: active, historyOrders: history };
  }, [orders]);

  const isActive = tab === "active";
  const tableData = isActive ? activeOrders : historyOrders;
  const tableColumns = isActive ? activeColumns : historyColumns;
  const getRowExpiry = isActive
    ? (order: Order) => order.expires_at
    : undefined;

  return (
    <div className="mx-auto flex w-full max-w-[110rem] flex-col gap-6">
      <header>
        <h1 className="text-2xl font-semibold text-fg-strong">Order Blotter</h1>
        <p className="mt-1 text-sm text-fg-subtle">
          Working orders and execution history.
        </p>
      </header>

      {ordersQuery.isError ? (
        <div
          role="alert"
          className="rounded-lg border border-danger-border/60 bg-danger-surface/30 p-4 text-sm text-danger-soft"
        >
          <p>
            Could not load orders:{" "}
            {ordersQuery.error instanceof Error
              ? ordersQuery.error.message
              : "unknown error"}
          </p>
          <button
            type="button"
            onClick={() => void ordersQuery.refetch()}
            className="mt-3 rounded border border-danger-border px-2 py-1 text-xs text-danger-soft hover:bg-danger-surface/40"
          >
            Retry
          </button>
        </div>
      ) : (
        <>
          <div
            role="tablist"
            aria-label="Blotter view"
            className="flex gap-1 border-b border-border"
          >
            <button
              type="button"
              role="tab"
              aria-selected={isActive}
              onClick={() => setTab("active")}
              className={cn(
                "-mb-px border-b-2 px-3 py-2 text-xs font-medium uppercase tracking-wider transition-colors",
                isActive
                  ? "border-accent text-fg-strong"
                  : "border-transparent text-fg-subtle hover:text-fg-soft",
              )}
            >
              Active orders{" "}
              <span className="text-fg-faint">({activeOrders.length})</span>
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={!isActive}
              onClick={() => setTab("history")}
              className={cn(
                "-mb-px border-b-2 px-3 py-2 text-xs font-medium uppercase tracking-wider transition-colors",
                !isActive
                  ? "border-accent text-fg-strong"
                  : "border-transparent text-fg-subtle hover:text-fg-soft",
              )}
            >
              Order history{" "}
              <span className="text-fg-faint">({historyOrders.length})</span>
            </button>
          </div>

          <section className="flex flex-col gap-3">
            <OrdersTable
              key={tab}
              data={tableData}
              columns={tableColumns}
              isLoading={ordersQuery.isPending}
              initialSorting={DEFAULT_SORTING}
              showStatusFilter={!isActive}
              showPagination={!isActive}
              pageSize={HISTORY_PAGE_SIZE}
              getRowId={(order) => order.id}
              getRowExpiry={getRowExpiry}
              emptyMessage={
                isActive ? "No active orders" : "No historical orders"
              }
            />
          </section>
        </>
      )}
    </div>
  );
}
