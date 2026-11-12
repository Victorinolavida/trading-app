import type { ReactNode } from "react";

export type BadgeTone = "green" | "red" | "blue" | "purple" | "amber" | "zinc";

const tones: Record<BadgeTone, string> = {
  green: "border-success/30 bg-success/15 text-success",
  red: "border-danger/30 bg-danger/15 text-danger",
  blue: "border-accent/30 bg-accent/15 text-accent-soft",
  purple: "border-special/30 bg-special/15 text-special",
  amber: "border-warning/30 bg-warning/15 text-warning",
  zinc: "border-fg-muted/30 bg-fg-muted/15 text-fg-soft",
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
