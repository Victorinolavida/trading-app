import { useState } from "react";
import { useSearchSecurities } from "../hooks/useSecurities";

export function SecuritySearch() {
  const [term, setTerm] = useState("");
  const { data, isLoading } = useSearchSecurities(term, 300);

  const securities = data?.securities ?? [];

  const goToSecurity = (id: string) => {
    window.location.hash = `#/securities/${id}`;
  };

  return (
    <div className="w-full">
      <label
        className="mb-2 block text-sm text-gray-400"
        htmlFor="security-search"
      >
        Search security
      </label>
      <input
        id="security-search"
        type="text"
        value={term}
        onChange={(e) => setTerm(e.target.value)}
        placeholder="Search by name or ISIN..."
        className="w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
      />
      {isLoading && term.trim().length > 0 && (
        <p className="mt-2 text-sm text-gray-400">Searching...</p>
      )}
      {securities.length > 0 && (
        <ul className="mt-2 max-h-80 overflow-auto rounded-md border border-zinc-800 bg-zinc-950">
          {securities.map((s) => (
            <li
              key={s.id}
              onClick={() => goToSecurity(s.id)}
              className="cursor-pointer border-b border-zinc-900 px-3 py-2 hover:bg-zinc-900 last:border-b-0"
            >
              <div className="text-sm font-medium">{s.name || s.isin}</div>
              <div className="text-xs text-gray-400">ISIN: {s.isin}</div>
              <div className="text-xs text-gray-400">
                {s.security_type}
                {s.serie ? ` · Serie ${s.serie}` : ""} · Maturity:{" "}
                {s.maturity_date}
              </div>
            </li>
          ))}
        </ul>
      )}
      {term.trim().length > 0 && !isLoading && securities.length === 0 && (
        <p className="mt-2 text-sm text-gray-400">No securities found</p>
      )}
    </div>
  );
}
