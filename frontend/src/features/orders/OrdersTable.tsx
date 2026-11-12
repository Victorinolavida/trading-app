import {
  FlexRender,
  useTable,
  type Column,
  type ColumnHelper,
  type Row,
  type RowData,
  type SortingState,
  type ReactTable,
} from "@tanstack/react-table";
import { cn } from "../../components/ui/card";
import { orderTableFeatures, orderGlobalFilter } from "./columns";
import {
  HISTORY_PAGE_SIZE,
  HISTORY_STATUSES,
  ORDER_DIRECTIONS,
  ORDER_TYPES,
} from "./constants";
import { useIsExpired } from "./useNow";

type OrdersTableInstance<TData extends RowData> = ReactTable<
  typeof orderTableFeatures,
  TData
>;

// TanStack's column helper returns heterogeneous column defs (one TValue per
// accessor). This alias preserves that exact contract without repeating it.
type OrdersTableColumns<TData extends RowData> = ReturnType<
  ColumnHelper<typeof orderTableFeatures, TData>["columns"]
>;

interface OrdersTableProps<TData extends RowData> {
  data: readonly TData[];
  columns: OrdersTableColumns<TData>;
  isLoading?: boolean;
  emptyMessage?: string;
  initialSorting?: SortingState;
  showSearch?: boolean;
  showDirectionFilter?: boolean;
  showTypeFilter?: boolean;
  showStatusFilter?: boolean;
  showColumnVisibility?: boolean;
  showPagination?: boolean;
  pageSize?: number;
  getRowId?: (row: TData) => string;
  getRowExpiry?: (row: TData) => string | null | undefined;
}

function readColumnFilter<TData extends RowData>(
  table: OrdersTableInstance<TData>,
  columnId: string,
): string {
  const value = table.getColumn(columnId)?.getFilterValue();
  return typeof value === "string" ? value : "";
}

function writeColumnFilter<TData extends RowData>(
  table: OrdersTableInstance<TData>,
  columnId: string,
  value: string,
) {
  table.getColumn(columnId)?.setFilterValue(value === "" ? undefined : value);
}

function alignClass(align: "left" | "right" | undefined): string {
  return align === "right" ? "text-right" : "text-left";
}

function columnLabel<TData extends RowData>(
  column: Column<typeof orderTableFeatures, TData, unknown>,
): string {
  const header = column.columnDef.header;
  return typeof header === "string" ? header : column.id;
}

function FilterSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: readonly string[];
  onChange: (value: string) => void;
}) {
  return (
    <label className="inline-flex items-center gap-1.5 text-[11px] text-zinc-500">
      <span>{label}</span>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-7 rounded border border-zinc-700 bg-zinc-900 px-1.5 text-xs text-zinc-200 outline-none focus:border-blue-500"
      >
        <option value="">All</option>
        {options.map((option) => (
          <option key={option} value={option}>
            {option.replace(/_/g, " ")}
          </option>
        ))}
      </select>
    </label>
  );
}

function ColumnVisibilityMenu<TData extends RowData>({
  table,
}: {
  table: OrdersTableInstance<TData>;
}) {
  return (
    <details className="relative">
      <summary className="list-none cursor-pointer select-none rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-800 [&::-webkit-details-marker]:hidden">
        Columns
      </summary>
      <div className="absolute right-0 z-20 mt-1 max-h-72 w-48 overflow-auto rounded-md border border-zinc-700 bg-zinc-900 p-2 shadow-xl">
        <label className="flex items-center gap-2 border-b border-zinc-800 px-1 py-1.5 text-xs text-zinc-300">
          <input
            type="checkbox"
            checked={table.getIsAllColumnsVisible()}
            onChange={table.getToggleAllColumnsVisibilityHandler()}
          />
          All columns
        </label>
        {table.getAllLeafColumns().map((column) => (
          <label
            key={column.id}
            className="flex items-center gap-2 px-1 py-1 text-xs text-zinc-300"
          >
            <input
              type="checkbox"
              checked={column.getIsVisible()}
              onChange={column.getToggleVisibilityHandler()}
              disabled={!column.getCanHide()}
            />
            {columnLabel(column)}
          </label>
        ))}
      </div>
    </details>
  );
}

function TableToolbar<TData extends RowData>({
  table,
  showSearch,
  showDirectionFilter,
  showTypeFilter,
  showStatusFilter,
  showColumnVisibility,
}: {
  table: OrdersTableInstance<TData>;
  showSearch: boolean;
  showDirectionFilter: boolean;
  showTypeFilter: boolean;
  showStatusFilter: boolean;
  showColumnVisibility: boolean;
}) {
  const globalFilter =
    typeof table.state.globalFilter === "string"
      ? table.state.globalFilter
      : "";

  return (
    <div className="flex flex-wrap items-center gap-2">
      {showSearch && (
        <input
          type="search"
          value={globalFilter}
          onChange={(event) => table.setGlobalFilter(event.target.value)}
          placeholder="Search order ID or security ID"
          className="h-7 min-w-56 flex-1 rounded border border-zinc-700 bg-zinc-900 px-2 text-xs text-zinc-200 placeholder:text-zinc-600 outline-none focus:border-blue-500 sm:max-w-xs"
        />
      )}

      {showDirectionFilter && (
        <FilterSelect
          label="Side"
          value={readColumnFilter(table, "direction")}
          options={ORDER_DIRECTIONS}
          onChange={(value) => writeColumnFilter(table, "direction", value)}
        />
      )}

      {showTypeFilter && (
        <FilterSelect
          label="Type"
          value={readColumnFilter(table, "order_type")}
          options={ORDER_TYPES}
          onChange={(value) => writeColumnFilter(table, "order_type", value)}
        />
      )}

      {showStatusFilter && (
        <FilterSelect
          label="Status"
          value={readColumnFilter(table, "status")}
          options={HISTORY_STATUSES}
          onChange={(value) => writeColumnFilter(table, "status", value)}
        />
      )}

      {showColumnVisibility && <ColumnVisibilityMenu table={table} />}
    </div>
  );
}

function Pagination<TData extends RowData>({
  table,
}: {
  table: OrdersTableInstance<TData>;
}) {
  const { pageIndex, pageSize } = table.state.pagination;
  const pageCount = Math.max(1, table.getPageCount());
  const rowCount = table.getRowCount();

  return (
    <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-zinc-500">
      <span>
        {rowCount} {rowCount === 1 ? "row" : "rows"}
      </span>
      <div className="flex items-center gap-1.5">
        <button
          type="button"
          onClick={() => table.firstPage()}
          disabled={!table.getCanPreviousPage()}
          className="rounded border border-zinc-700 px-2 py-1 text-zinc-300 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-40"
        >
          First
        </button>
        <button
          type="button"
          onClick={() => table.previousPage()}
          disabled={!table.getCanPreviousPage()}
          className="rounded border border-zinc-700 px-2 py-1 text-zinc-300 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Prev
        </button>
        <span className="tabular-nums text-zinc-400">
          Page {pageIndex + 1} / {pageCount}
        </span>
        <button
          type="button"
          onClick={() => table.nextPage()}
          disabled={!table.getCanNextPage()}
          className="rounded border border-zinc-700 px-2 py-1 text-zinc-300 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Next
        </button>
        <button
          type="button"
          onClick={() => table.lastPage()}
          disabled={!table.getCanNextPage()}
          className="rounded border border-zinc-700 px-2 py-1 text-zinc-300 hover:bg-zinc-800 disabled:cursor-not-allowed disabled:opacity-40"
        >
          Last
        </button>
        <span className="tabular-nums">({pageSize}/page)</span>
      </div>
    </div>
  );
}

function OrdersTableRow<TData extends RowData>({
  row,
  getRowExpiry,
}: {
  row: Row<typeof orderTableFeatures, TData>;
  getRowExpiry?: (row: TData) => string | null | undefined;
}) {
  const stale = useIsExpired(getRowExpiry?.(row.original));

  return (
    <tr
      className={cn(
        "border-b border-zinc-800/70 transition-colors hover:bg-zinc-800/40",
        stale && "bg-red-500/5 opacity-60",
      )}
    >
      {row.getVisibleCells().map((cell) => {
        const align = cell.column.columnDef.meta?.align;
        return (
          <td
            key={cell.id}
            className={cn(
              "whitespace-nowrap px-2 py-1.5",
              alignClass(align),
              align === "right" && "tabular-nums",
            )}
          >
            <FlexRender cell={cell} />
          </td>
        );
      })}
    </tr>
  );
}

export function OrdersTable<TData extends RowData>({
  data,
  columns,
  isLoading = false,
  emptyMessage = "No orders",
  initialSorting,
  showSearch = true,
  showDirectionFilter = true,
  showTypeFilter = true,
  showStatusFilter = false,
  showColumnVisibility = true,
  showPagination = false,
  pageSize = HISTORY_PAGE_SIZE,
  getRowId,
  getRowExpiry,
}: OrdersTableProps<TData>) {
  const table = useTable({
    features: orderTableFeatures,
    columns,
    data,
    getRowId,
    globalFilterFn: orderGlobalFilter,
    initialState: {
      sorting: initialSorting,
      pagination: {
        pageIndex: 0,
        pageSize: showPagination ? pageSize : 10_000,
      },
    },
  });

  const rows = table.getRowModel().rows;
  const visibleColumnCount = Math.max(table.getVisibleLeafColumns().length, 1);

  return (
    <div className="flex min-h-0 flex-col gap-2">
      <TableToolbar
        table={table}
        showSearch={showSearch}
        showDirectionFilter={showDirectionFilter}
        showTypeFilter={showTypeFilter}
        showStatusFilter={showStatusFilter}
        showColumnVisibility={showColumnVisibility}
      />

      <div className="max-h-[min(58vh,42rem)] overflow-auto rounded-lg border border-zinc-800 bg-zinc-900/70">
        <table className="w-full border-collapse text-xs">
          <thead className="sticky top-0 z-10 bg-zinc-900">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id} className="border-b border-zinc-700">
                {headerGroup.headers.map((header) => {
                  const sorted = header.column.getIsSorted();
                  const canSort = header.column.getCanSort();
                  const align = header.column.columnDef.meta?.align;

                  return (
                    <th
                      key={header.id}
                      scope="col"
                      aria-sort={
                        sorted === "asc"
                          ? "ascending"
                          : sorted === "desc"
                            ? "descending"
                            : "none"
                      }
                      className={cn(
                        "whitespace-nowrap px-2 py-1.5 text-[10px] font-medium uppercase tracking-wider text-zinc-500",
                        alignClass(align),
                      )}
                    >
                      {header.isPlaceholder ? null : canSort ? (
                        <button
                          type="button"
                          onClick={header.column.getToggleSortingHandler()}
                          className={cn(
                            "inline-flex items-center gap-1 hover:text-zinc-200",
                            align === "right" && "w-full justify-end",
                          )}
                        >
                          <FlexRender header={header} />
                          <span className="text-[9px] text-zinc-600">
                            {sorted === "asc"
                              ? "▲"
                              : sorted === "desc"
                                ? "▼"
                                : "↕"}
                          </span>
                        </button>
                      ) : (
                        <FlexRender header={header} />
                      )}
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 6 }, (_, index) => (
                <tr
                  key={`skeleton-${index}`}
                  className="border-b border-zinc-800/70"
                >
                  <td colSpan={visibleColumnCount} className="px-2 py-2">
                    <div
                      className="h-4 animate-pulse rounded bg-zinc-800"
                      style={{ width: `${45 + ((index * 11) % 45)}%` }}
                    />
                  </td>
                </tr>
              ))
            ) : rows.length === 0 ? (
              <tr>
                <td
                  colSpan={visibleColumnCount}
                  className="px-4 py-10 text-center text-xs text-zinc-500"
                >
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              rows.map((row) => (
                <OrdersTableRow
                  key={row.id}
                  row={row}
                  getRowExpiry={getRowExpiry}
                />
              ))
            )}
          </tbody>
        </table>
      </div>

      {showPagination && !isLoading && <Pagination table={table} />}
    </div>
  );
}
