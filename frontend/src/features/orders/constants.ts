import type { AllocationStatus } from "./types";

export const ACTIVE_STATUS = "active" as const;

export const HISTORY_STATUSES = [
  "fully_allocated",
  "cancelled",
  "expired",
  "killed",
] as const;

export const ORDER_STATUSES = [ACTIVE_STATUS, ...HISTORY_STATUSES] as const;

export const ORDER_DIRECTIONS = ["bid", "ask"] as const;

export const ORDER_TYPES = ["limit", "market"] as const;

export const ALLOCATION_STATUSES = ["none", "partial", "allocated"] as const;

export const HISTORY_PAGE_SIZE = 25;

export const allocationStatusClasses: Record<AllocationStatus, string> = {
  none: "text-fg-muted",
  partial: "text-warning",
  allocated: "text-success",
};
