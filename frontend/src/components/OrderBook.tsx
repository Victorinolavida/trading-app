import { cn } from "./ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./ui/table";

export type BookLevel = {
  price: number | string;
  quantity: number | string;
  nominal_value: number | string;
};

type Side = "bid" | "ask";

const SIDE = {
  bid: {
    title: "Bids",
    color: "text-success",
    empty: "No bids",
    noun: "Bid",
  },
  ask: { title: "Asks", color: "text-danger", empty: "No asks", noun: "Ask" },
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
    <section className="min-w-0">
      <h3 className={cn("ui-section-title mb-2", s.color)}>{s.title}</h3>
      <Table className="tabular-nums">
        <TableHeader>
          <tr>
            <TableHead className="text-left">Level</TableHead>
            <TableHead title={`${s.noun} price`} className="text-right">
              Price
            </TableHead>
            <TableHead title={`${s.noun} quantity`} className="text-right">
              Quantity
            </TableHead>
            <TableHead title={`${s.noun} nominal value`} className="text-right">
              Nominal Value
            </TableHead>
          </tr>
        </TableHeader>

        <TableBody>
          {rows.length === 0 ? (
            <tr>
              <TableCell colSpan={4} className="py-6 text-center text-fg-muted">
                {s.empty}
              </TableCell>
            </tr>
          ) : (
            rows.map((l, i) => (
              <TableRow key={l.price}>
                <TableCell className="text-left text-fg-muted">
                  {i + 1}
                </TableCell>
                <TableCell className={cn("text-right", s.color)}>
                  {Number(l.price).toFixed(2)}
                </TableCell>
                <TableCell className="text-right text-fg-muted">
                  {l.quantity}
                </TableCell>
                <TableCell className="text-right text-fg-muted">
                  {l.nominal_value}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </section>
  );
}

export function OrderBook({
  book,
}: {
  book: { bids: BookLevel[]; asks: BookLevel[] };
}) {
  return (
    <div className="ui-panel p-4">
      <h2 className="ui-section-title mb-3">Order Book</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        <BookSide side="bid" levels={book.bids} />
        <BookSide side="ask" levels={book.asks} />
      </div>
    </div>
  );
}
