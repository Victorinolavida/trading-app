import { useMemo } from "react";
import {
  Area,
  AreaChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { BookLevel } from "./OrderBook";

type DepthPoint = {
  price: number;
  bid?: number;
  ask?: number;
};

type Parsed = {
  price: number;
  qty: number;
};

type DepthResult = {
  data: DepthPoint[];
  mid: number | undefined;
};

const SERIES_LABEL = {
  bid: "Bid depth",
  ask: "Ask depth",
} as const;

type SeriesKey = keyof typeof SERIES_LABEL;

function isSeriesKey(k: string): k is SeriesKey {
  return k in SERIES_LABEL;
}

/**
 * Aggregate orders at the same price.
 */
function parseLevels(levels: BookLevel[]): Parsed[] {
  const grouped = new Map<number, number>();

  for (const level of levels) {
    const price = Number(level.price);
    const qty = Number(level.quantity);

    if (!Number.isFinite(price) || !Number.isFinite(qty) || qty < 0) {
      continue;
    }

    grouped.set(price, (grouped.get(price) ?? 0) + qty);
  }

  return Array.from(grouped, ([price, qty]) => ({
    price,
    qty,
  }));
}

function buildDepth(
  bids: BookLevel[],
  asks: BookLevel[],
  maxLevels = 20,
): DepthResult {
  const bidLevels = parseLevels(bids)
    .sort((a, b) => b.price - a.price)
    .slice(0, maxLevels);

  const askLevels = parseLevels(asks)
    .sort((a, b) => a.price - b.price)
    .slice(0, maxLevels);

  const bestBid = bidLevels[0]?.price;
  const bestAsk = askLevels[0]?.price;

  const mid =
    bestBid !== undefined && bestAsk !== undefined
      ? (bestBid + bestAsk) / 2
      : undefined;

  const bidPts: DepthPoint[] = bidLevels
    .map(({ price, qty }) => ({
      price,
      bid: qty,
    }))
    .reverse();

  const askPts: DepthPoint[] = askLevels.map(({ price, qty }) => ({
    price,
    ask: qty,
  }));

  return {
    data: [...bidPts, ...askPts],
    mid,
  };
}

type Props = {
  bids: BookLevel[];
  asks: BookLevel[];
  height?: number;
  maxLevels?: number;
};

export function DepthChart({
  bids,
  asks,
  height = 220,
  maxLevels = 20,
}: Props) {
  const { data, mid } = useMemo(
    () => buildDepth(bids, asks, maxLevels),
    [bids, asks, maxLevels],
  );

  if (data.length === 0) return null;

  return (
    <div className="w-full bg-zinc-900 p-4 rounded-lg shadow-lg">
      <h3 className="mb-2 text-sm font-semibold text-zinc-300">Order Depth</h3>
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart
          data={data}
          margin={{
            top: 8,
            right: 8,
            bottom: 0,
            left: 0,
          }}
        >
          <XAxis
            dataKey="price"
            type="number"
            domain={["dataMin", "dataMax"]}
            tickFormatter={(v) => Number(v).toFixed(2)}
            tick={{ fill: "#9ca3af", fontSize: 11 }}
            stroke="#3f3f46"
            allowDataOverflow={false}
          />

          <YAxis
            type="number"
            orientation="right"
            width={55}
            tick={{ fill: "#9ca3af", fontSize: 11 }}
            stroke="#3f3f46"
            tickFormatter={(v) => Number(v).toLocaleString()}
          />

          <Tooltip
            contentStyle={{
              background: "#18181b",
              border: "1px solid #3f3f46",
              borderRadius: 8,
            }}
            labelStyle={{ color: "#fff" }}
            labelFormatter={(label) => `Price: ${Number(label).toString()}`}
            formatter={(value, name) => {
              const key = String(name);

              return [
                Number(value ?? 0).toLocaleString(),
                isSeriesKey(key) ? SERIES_LABEL[key] : key,
              ];
            }}
          />

          <Area
            type="stepAfter"
            dataKey="bid"
            name="bid"
            stroke="#4ade80"
            fill="#4ade80"
            fillOpacity={0.2}
            strokeWidth={2}
            isAnimationActive={false}
            connectNulls={false}
          />

          <Area
            type="stepBefore"
            dataKey="ask"
            name="ask"
            stroke="#f87171"
            fill="#f87171"
            fillOpacity={0.2}
            strokeWidth={2}
            isAnimationActive={false}
            connectNulls={false}
          />

          {mid !== undefined && (
            <ReferenceLine
              x={mid}
              stroke="#a1a1aa"
              strokeDasharray="4 4"
              label={{
                value: `Mid ${mid}`,
                fill: "#a1a1aa",
                fontSize: 10,
                position: "insideTop",
              }}
            />
          )}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
