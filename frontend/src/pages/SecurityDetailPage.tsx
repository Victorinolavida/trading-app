import { useEffect, useState } from "react";
import { getSecurityByParams, getSecuritiesByParams, type Security } from "../lib/api";
import { Sidebar } from "../components/Sidebar";

type Order = {
  id: string;
  side: "ask" | "bid";
  price: number;
  quantity: number;
};

export function SecurityDetailPage({ securityId }: { securityId: string }) {
  const [security, setSecurity] = useState<Security | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // placeholder orders
  const [orders] = useState<Order[]>([
    { id: "1", side: "bid", price: 98.45, quantity: 100 },
    { id: "2", side: "bid", price: 98.42, quantity: 250 },
    { id: "3", side: "ask", price: 98.5, quantity: 150 },
    { id: "4", side: "ask", price: 98.52, quantity: 300 },
  ]);

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    setError(null);

    // try to fetch by id first
    getSecurityByParams({ id: securityId })
      .then((res) => {
        if (!mounted) return;
        setSecurity(res);
        setLoading(false);
      })
      .catch(async () => {
        try {
          const res = await getSecuritiesByParams({ id: securityId });
          if (!mounted) return;
          if (res.securities?.[0]) {
            setSecurity(res.securities[0]);
            setLoading(false);
            return;
          }
        } catch {
          // ignore
        }
        if (mounted) {
          setError("Security not found");
          setLoading(false);
        }
      });

    return () => {
      mounted = false;
    };
  }, [securityId]);

  const bids = orders.filter((o) => o.side === "bid").sort((a, b) => b.price - a.price);
  const asks = orders.filter((o) => o.side === "ask").sort((a, b) => a.price - b.price);

  return (
    <div className="flex min-h-screen overflow-x-hidden">
      <Sidebar />
      <main className="ml-64 flex-1 min-h-screen overflow-x-hidden p-8">
        <div className="mx-auto w-full max-w-5xl">
          <button
            className="mb-4 cursor-pointer text-sm text-blue-400"
            onClick={() => (window.location.hash = "")}
          >
            ← Back
          </button>

          {loading && <p className="text-gray-400">Loading...</p>}
          {error && <p className="text-red-400">{error}</p>}

          {security && (
            <>
              <div className="mb-6 rounded-lg border border-zinc-800 bg-zinc-900 p-6">
                <h1 className="mb-2 text-2xl font-semibold">
                  {security.name || `Mexico Government Bond (${security.security_type})`}
                </h1>
                {security.name && <p className="mb-4 text-gray-400">{security.security_type === "M" ? "Bono M" : security.security_type === "S" ? "UDI Bond" : security.security_type}</p>}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div className="text-gray-400">ISIN</div>
                    <div>{security.isin}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">ID</div>
                    <div className="break-all">{security.id}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Type</div>
                    <div>{security.security_type}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Serie</div>
                    <div>{security.serie}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Currency</div>
                    <div>{security.currency}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Nominal Value</div>
                    <div>{security.nominal_value}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Issue Date</div>
                    <div>{security.issue_date}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Maturity Date</div>
                    <div>{security.maturity_date}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">FX</div>
                    <div>{security.fx}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Enabled</div>
                    <div>{security.enable ? "Yes" : "No"}</div>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
                  <h2 className="mb-3 text-lg font-semibold">Bids (Bid)</h2>
                  <div className="space-y-1">
                    {bids.map((o) => (
                      <div key={o.id} className="flex justify-between text-sm">
                        <span className="text-green-400">{o.price.toFixed(2)}</span>
                        <span className="text-gray-400">{o.quantity}</span>
                      </div>
                    ))}
                    {bids.length === 0 && <p className="text-sm text-gray-400">No bids</p>}
                  </div>
                </div>
                <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
                  <h2 className="mb-3 text-lg font-semibold">Asks (Ask)</h2>
                  <div className="space-y-1">
                    {asks.map((o) => (
                      <div key={o.id} className="flex justify-between text-sm">
                        <span className="text-red-400">{o.price.toFixed(2)}</span>
                        <span className="text-gray-400">{o.quantity}</span>
                      </div>
                    ))}
                    {asks.length === 0 && <p className="text-sm text-gray-400">No asks</p>}
                  </div>
                </div>
              </div>

              <div className="mt-4 rounded-lg border border-zinc-800 bg-zinc-900 p-4">
                <h2 className="mb-3 text-lg font-semibold">Price Chart (Placeholder)</h2>
                <div className="flex h-64 items-center justify-center text-gray-400">
                  Chart will go here
                </div>
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  );
}
