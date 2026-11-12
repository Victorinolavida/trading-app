// import { useCallback, useEffect, useState } from "react";
// import {
//   getSecurityByParams,
//   getSecuritiesByParams,
//   createOrder,
//   type Security,
// } from "../lib/api";
// import { useOrderBookSSE } from "../hooks/useOrderBookSSE";
// import { Sidebar } from "../components/Sidebar";
import { useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { securityQueries } from "../features/securities/queries";

export default function SecurityDetailPage() {
  const { isin } = useParams();
  const { data: security } = useQuery(securityQueries.detail({ isin: isin! }));

  return <>{JSON.stringify(security)}</>;
  // const [security, setSecurity] = useState<Security | null>(null);
  // const [loading, setLoading] = useState(true);
  // const [error, setError] = useState<string | null>(null);

  // const [side, setSide] = useState<"bid" | "ask">("bid");
  // const [orderType, setOrderType] = useState<"limit" | "market">("limit");
  // const [price, setPrice] = useState("");
  // const [yieldVal, setYieldVal] = useState("");
  // const [quantity, setQuantity] = useState("");
  // const [creating, setCreating] = useState(false);
  // const [createMsg, setCreateMsg] = useState<string | null>(null);

  // // State for order book
  // const [book, setBook] = useState<{
  //   bids: any[];
  //   asks: any[];
  // }>({ bids: [], asks: [] });

  // // SSE update handler
  // const handleOrderBookUpdate = useCallback((update: any) => {
  //   setBook({
  //     bids: update.payload?.bids ?? [],
  //     asks: update.payload?.asks ?? [],
  //   });
  // }, []);

  // // Executed trade handler
  // const handleExecutedTrade = useCallback((trade: any) => {
  //   console.log("Executed trade:", trade);
  //   // Could show a notification here
  // }, []);

  // useEffect(() => {
  //   let mounted = true;
  //   setLoading(true);
  //   setError(null);

  //   // try to fetch by id first
  //   getSecurityByParams({ id: securityId })
  //     .then((res) => {
  //       if (!mounted) return;
  //       setSecurity(res);
  //       setLoading(false);
  //     })
  //     .catch(async () => {
  //       try {
  //         const res = await getSecuritiesByParams({ id: securityId });
  //         if (!mounted) return;
  //         if (res.securities?.[0]) {
  //           setSecurity(res.securities[0]);
  //           setLoading(false);
  //           return;
  //         }
  //       } catch {
  //         // ignore
  //       }
  //       if (mounted) {
  //         setError("Security not found");
  //         setLoading(false);
  //       }
  //     });

  //   return () => {
  //     mounted = false;
  //   };
  // }, [securityId]);

  // // Subscribe to order book SSE when security is loaded
  // useOrderBookSSE(
  //   security?.id ?? "",
  //   handleOrderBookUpdate,
  //   handleExecutedTrade,
  // );

  // const handleCreateOrder = async (e: React.FormEvent) => {
  //   e.preventDefault();
  //   if (!security) return;
  //   setCreating(true);
  //   setCreateMsg(null);
  //   try {
  //     const res = await createOrder({
  //       security_id: security.id,
  //       direction: side,
  //       order_type: orderType,
  //       price: price || "0",
  //       yield: yieldVal || "0",
  //       quantity: Number(quantity),
  //     });
  //     setCreateMsg(
  //       `Order created: group ${res.order_group_id}, order ${res.order_id}, status ${res.status}`,
  //     );
  //     setQuantity("");
  //   } catch (err: any) {
  //     setCreateMsg(err?.message || "Failed to create order");
  //   } finally {
  //     setCreating(false);
  //   }
  // };

  // return (
  //   <div className="flex min-h-screen overflow-x-hidden">
  //     <Sidebar />
  //     <main className="ml-64 flex-1 min-h-screen overflow-x-hidden p-8">
  //       <div className="mx-auto w-full max-w-5xl">
  //         <button
  //           className="mb-4 cursor-pointer text-sm text-blue-400"
  //           onClick={() => (window.location.hash = "")}
  //         >
  //           ← Back
  //         </button>

  //         {loading && <p className="text-gray-400">Loading...</p>}
  //         {error && <p className="text-red-400">{error}</p>}

  //         {security && (
  //           <>
  //             <div className="mb-6 rounded-lg border border-zinc-800 bg-zinc-900 p-6">
  //               <h1 className="mb-2 text-2xl font-semibold">
  //                 {security.name ||
  //                   `Mexico Government Bond (${security.security_type})`}
  //               </h1>
  //               {security.name && (
  //                 <p className="mb-4 text-gray-400">
  //                   {security.security_type === "M"
  //                     ? "Bono M"
  //                     : security.security_type === "S"
  //                       ? "UDI Bond"
  //                       : security.security_type}
  //                 </p>
  //               )}
  //               <div className="grid grid-cols-2 gap-4 text-sm">
  //                 <div>
  //                   <div className="text-gray-400">ISIN</div>
  //                   <div>{security.isin}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">ID</div>
  //                   <div className="break-all">{security.id}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Type</div>
  //                   <div>{security.security_type}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Serie</div>
  //                   <div>{security.serie}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Currency</div>
  //                   <div>{security.currency}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Nominal Value</div>
  //                   <div>{security.nominal_value}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Issue Date</div>
  //                   <div>{security.issue_date}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Maturity Date</div>
  //                   <div>{security.maturity_date}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">FX</div>
  //                   <div>{security.fx}</div>
  //                 </div>
  //                 <div>
  //                   <div className="text-gray-400">Enabled</div>
  //                   <div>{security.enable ? "Yes" : "No"}</div>
  //                 </div>
  //               </div>
  //             </div>

  //             <div className="mb-4 rounded-lg border border-zinc-800 bg-zinc-900 p-4">
  //               <h2 className="mb-3 text-lg font-semibold">Create Order</h2>
  //               <form
  //                 onSubmit={handleCreateOrder}
  //                 className="grid grid-cols-2 gap-3"
  //               >
  //                 <div>
  //                   <label className="mb-1 block text-xs text-gray-400">
  //                     Side
  //                   </label>
  //                   <select
  //                     value={side}
  //                     onChange={(e) => setSide(e.target.value as "bid" | "ask")}
  //                     className="w-full rounded bg-zinc-800 p-2 text-sm"
  //                   >
  //                     <option value="bid">Bid</option>
  //                     <option value="ask">Ask</option>
  //                   </select>
  //                 </div>
  //                 <div>
  //                   <label className="mb-1 block text-xs text-gray-400">
  //                     Type
  //                   </label>
  //                   <select
  //                     value={orderType}
  //                     onChange={(e) =>
  //                       setOrderType(e.target.value as "limit" | "market")
  //                     }
  //                     className="w-full rounded bg-zinc-800 p-2 text-sm"
  //                   >
  //                     <option value="limit">Limit</option>
  //                     <option value="market">Market</option>
  //                   </select>
  //                 </div>
  //                 <div>
  //                   <label className="mb-1 block text-xs text-gray-400">
  //                     Price
  //                   </label>
  //                   <input
  //                     value={price}
  //                     onChange={(e) => setPrice(e.target.value)}
  //                     placeholder="0"
  //                     className="w-full rounded bg-zinc-800 p-2 text-sm"
  //                   />
  //                 </div>
  //                 <div>
  //                   <label className="mb-1 block text-xs text-gray-400">
  //                     Yield
  //                   </label>
  //                   <input
  //                     value={yieldVal}
  //                     onChange={(e) => setYieldVal(e.target.value)}
  //                     placeholder="0"
  //                     className="w-full rounded bg-zinc-800 p-2 text-sm"
  //                   />
  //                 </div>
  //                 <div className="col-span-2">
  //                   <label className="mb-1 block text-xs text-gray-400">
  //                     Quantity
  //                   </label>
  //                   <input
  //                     value={quantity}
  //                     onChange={(e) => setQuantity(e.target.value)}
  //                     required
  //                     placeholder="e.g. 100"
  //                     className="w-full rounded bg-zinc-800 p-2 text-sm"
  //                   />
  //                 </div>
  //                 <button
  //                   type="submit"
  //                   disabled={creating}
  //                   className="col-span-2 rounded bg-blue-600 px-4 py-2 text-sm hover:bg-blue-500 disabled:opacity-50"
  //                 >
  //                   {creating ? "Creating..." : "Create Order"}
  //                 </button>
  //               </form>
  //               {createMsg && (
  //                 <p className="mt-2 text-sm text-gray-300">{createMsg}</p>
  //               )}
  //             </div>

  //             <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
  //               <h2 className="mb-3 text-lg font-semibold">Order Book</h2>
  //               <div className="space-y-4">
  //                 <div>
  //                   <h3 className="mb-2 text-sm font-semibold text-green-400">
  //                     Bids
  //                   </h3>
  //                   <div className="space-y-1">
  //                     <div className="mb-1 flex justify-between text-xs text-gray-400">
  //                       <span>Level</span>
  //                       <span title="Bid price">Price</span>
  //                       <span title="Bid quantity">Quantity</span>
  //                       <span title="Bid nominal value">Nominal Value</span>
  //                     </div>
  //                     {book.bids.slice(0, 5).map((b: any, i: number) => (
  //                       <div
  //                         key={b.price}
  //                         className="flex justify-between text-sm"
  //                       >
  //                         <span className="text-gray-400">{i + 1}</span>
  //                         <span className="text-green-400">
  //                           {Number(b.price).toFixed(2)}
  //                         </span>
  //                         <span className="text-gray-400">{b.quantity}</span>
  //                         <span className="text-gray-400">
  //                           {b.nominal_value}
  //                         </span>
  //                       </div>
  //                     ))}
  //                     {book.bids.length === 0 && (
  //                       <p className="text-sm text-gray-400">No bids</p>
  //                     )}
  //                   </div>
  //                 </div>
  //                 <div>
  //                   <h3 className="mb-2 text-sm font-semibold text-red-400">
  //                     Asks
  //                   </h3>
  //                   <div className="space-y-1">
  //                     <div className="mb-1 flex justify-between text-xs text-gray-400">
  //                       <span>Level</span>
  //                       <span title="Ask price">Price</span>
  //                       <span title="Ask quantity">Quantity</span>
  //                       <span title="Ask nominal value">Nominal Value</span>
  //                     </div>
  //                     {book.asks.slice(0, 5).map((a: any, i: number) => (
  //                       <div
  //                         key={a.price}
  //                         className="flex justify-between text-sm"
  //                       >
  //                         <span className="text-gray-400">{i + 1}</span>
  //                         <span className="text-red-400">
  //                           {Number(a.price).toFixed(2)}
  //                         </span>
  //                         <span className="text-gray-400">{a.quantity}</span>
  //                         <span className="text-gray-400">
  //                           {a.nominal_value}
  //                         </span>
  //                       </div>
  //                     ))}
  //                     {book.asks.length === 0 && (
  //                       <p className="text-sm text-gray-400">No asks</p>
  //                     )}
  //                   </div>
  //                 </div>
  //               </div>
  //             </div>

  //             <div className="mt-4 rounded-lg border border-zinc-800 bg-zinc-900 p-4">
  //               <h2 className="mb-3 text-lg font-semibold">
  //                 Price Chart (Placeholder)
  //               </h2>
  //               <div className="flex h-64 items-center justify-center text-gray-400">
  //                 Chart will go here
  //               </div>
  //             </div>
  //           </>
  //         )}
  //       </div>
  //     </main>
  //   </div>
  // );
}
