import type { ReactNode } from "react";

export type BadgeTone = "green" | "red" | "blue" | "purple" | "amber" | "zinc";

const tones: Record<BadgeTone, string> = {
  green: "border-green-500/30 bg-green-500/15 text-green-400",
  red: "border-red-500/30 bg-red-500/15 text-red-400",
  blue: "border-blue-500/30 bg-blue-500/15 text-blue-400",
  purple: "border-purple-500/30 bg-purple-500/15 text-purple-400",
  amber: "border-amber-500/30 bg-amber-500/15 text-amber-400",
  zinc: "border-zinc-500/30 bg-zinc-500/15 text-zinc-300",
};

export function Badge({
  tone = "zinc",
  children,
}: {
  tone?: BadgeTone;
  children: ReactNode;
}) {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize ${tones[tone]}`}
    >
      {children}
    </span>
  );
}

const directionTones: Record<string, BadgeTone> = {
  bid: "green",
  ask: "red",
};

const orderTypeTones: Record<string, BadgeTone> = {
  limit: "blue",
  market: "purple",
};

const statusTones: Record<string, BadgeTone> = {
  active: "green",
  fully_allocated: "blue",
  cancelled: "zinc",
  expired: "amber",
  killed: "red",
};

export function DirectionBadge({ direction }: { direction?: string }) {
  const value = (direction ?? "-").toLowerCase();
  return <Badge tone={directionTones[value] ?? "zinc"}>{value}</Badge>;
}

export function OrderTypeBadge({ orderType }: { orderType?: string }) {
  const value = (orderType ?? "-").toLowerCase();
  return <Badge tone={orderTypeTones[value] ?? "zinc"}>{value}</Badge>;
}

export function StatusBadge({ status }: { status?: string }) {
  const value = (status ?? "-").toLowerCase();
  return (
    <Badge tone={statusTones[value] ?? "zinc"}>
      {value.replace(/_/g, " ")}
    </Badge>
  );
}
