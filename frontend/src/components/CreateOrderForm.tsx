import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card, CardTitle } from "./ui/card";
import { useEffect, useRef, useState } from "react";
import { useCreateOrder } from "../features/orders/hooks";

const optionalPositiveNumber = (message: string) =>
  z.preprocess(
    (value) =>
      value === "" || value === null || value === undefined ? undefined : value,
    z.coerce.number().positive(message).optional(),
  );

const createOrderSchema = z
  .object({
    securityId: z.uuid(),
    side: z.enum(["bid", "ask"]),
    orderType: z.enum(["limit", "market"]),
    price: optionalPositiveNumber("price must be greater than 0"),
    yield: optionalPositiveNumber("yield must be greater than 0"),
    quantity: z.coerce
      .number()
      .int()
      .positive("quantity must be greater than 0"),
    expirationValue: z.coerce
      .number()
      .int()
      .positive("expiration must be greater than 0")
      .optional(),
    expirationUnit: z.enum(["minute", "second", "hour"]),
  })
  .superRefine((values, ctx) => {
    if (
      values.expirationValue === undefined ||
      !Number.isInteger(values.expirationValue) ||
      values.expirationValue <= 0
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["expirationValue"],
        message: "expiration must be greater than 0",
      });
    }

    if (values.orderType === "limit" && values.price === undefined) {
      ctx.addIssue({
        code: "custom",
        path: ["price"],
        message: "price must be greater than 0",
      });
    }
  });

type CreateOrderFormInput = z.input<typeof createOrderSchema>;
type CreateOrderFormOutput = z.output<typeof createOrderSchema>;

const EXPIRATION_UNIT_SUFFIX: Record<string, string> = {
  minute: "m",
  second: "s",
  hour: "h",
};

export function CreateOrderForm({ securityId }: { securityId: string }) {
  const [defaultExpiration, setDefaultExpiration] = useState<boolean>(true);

  const previousUnit = useRef<CreateOrderFormInput["expirationUnit"] | null>(
    null,
  );
  const previousValue = useRef<CreateOrderFormInput["expirationValue"] | null>(
    null,
  );

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    watch,
    formState: { isValid, errors },
  } = useForm<CreateOrderFormInput, unknown, CreateOrderFormOutput>({
    defaultValues: {
      securityId: securityId,
      expirationValue: 24,
      expirationUnit: "hour",
    },
    resolver: zodResolver(createOrderSchema),
    mode: "onChange",
  });

  const createOrderMutation = useCreateOrder();

  const handleCreateOrder = handleSubmit((values) => {
    const expiration = `${values.expirationValue}${EXPIRATION_UNIT_SUFFIX[values.expirationUnit] ?? "24h"}`;

    createOrderMutation.mutate({
      security_id: securityId,
      direction: values.side,
      order_type: values.orderType,
      price: values.orderType === "market" ? undefined : String(values.price),
      yield: values.yield != null ? String(values.yield) : undefined,
      quantity: values.quantity,
      expiration,
    });
    reset();
  });

  const orderType = watch("orderType");

  useEffect(() => {
    if (orderType == "market") {
      const currentUnit = watch("expirationUnit");
      const currentVal = watch("expirationValue");
      previousUnit.current = currentUnit;
      previousValue.current = currentVal;
      setValue("expirationUnit", "second");
      setValue("expirationValue", 1);
      return;
    }
    if (orderType == "limit" && previousUnit.current && previousValue.current) {
      setValue("expirationUnit", previousUnit.current);
      setValue("expirationValue", previousValue.current);
      return;
    }
  }, [orderType]);

  return (
    <Card>
      <CardTitle className="mb-3">Create Order</CardTitle>
      <form
        className="grid grid-cols-2 gap-3"
        onSubmit={handleCreateOrder}
        noValidate
      >
        <div>
          <label className="mb-1 block text-xs text-gray-400" htmlFor="side">
            Side
          </label>
          <select
            id="side"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            {...register("side")}
          >
            <option value="bid">Bid</option>
            <option value="ask">Ask</option>
          </select>
          {errors.side && (
            <p className="mt-1 text-sm text-red-400">{errors.side.message}</p>
          )}
        </div>
        <div>
          <label
            className="mb-1 block text-xs text-gray-400"
            htmlFor="orderType"
          >
            Type
          </label>
          <select
            id="orderType"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            {...register("orderType")}
          >
            <option value="limit">Limit</option>
            <option value="market">Market</option>
          </select>
          {errors.orderType && (
            <p className="mt-1 text-sm text-red-400">
              {errors.orderType.message}
            </p>
          )}
        </div>
        {orderType == "limit" && (
          <>
            <div>
              <label
                className="mb-1 block text-xs text-gray-400"
                htmlFor="price"
              >
                Price
              </label>
              <input
                id="price"
                type="number"
                step="any"
                min="0"
                className="w-full rounded bg-zinc-800 p-2 text-sm"
                {...register("price")}
              />
              {errors.price && (
                <p className="mt-1 text-sm text-red-400">
                  {errors.price.message}
                </p>
              )}
            </div>
            <div>
              <label
                className="mb-1 block text-xs text-gray-400"
                htmlFor="yield"
              >
                Yield
              </label>
              <input
                id="yield"
                type="number"
                step="any"
                min="0"
                className="w-full rounded bg-zinc-800 p-2 text-sm"
                {...register("yield")}
              />
              {errors.yield && (
                <p className="mt-1 text-sm text-red-400">
                  {errors.yield.message}
                </p>
              )}
            </div>
          </>
        )}

        <div>
          <label
            className="mb-1 block text-xs text-gray-400"
            htmlFor="quantity"
          >
            Quantity
          </label>
          <input
            id="quantity"
            type="number"
            min="1"
            step="1"
            className="w-full rounded bg-zinc-800 p-2 text-sm"
            {...register("quantity")}
          />
          {errors.quantity && (
            <p className="mt-1 text-sm text-red-400">
              {errors.quantity.message}
            </p>
          )}
        </div>
        {orderType == "limit" && (
          <>
            <div className="col-span-2">
              <div className="flex items-center gap-2">
                <input
                  id="expirationEnabled"
                  type="checkbox"
                  className="h-5 w-5 cursor-pointer rounded border-zinc-700 bg-zinc-800 text-blue-600 focus:ring-2 focus:ring-blue-600"
                  onChange={(e) => setDefaultExpiration(e.target.checked)}
                  checked={defaultExpiration}
                />
                <label
                  className="text-sm text-gray-400"
                  htmlFor="expirationEnabled"
                >
                  day expiration
                </label>
              </div>
            </div>
            <div className="col-span-2 grid grid-cols-2 gap-3">
              <div>
                <label
                  className="mb-1 block text-xs text-gray-400"
                  htmlFor="expirationValue"
                >
                  Expires in
                </label>
                <input
                  id="expirationValue"
                  type="number"
                  min="1"
                  step="1"
                  disabled={defaultExpiration}
                  className="w-full rounded bg-zinc-800 p-2 text-sm disabled:opacity-60"
                  {...register("expirationValue")}
                />
                {errors.expirationValue && (
                  <p className="mt-1 text-sm text-red-400">
                    {errors.expirationValue.message}
                  </p>
                )}
              </div>
              <div>
                <label
                  className="mb-1 block text-xs text-gray-400"
                  htmlFor="expirationUnit"
                >
                  Unit
                </label>
                <select
                  id="expirationUnit"
                  className="w-full rounded bg-zinc-800 p-2 text-sm disabled:opacity-60"
                  {...register("expirationUnit")}
                  disabled={defaultExpiration}
                >
                  <option value="minute">Minute</option>
                  <option value="second">Second</option>
                  <option value="hour">Hour</option>
                </select>
              </div>
            </div>
          </>
        )}

        <div className="col-span-2">
          <button
            className="w-full cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:opacity-60"
            type="submit"
            disabled={createOrderMutation.isPending || !isValid}
          >
            {createOrderMutation.isPending ? "Creating..." : "Create Order"}
          </button>
        </div>
      </form>
    </Card>
  );
}
