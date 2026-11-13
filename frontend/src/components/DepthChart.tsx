import { useMemo } from "react";
import {
  Area,
  AreaChart,
  Legend,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { BookLevel } from "./OrderBook";
import { chartColors } from "./ui/theme";

type DepthPoint = {
  price: number;
  bid?: number;
  ask?: number;
};

type Parsed = { price: number; qty: number };

type DepthResult = {
  data: DepthPoint[];
  mid?: number;
  domain: [number, number];
};

const SERIES_LABEL = {
  bid: "Bids",
  ask: "Asks",
} as const;

type SeriesKey = keyof typeof SERIES_LABEL;

function isSeriesKey(k: string): k is SeriesKey {
  return k in SERIES_LABEL;
}

function parseLevels(levels: BookLevel[]): Parsed[] {
  const grouped = new Map<number, number>();
  for (const level of levels) {
    const price = Number(level.price);
    const qty = Number(level.quantity);
    if (!Number.isFinite(price) || !Number.isFinite(qty) || qty < 0) continue;
    grouped.set(price, (grouped.get(price) ?? 0) + qty);
  }
  return Array.from(grouped, ([price, qty]) => ({ price, qty }));
}

function cumulate(levels: Parsed[]): { price: number; total: number }[] {
  let acc = 0;
  return levels.map(({ price, qty }) => {
    acc += qty;
    return { price, total: acc };
  });
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

  const bidCum = cumulate(bidLevels);
  const askCum = cumulate(askLevels);

  const bestBid = bidLevels[0]?.price;
  const bestAsk = askLevels[0]?.price;
  const mid =
    bestBid !== undefined && bestAsk !== undefined
      ? (bestBid + bestAsk) / 2
      : undefined;

  const allPrices = [...bidCum, ...askCum].map((d) => d.price);
  if (allPrices.length === 0) {
    return { data: [], mid, domain: [0, 1] };
  }
  const min = Math.min(...allPrices);
  const max = Math.max(...allPrices);
  const ref = mid ?? bestBid ?? bestAsk ?? 0;
  const spread =
    bestBid !== undefined && bestAsk !== undefined ? bestAsk - bestBid : 0;

  const pad = Math.max(
    (max - min) * 0.1,
    spread * 2,
    Math.abs(ref) * 0.005,
    0.01,
  );
  const domain: [number, number] = [min - pad, max + pad];

  const bidPts: DepthPoint[] = bidCum
    .map(({ price, total }) => ({ price, bid: total }))
    .reverse(); 

  const askPts: DepthPoint[] = askCum.map(({ price, total }) => ({
    price,
    ask: total,
  }));

  if (bidPts.length > 0) {
    bidPts.unshift({ price: domain[0], bid: bidPts[0].bid! });
  }
  if (askPts.length > 0) {
    askPts.push({ price: domain[1], ask: askPts[askPts.length - 1].ask! });
  }

  const midRow: DepthPoint[] =
    mid !== undefined
      ? [{ price: mid, bid: bidCum[0].total, ask: askCum[0].total }]
      : [];

  return {
    data: [...bidPts, ...midRow, ...askPts],
    mid,
    domain,
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
  height = 260,
  maxLevels = 20,
}: Props) {
  const { data, mid, domain } = useMemo(
    () => buildDepth(bids, asks, maxLevels),
    [bids, asks, maxLevels],
  );

  if (data.length === 0) return null;

  return (
    <div className="w-full bg-surface-raised p-4 rounded-lg shadow-lg">
      <h3 className="mb-2 text-sm font-semibold text-fg-soft">
        Order Book Depth
      </h3>
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart
          data={data}
          margin={{ top: 8, right: 8, bottom: 0, left: 0 }}
        >
          <XAxis
            dataKey="price"
            type="number"
            domain={domain}
            allowDataOverflow
            tickFormatter={(v) => Number(v).toFixed(2)}
            tick={{ fill: chartColors.axis, fontSize: 11 }}
            stroke={chartColors.grid}
          />
          <YAxis
            type="number"
            orientation="right"
            width={55}
            tick={{ fill: chartColors.axis, fontSize: 11 }}
            stroke={chartColors.grid}
            tickFormatter={(v) => Number(v).toLocaleString()}
          />

          <Tooltip
            contentStyle={{
              background: chartColors.tooltipSurface,
              border: `1px solid ${chartColors.tooltipBorder}`,
              borderRadius: 8,
            }}
            labelStyle={{ color: chartColors.tooltipText }}
            labelFormatter={(label) => `Price: ${Number(label)}`}
            formatter={(value, name) => {
              const key = String(name);
              return [
                Number(value ?? 0).toLocaleString(),
                isSeriesKey(key) ? `${SERIES_LABEL[key]} (acum.)` : key,
              ];
            }}
          />

          <Legend
            verticalAlign="top"
            align="right"
            iconType="plainline"
            wrapperStyle={{ paddingBottom: 16 }}
            formatter={(v) => (isSeriesKey(v) ? SERIES_LABEL[v] : v)}
          />

          <Area
            type="stepAfter"
            dataKey="bid"
            name="bid"
            stroke={chartColors.bid}
            fill={chartColors.bid}
            fillOpacity={0.2}
            strokeWidth={2}
            isAnimationActive={false}
            connectNulls={false}
          />

          <Area
            type="stepBefore"
            dataKey="ask"
            name="ask"
            stroke={chartColors.ask}
            fill={chartColors.ask}
            fillOpacity={0.2}
            strokeWidth={2}
            isAnimationActive={false}
            connectNulls={false}
          />

          {mid !== undefined && (
            <ReferenceLine
              x={mid}
              stroke={chartColors.mid}
              strokeDasharray="4 4"
              strokeWidth={1.5}
              label={{
                value: `Mid ${mid.toFixed(2)}`,
                fill: chartColors.mid,
                fontSize: 11,
                position: "insideTop",
                offset: 8,
              }}
            />
          )}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
