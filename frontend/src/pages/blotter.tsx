import { Button } from "../components/ui/button";
import { useCallback, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { SortingState } from "@tanstack/react-table";
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
        <h1 className="ui-title">Order Blotter</h1>
        <p className="mt-1 text-sm text-fg-muted">
          Working orders and execution history.
        </p>
      </header>

      {ordersQuery.isError ? (
        <div role="alert" className="ui-alert-danger p-4 text-sm">
          <p>
            Could not load orders:{" "}
            {ordersQuery.error instanceof Error
              ? ordersQuery.error.message
              : "unknown error"}
          </p>
          <Button
            onClick={() => void ordersQuery.refetch()}
            size="sm"
            className="mt-3"
          >
            Retry
          </Button>
        </div>
      ) : (
        <>
          <div
            role="tablist"
            aria-label="Blotter view"
            className="flex gap-1 border-b border-border"
          >
            <Button
              role="tab"
              aria-selected={isActive}
              onClick={() => setTab("active")}
              variant="tab"
              size="sm"
              className="gap-1 px-3 py-2"
            >
              Active orders{" "}
              <span className="text-fg-faint">({activeOrders.length})</span>
            </Button>
            <Button
              role="tab"
              aria-selected={!isActive}
              onClick={() => setTab("history")}
              variant="tab"
              size="sm"
              className="gap-1 px-3 py-2"
            >
              Order history{" "}
              <span className="text-fg-faint">({historyOrders.length})</span>
            </Button>
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
