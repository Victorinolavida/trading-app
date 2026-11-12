import { memo } from "react";
import {
  columnFilteringFeature,
  columnVisibilityFeature,
  constructFilterFn,
  constructSortFn,
  createColumnHelper,
  createFilteredRowModel,
  createPaginatedRowModel,
  createSortedRowModel,
  filterFn_equalsString,
  globalFilteringFeature,
  rowPaginationFeature,
  rowSortingFeature,
  sortFn_datetime,
  sortFn_text,
  tableFeatures,
} from "@tanstack/react-table";
import { toast } from "sonner";
import {
  DirectionBadge,
  OrderTypeBadge,
  StatusBadge,
} from "../../components/Badge";
import { cn } from "../../components/ui/card";
import type { AllocationStatus, Order } from "./types";
import { formatDuration, useNow } from "./useNow";

export const blotterFeatures = tableFeatures({
  rowSortingFeature,
  columnFilteringFeature,
  globalFilteringFeature,
  columnVisibilityFeature,
  rowPaginationFeature,
  sortedRowModel: createSortedRowModel(),
  filteredRowModel: createFilteredRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
  columnMeta: {} as { align?: "left" | "right" },
});

const helper = createColumnHelper<typeof blotterFeatures, Order>();

const numericSort = constructSortFn({
  sort: (left, right) => {
    const a = Number(left);
    const b = Number(right);
    if (Number.isNaN(a) && Number.isNaN(b)) return 0;
    if (Number.isNaN(a)) return 1;
    if (Number.isNaN(b)) return -1;
    return a - b;
  },
});

const orderSearchFilter = constructFilterFn<typeof blotterFeatures, Order>({
  filter: (_value, filterValue, row) => {
    const query = String(filterValue ?? "")
      .trim()
      .toLowerCase();
    if (!query) return true;

    return (
      row.original.id.toLowerCase().includes(query) ||
      row.original.security_id.toLowerCase().includes(query)
    );
  },
  autoRemove: (value) => typeof value !== "string" || value.trim() === "",
});

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: "short",
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

const allocationStatusClasses: Record<AllocationStatus, string> = {
  none: "text-zinc-500",
  partial: "text-amber-400",
  allocated: "text-green-400",
};

const allocationBarClasses: Record<AllocationStatus, string> = {
  none: "bg-zinc-600",
  partial: "bg-amber-500",
  allocated: "bg-green-500",
};

function formatDecimal(
  value: string | number | null | undefined,
  formatter: Intl.NumberFormat,
): string {
  if (value === null || value === undefined || value === "") return "—";
  const numeric = Number(value);
  return Number.isFinite(numeric) ? formatter.format(numeric) : String(value);
}

function TimestampCell({ value }: { value: string }) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return <span>—</span>;

  return (
    <time
      dateTime={value}
      title={value}
      className="whitespace-nowrap text-zinc-300"
    >
      {dateTimeFormatter.format(date)}
    </time>
  );
}

async function copyText(value: string, label: string) {
  try {
    await navigator.clipboard.writeText(value);
    toast.success(`${label} copied`);
  } catch {
    toast.error(`Could not copy ${label.toLowerCase()}`);
  }
}

function CopyableCell({ value, label }: { value: string; label: string }) {
  return (
    <button
      type="button"
      title={value}
      aria-label={`Copy ${label} ${value}`}
      onClick={() => void copyText(value, label)}
      className="group inline-flex items-center gap-1 rounded font-mono text-xs text-zinc-300 hover:text-white focus:outline-none focus-visible:ring-1 focus-visible:ring-blue-500"
    >
      <span>{value.slice(0, 8)}</span>
      <span className="text-[9px] uppercase text-zinc-600 opacity-0 transition-opacity group-hover:opacity-100">
        copy
      </span>
    </button>
  );
}

function NumericCell({
  value,
  formatter,
}: {
  value: string | number | null | undefined;
  formatter: Intl.NumberFormat;
}) {
  return (
    <span
      className="tabular-nums"
      title={value === null || value === undefined ? undefined : String(value)}
    >
      {formatDecimal(value, formatter)}
    </span>
  );
}

function AllocationCell({ order }: { order: Order }) {
  const percentage = Math.min(
    100,
    Math.max(0, Number(order.allocation_ptc) || 0),
  );

  return (
    <div className="flex min-w-36 flex-col gap-1">
      <div className="flex items-center justify-between gap-2 text-xs tabular-nums">
        <span className="text-zinc-200">
          {quantityFormatter.format(order.allocated_quantity)} /{" "}
          {quantityFormatter.format(order.quantity)}
        </span>
        <span className="text-zinc-400">
          {percentFormatter.format(percentage)}%
        </span>
      </div>
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-800">
        <div
          className={cn(
            "h-full rounded-full transition-[width] duration-300",
            allocationBarClasses[order.allocation_status],
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <span
        className={cn(
          "text-[9px] uppercase tracking-wider",
          allocationStatusClasses[order.allocation_status],
        )}
      >
        {order.allocation_status}
      </span>
    </div>
  );
}

const ExpiresInCell = memo(function ExpiresInCell({
  expiresAt,
}: {
  expiresAt: string;
}) {
  const now = useNow();
  const expiresAtMs = Date.parse(expiresAt);

  if (!Number.isFinite(expiresAtMs)) return <span>—</span>;

  const remaining = expiresAtMs - now;
  if (remaining <= 0) {
    return (
      <span className="font-medium text-red-400" title={expiresAt}>
        Expired
      </span>
    );
  }

  return (
    <span className="tabular-nums text-zinc-300" title={expiresAt}>
      {formatDuration(remaining)}
    </span>
  );
});

const baseColumns = [
  helper.accessor("created_at", {
    id: "created_at",
    header: "Created",
    cell: ({ getValue }) => <TimestampCell value={getValue()} />,
    sortFn: sortFn_datetime,
    enableGlobalFilter: false,
  }),
  helper.accessor("id", {
    id: "id",
    header: "Order ID",
    cell: ({ getValue }) => (
      <CopyableCell value={getValue()} label="Order ID" />
    ),
    sortFn: sortFn_text,
    enableGlobalFilter: true,
  }),
  helper.accessor("security_id", {
    id: "security_id",
    header: "Security",
    cell: ({ getValue }) => (
      <CopyableCell value={getValue()} label="Security ID" />
    ),
    sortFn: sortFn_text,
    enableGlobalFilter: false,
  }),
  helper.accessor("direction", {
    id: "direction",
    header: "Side",
    cell: ({ getValue }) => <DirectionBadge direction={getValue()} />,
    sortFn: sortFn_text,
    filterFn: filterFn_equalsString,
    enableGlobalFilter: false,
  }),
  helper.accessor("order_type", {
    id: "order_type",
    header: "Type",
    cell: ({ getValue }) => <OrderTypeBadge orderType={getValue()} />,
    sortFn: sortFn_text,
    filterFn: filterFn_equalsString,
    enableGlobalFilter: false,
  }),
  helper.accessor("price", {
    id: "price",
    header: "Price",
    meta: { align: "right" },
    cell: ({ getValue }) => (
      <NumericCell value={getValue()} formatter={priceFormatter} />
    ),
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("yield", {
    id: "yield",
    header: "Yield",
    meta: { align: "right" },
    cell: ({ getValue }) => (
      <NumericCell value={getValue()} formatter={yieldFormatter} />
    ),
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("quantity", {
    id: "quantity",
    header: "Quantity",
    meta: { align: "right" },
    cell: ({ getValue }) => (
      <NumericCell value={getValue()} formatter={quantityFormatter} />
    ),
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("nominal_value", {
    id: "nominal_value",
    header: "Nominal",
    meta: { align: "right" },
    cell: ({ getValue }) => (
      <NumericCell value={getValue()} formatter={nominalFormatter} />
    ),
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("allocation_ptc", {
    id: "allocation",
    header: "Allocated",
    meta: { align: "right" },
    cell: ({ row }) => <AllocationCell order={row.original} />,
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("fx", {
    id: "fx",
    header: "FX",
    meta: { align: "right" },
    cell: ({ getValue }) => (
      <NumericCell value={getValue()} formatter={fxFormatter} />
    ),
    sortFn: numericSort,
    enableGlobalFilter: false,
  }),
  helper.accessor("status", {
    id: "status",
    header: "Status",
    cell: ({ getValue }) => <StatusBadge status={getValue()} />,
    sortFn: sortFn_text,
    filterFn: filterFn_equalsString,
    enableGlobalFilter: false,
  }),
];

export const activeColumns = [
  ...baseColumns,
  helper.accessor("expires_at", {
    id: "expires_in",
    header: "Expires in",
    meta: { align: "right" },
    cell: ({ getValue }) => <ExpiresInCell expiresAt={getValue()} />,
    sortFn: sortFn_datetime,
    enableGlobalFilter: false,
  }),
];

export const historyColumns = [
  ...baseColumns,
  helper.accessor("updated_at", {
    id: "updated_at",
    header: "Updated",
    cell: ({ getValue }) => <TimestampCell value={getValue()} />,
    sortFn: sortFn_datetime,
    enableGlobalFilter: false,
  }),
];

export const orderGlobalFilter = orderSearchFilter;
