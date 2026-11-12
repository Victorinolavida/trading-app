import { SecuritySearch } from "../components/SecuritySearch";

export default function HomePage() {
  return (
    <div className="mx-auto w-full max-w-md">
      <h1 className="mb-6 text-2xl font-semibold">Securities</h1>
      <SecuritySearch />
    </div>
  );
}
