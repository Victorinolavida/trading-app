import { useEffect } from "react";
import { Link } from "react-router";
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { ArrowRight, BellOff } from "lucide-react";
import { Button } from "../components/ui/button";
import {
  notificationKeys,
  notificationsApi,
  type Notification,
} from "../features/notifications/api";

const titles: Record<string, string> = {
  "order.created": "Order created",
  "order.canceled": "Order canceled",
  "order.expired": "Order expired",
  "order.executed": "Order executed",
  "order.discarded": "Order discarded",
};

function NotificationRow({
  item,
  onRead,
  marking,
}: {
  item: Notification;
  onRead: (id: string) => void;
  marking: boolean;
}) {
  const order = item.payload?.order;
  const date = new Date(item.created_at);
  const time = Number.isNaN(date.getTime())
    ? item.created_at
    : date.toLocaleString();

  return (
    <li className="border-b border-border-subtle px-5 py-5 last:border-b-0 sm:px-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span
            className={`h-2 w-2 shrink-0 ${item.read_at ? "bg-border-strong" : "bg-accent"}`}
            aria-hidden="true"
          />
          <h2 className="font-semibold text-fg-strong">
            {titles[item.event_type] ?? "Order notification"}
          </h2>
        </div>
        <time
          dateTime={item.created_at}
          className="font-mono text-xs text-fg-muted"
        >
          {time}
        </time>
      </div>
      <div className="ml-5 mt-2 flex flex-wrap gap-x-5 gap-y-1 text-sm text-fg-muted">
        {order?.security_id && (
          <span>
            Security:{" "}
            <Link
              className="ui-link break-all"
              to={`/securities/${encodeURIComponent(order.security_id)}`}
            >
              {order.security_id}
            </Link>
          </span>
        )}
        {order?.quantity != null && (
          <span>
            Quantity:{" "}
            <span className="font-mono text-fg">
              {order.quantity.toLocaleString()}
            </span>
          </span>
        )}
        {order?.price && (
          <span>
            Price: <span className="font-mono text-fg">{order.price}</span>
          </span>
        )}
        {item.payload?.reason === "internal_crossing" && (
          <span>Reason: internal crossing</span>
        )}
      </div>
      <div className="ml-5 mt-3 flex flex-wrap items-center gap-4 text-sm">
        {order?.id && (
          <Link
            className="ui-link inline-flex items-center gap-1"
            to={`/orders/${encodeURIComponent(order.id)}`}
          >
            View order <ArrowRight className="h-3.5 w-3.5" aria-hidden="true" />
          </Link>
        )}
        {!item.read_at && (
          <Button
            variant="link"
            size="inline"
            disabled={marking}
            onClick={() => onRead(item.id)}
          >
            Mark as read
          </Button>
        )}
      </div>
    </li>
  );
}

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const list = useInfiniteQuery({
    queryKey: notificationKeys.list,
    queryFn: ({ pageParam, signal }) =>
      notificationsApi.list(pageParam, signal),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.next_cursor || undefined,
  });
  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: notificationKeys.all });
  const markAll = useMutation({
    mutationFn: notificationsApi.markAllRead,
    onSettled: refresh,
  });
  const markOne = useMutation({
    mutationFn: notificationsApi.markRead,
    onSettled: refresh,
  });
  const markAllMutate = markAll.mutate;

  // Viewing the inbox marks all saved notifications read, including those not
  // loaded on the current page. Live arrivals are handled by the SSE hook.
  useEffect(() => {
    markAllMutate();
    // Also reconcile missed live signals while the inbox remains open.
    const interval = window.setInterval(() => markAllMutate(), 60_000);
    return () => window.clearInterval(interval);
  }, [markAllMutate]);

  const notifications =
    list.data?.pages.flatMap((page) => page.notifications ?? []) ?? [];
  const error = list.error ?? markAll.error ?? markOne.error;

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <header className="page-heading flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="ui-title">Notifications</h1>
          <p className="mt-1 text-sm text-fg-muted">
            Updates about your orders.
          </p>
        </div>
        <Button
          size="sm"
          onClick={() => markAll.mutate()}
          disabled={markAll.isPending}
        >
          Mark all as read
        </Button>
      </header>

      {error && (
        <div role="alert" className="ui-alert-danger p-4 text-sm">
          Could not update notifications.{" "}
          <Button
            variant="link"
            size="inline"
            onClick={() => {
              void list.refetch();
              markAll.mutate();
            }}
          >
            Retry
          </Button>
        </div>
      )}
      <section
        className="ui-panel"
        aria-label="Notification inbox"
        aria-busy={list.isPending}
      >
        {list.isPending ? (
          <p className="p-6 text-sm text-fg-muted">Loading notifications...</p>
        ) : notifications.length === 0 && !list.isError ? (
          <div className="flex flex-col items-center gap-3 px-6 py-16 text-center text-fg-muted">
            <BellOff className="h-6 w-6" aria-hidden="true" />
            <p className="font-medium text-fg">No notifications yet</p>
            <p className="text-sm">Order updates will appear here.</p>
          </div>
        ) : (
          <ul>
            {notifications.map((item) => (
              <NotificationRow
                key={item.id}
                item={item}
                onRead={(id) => markOne.mutate(id)}
                marking={markOne.isPending}
              />
            ))}
          </ul>
        )}
        {list.hasNextPage && (
          <div className="border-t border-border px-6 py-4">
            <Button
              size="sm"
              disabled={list.isFetchingNextPage}
              onClick={() => void list.fetchNextPage()}
            >
              {list.isFetchingNextPage ? "Loading..." : "Load more"}
            </Button>
          </div>
        )}
      </section>
    </div>
  );
}
