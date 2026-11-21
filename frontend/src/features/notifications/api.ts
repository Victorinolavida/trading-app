import { ApiClient } from "../../lib/http";

export interface Notification {
  id: string;
  event_type: string;
  payload: {
    order?: {
      id?: string;
      security_id?: string;
      status?: string;
      price?: string;
      yield?: string;
      quantity?: number;
      allocation_status?: string;
    };
    reason?: string;
  };
  created_at: string;
  read_at: string | null;
}

export interface NotificationPage {
  notifications: Notification[];
  next_cursor?: string | null;
}

const client = new ApiClient();

export const notificationsApi = {
  list: (cursor?: string, signal?: AbortSignal) =>
    client.get<NotificationPage>("/v0/notifications", {
      params: cursor ? { cursor } : undefined,
      signal,
    }),
  unreadCount: (signal?: AbortSignal) =>
    client.get<{ count: number }>("/v0/notifications/unread-count", { signal }),
  markRead: (id: string) =>
    client.patch<void>(`/v0/notifications/${encodeURIComponent(id)}/read`),
  markAllRead: () => client.post<void>("/v0/notifications/read-all"),
};

export const notificationKeys = {
  all: ["notifications"] as const,
  list: ["notifications", "list"] as const,
  count: ["notifications", "count"] as const,
};
