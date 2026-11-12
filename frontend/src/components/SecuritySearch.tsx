import { ChangeEvent, useCallback, useState } from "react";
import { useDebounce } from "@uidotdev/usehooks";
import { useSearchSecuritiesByParams } from "../features/securities/hooks";
import { useNavigate } from "react-router";

export function SecuritySearch() {
  const [term, setTerm] = useState("");
  const debouncedTerm = useDebounce(term, 500);
  const { isLoading, data } = useSearchSecuritiesByParams({
    query: debouncedTerm,
  });

  const navigate = useNavigate();
  const handleOnChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const { value } = e.target;
      setTerm(value.trim());
    },
    [setTerm],
  );

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
        onChange={handleOnChange}
        placeholder="Search by name or ISIN..."
        className="w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
      />
      {isLoading && term.length > 0 && (
        <p className="mt-2 text-sm text-gray-400">Searching...</p>
      )}

      {!isLoading && term.length > 0 && data && data.securities.length <= 0 && (
        <p className="mt-2 text-sm text-gray-400">No securities found</p>
      )}
      {!isLoading && data && data.securities && (
        <ul className="mt-2 max-h-80 overflow-auto rounded-md border border-zinc-800 bg-zinc-950">
          {data.securities.map((s) => (
            <li
              key={s.id}
              className="cursor-pointer border-b border-zinc-900 px-3 py-2 hover:bg-zinc-900 last:border-b-0"
              onClick={() => navigate("/securities/" + s.id)}
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
    </div>
  );
}
