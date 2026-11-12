import { cn } from "./ui/card"; // or wherever your cn lives

export type BookLevel = {
  price: number | string;
  quantity: number | string;
  nominal_value: number | string;
};

type Side = "bid" | "ask";

const SIDE = {
  bid: {
    title: "Bids",
    color: "text-green-400",
    empty: "No bids",
    noun: "Bid",
  },
  ask: { title: "Asks", color: "text-red-400", empty: "No asks", noun: "Ask" },
} as const;

function BookSide({
  side,
  levels,
  depth = 5,
}: {
  side: Side;
  levels: BookLevel[];
  depth?: number;
}) {
  const s = SIDE[side];
  const rows = levels.slice(0, depth);

  return (
    <section>
      <h3 className={cn("mb-2 text-sm font-semibold", s.color)}>{s.title}</h3>

      <table className="w-full tabular-nums">
        <thead className="text-xs text-gray-400">
          <tr>
            <th scope="col" className="pb-1 text-left font-normal">
              Level
            </th>
            <th
              scope="col"
              title={`${s.noun} price`}
              className="pb-1 text-right font-normal"
            >
              Price
            </th>
            <th
              scope="col"
              title={`${s.noun} quantity`}
              className="pb-1 text-right font-normal"
            >
              Quantity
            </th>
            <th
              scope="col"
              title={`${s.noun} nominal value`}
              className="pb-1 text-right font-normal"
            >
              Nominal Value
            </th>
          </tr>
        </thead>

        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td colSpan={4} className="py-1 text-gray-400">
                {s.empty}
              </td>
            </tr>
          ) : (
            rows.map((l, i) => (
              <tr key={l.price}>
                <td className="py-0.5 text-left text-gray-400">{i + 1}</td>
                <td className={cn("py-0.5 text-right", s.color)}>
                  {Number(l.price).toFixed(2)}
                </td>
                <td className="py-0.5 text-right text-gray-400">
                  {l.quantity}
                </td>
                <td className="py-0.5 text-right text-gray-400">
                  {l.nominal_value}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </section>
  );
}

export function OrderBook({
  book,
}: {
  book: { bids: BookLevel[]; asks: BookLevel[] };
}) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
      <h2 className="mb-3 text-lg font-semibold">Order Book</h2>
      <div className="space-y-4">
        <BookSide side="bid" levels={book.bids} />
        <BookSide side="ask" levels={book.asks} />
      </div>
    </div>
  );
}
