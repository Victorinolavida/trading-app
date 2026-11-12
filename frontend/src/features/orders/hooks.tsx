import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ordersApi } from "./api";
import { toast } from "sonner";
import { getErrorMessage } from "../../lib/apiError";
import { orderKeys } from "./query";
import { CreateOrderRequest } from "./types";

export const useCreateOrder = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateOrderRequest) => ordersApi.createOrder(data),
    onError: (error) => toast.error(getErrorMessage(error)),
    onSuccess: (res) => {
      toast.success(
        `Order created: quantity ${res.quantity},, status ${res.status}`,
      );
      queryClient.invalidateQueries({
        queryKey: orderKeys.events(res.security_id),
      });
    },
  });
};

export const useCancelOrder = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (orderId: string) => ordersApi.cancelOrder(orderId),
    onError: (error) =>
      toast.error(getErrorMessage(error) ?? "Could not cancel order"),
    onSuccess: (order) => {
      toast.success("Order cancelled successfully", {
        description: order.id ? `Order ${order.id}` : undefined,
      });
      queryClient.invalidateQueries({ queryKey: orderKeys.all });
    },
  });
};
