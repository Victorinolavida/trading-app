import { SecuritySearch } from "../components/SecuritySearch";

export default function HomePage() {
  return (
    <div className="mx-auto w-full max-w-5xl">
      <header className="page-heading mb-8 flex items-end justify-between gap-6">
        <h1 className="ui-title">Securities</h1>
        <span aria-hidden="true" className="h-2 w-12 bg-accent" />
      </header>
      <div className="grid gap-6 md:grid-cols-[1fr_2fr] md:gap-12">
        <div>
          <h2 className="text-lg font-semibold text-fg-strong">Find a security</h2>
          <p className="mt-2 text-sm leading-6 text-fg-muted">Search by name or ISIN to view the order book and create an order.</p>
        </div>
        <SecuritySearch />
      </div>
    </div>
  );
}
