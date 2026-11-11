import { useMemo, useState } from "react";
import {
  createColumnHelper,
  tableFeatures,
  useTable,
} from "@tanstack/react-table";
import { Sidebar } from "../components/Sidebar";
import {
  DirectionBadge,
  OrderTypeBadge,
  StatusBadge,
} from "../components/Badge";
import { useOrders } from "../hooks/useOrders";
// import { useSecuritiesByIds } from "../hooks/useSecurities";
import type { Order } from "../lib/api";

type Tab = "active" | "all";

const features = tableFeatures({});
const helper = createColumnHelper<typeof features, Order>();

function makeColumns(isinBySecurityId: Map<string, string>) {
  return helper.columns([
    helper.accessor("id", {
      header: "Order ID",
      cell: (ctx) => {
        const id = ctx.getValue();
        return id ? <span title={id}>{id.slice(0, 8)}…</span> : "-";
      },
    }),
    helper.accessor("security_id", {
      header: "ISIN",
      cell: (ctx) => {
        const id = ctx.getValue();
        if (!id) return "-";
        const isin = isinBySecurityId.get(id);
        return (
          <a
            className="text-blue-400 hover:underline"
            href={`/securities/${id}`}
            title={id}
          >
            {isin ?? `${id.slice(0, 8)}…`}
          </a>
        );
      },
    }),
    helper.accessor("direction", {
      header: "Direction",
      cell: (ctx) => <DirectionBadge direction={ctx.getValue()} />,
    }),
    helper.accessor("order_type", {
      header: "Type",
      cell: (ctx) => <OrderTypeBadge orderType={ctx.getValue()} />,
    }),
    helper.accessor("status", {
      header: "Status",
      cell: (ctx) => <StatusBadge status={ctx.getValue()} />,
    }),
    helper.accessor("price", {
      header: "Price",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? (
          <span style={{ textAlign: "right" }} title={String(value)}>
            {String(value)}
          </span>
        ) : (
          "-"
        );
      },
    }),
    helper.accessor("yield", {
      header: "Yield",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? (
          <span style={{ textAlign: "right" }} title={String(value)}>
            {String(value)}
          </span>
        ) : (
          "-"
        );
      },
    }),
    helper.accessor("quantity", {
      header: "Qty",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? (
          <span style={{ textAlign: "right" }} title={String(value)}>
            {String(value)}
          </span>
        ) : (
          "-"
        );
      },
    }),
    helper.accessor("allocated_quantity", {
      header: "Allocated",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? (
          <span style={{ textAlign: "right" }} title={String(value)}>
            {String(value)}
          </span>
        ) : (
          "-"
        );
      },
    }),
    helper.accessor("nominal_value", {
      header: "Nominal",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? (
          <span style={{ textAlign: "right" }} title={String(value)}>
            {String(value)}
          </span>
        ) : (
          "-"
        );
      },
    }),
    helper.accessor("created_at", {
      header: "Created",
      cell: (ctx) => {
        const value = ctx.getValue();
        return value ? new Date(value).toLocaleString() : "-";
      },
    }),
  ]);
}

const tabs: { key: Tab; label: string }[] = [
  { key: "active", label: "Active" },
  { key: "all", label: "All" },
];

export function BlotterPage() {
  const [tab, setTab] = useState<Tab>("active");

  const { data, isLoading, isError, error } = useOrders(
    tab === "active" ? { status: "active" } : {},
  );
  const orders = useMemo(() => data?.orders ?? [], [data]);

  const securityIds = useMemo(
    () => [
      ...new Set(
        orders
          .map((o) => o.security_id)
          .filter((id): id is string => Boolean(id)),
      ),
    ],
    [orders],
  );
  const securitiesQuery = useSecuritiesByIds(securityIds);
  const isinBySecurityId = useMemo(
    () =>
      new Map(
        (securitiesQuery.data?.securities ?? []).map((s) => [s.id, s.isin]),
      ),
    [securitiesQuery.data],
  );
  const columns = useMemo(
    () => makeColumns(isinBySecurityId),
    [isinBySecurityId],
  );

  const table = useTable({ features, columns, data: orders });

  return (
    <div className="flex min-h-screen overflow-x-hidden">
      <Sidebar />
      <main className="ml-64 flex-1 min-h-screen overflow-x-hidden p-8">
        <div className="mx-auto w-full max-w-6xl">
          <h1 className="mb-6 text-2xl font-semibold">Blotter</h1>

          <div className="mb-4 flex gap-2 border-b border-zinc-800">
            {tabs.map((t) => (
              <button
                key={t.key}
                type="button"
                onClick={() => setTab(t.key)}
                className={`cursor-pointer border-b-2 px-3 py-2 text-sm ${
                  tab === t.key
                    ? "border-blue-500 text-white"
                    : "border-transparent text-gray-400 hover:text-gray-200"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          {isLoading && <p className="text-gray-400">Loading orders…</p>}
          {isError && (
            <p className="text-red-400">
              {error instanceof Error ? error.message : "Failed to load orders"}
            </p>
          )}

          {!isLoading && !isError && (
            <div className="overflow-x-auto rounded-lg border border-zinc-800 bg-zinc-900">
              <table className="w-full text-sm">
                <thead>
                  {table.getHeaderGroups().map((hg) => (
                    <tr
                      key={hg.id}
                      className="border-b border-zinc-800 text-left text-gray-400"
                    >
                      {hg.headers.map((h) => (
                        <th key={h.id} className="px-3 py-2 font-medium">
                          {h.isPlaceholder ? null : (
                            <table.FlexRender header={h} />
                          )}
                        </th>
                      ))}
                    </tr>
                  ))}
                </thead>
                <tbody>
                  {table.getRowModel().rows.map((row) => (
                    <tr
                      key={row.id}
                      className="border-b border-zinc-900 last:border-b-0 hover:bg-zinc-800/50"
                    >
                      {row.getAllCells().map((cell) => (
                        <td key={cell.id} className="px-3 py-2">
                          <table.FlexRender cell={cell} />
                        </td>
                      ))}
                    </tr>
                  ))}
                  {table.getRowModel().rows.length === 0 && (
                    <tr>
                      <td
                        colSpan={columns.length}
                        className="px-3 py-6 text-center text-gray-400"
                      >
                        No orders
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
