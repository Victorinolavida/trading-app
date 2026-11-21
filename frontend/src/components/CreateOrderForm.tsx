import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Card, CardTitle } from "./ui/card";
import { useEffect, useRef, useState } from "react";
import { useCreateOrder } from "../features/orders/hooks";
import {
  useCalculatePrice,
  useCalculateYield,
} from "../features/calculator/hooks";

const calculationFields = ["price", "yield"] as const;
const calculationLabels = { price: "Price", yield: "Yield" };

function isPositiveNumber(value: string): boolean {
  const number = Number(value);
  return Number.isFinite(number) && number > 0;
}

const createOrderSchema = z
  .object({
    securityId: z.uuid(),
    side: z.enum(["bid", "ask"]),
    orderType: z.enum(["limit", "market"]),
    calculationField: z.enum(calculationFields),
    calculationValue: z.string(),
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

    if (
      values.orderType === "limit" &&
      !isPositiveNumber(values.calculationValue)
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["calculationValue"],
        message: `${calculationLabels[values.calculationField]} must be greater than 0`,
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

const labelClass = "mb-1 block text-xs text-fg-muted";
const inputClass = "w-full rounded bg-surface-hover p-2 text-sm";
const errorClass = "mt-1 text-sm text-danger";

export function CreateOrderForm({ securityId }: { securityId: string }) {
  const [defaultExpiration, setDefaultExpiration] = useState<boolean>(true);
  const [previewInput, setPreviewInput] = useState<{
    field: (typeof calculationFields)[number];
    value: string;
  } | null>(null);

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
      securityId,
      calculationField: "price",
      calculationValue: "",
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
      ...(values.orderType === "limit"
        ? { [values.calculationField]: values.calculationValue }
        : {}),
      quantity: values.quantity,
      expiration,
    });
    setPreviewInput(null);
    reset();
  });

  function clearCalculation() {
    setPreviewInput(null);
    setValue("calculationValue", "", { shouldValidate: true });
  }

  const orderType = watch("orderType");
  const calculationField = watch("calculationField");
  const calculationValue = watch("calculationValue");
  const derivedField = calculationField === "price" ? "yield" : "price";
  const settlementDate = new Date().toISOString().slice(0, 10) + "T00:00:00Z";

  useEffect(() => {
    if (orderType !== "limit" || !isPositiveNumber(calculationValue)) return;

    const timeout = setTimeout(() => {
      setPreviewInput({ field: calculationField, value: calculationValue });
    }, 350);
    return () => clearTimeout(timeout);
  }, [orderType, calculationField, calculationValue]);

  const canPreview =
    orderType === "limit" &&
    previewInput?.field === calculationField &&
    previewInput.value === calculationValue;
  const cleanPrice =
    canPreview && calculationField === "price" ? calculationValue : "";
  const yieldValue =
    canPreview && calculationField === "yield" ? calculationValue : "";

  const yieldQuery = useCalculateYield({
    securityId,
    settlementDate,
    cleanPrice,
    enabled: cleanPrice !== "",
  });
  const priceQuery = useCalculatePrice({
    securityId,
    settlementDate,
    yield: yieldValue,
    enabled: yieldValue !== "",
  });

  const previewQuery = calculationField === "price" ? yieldQuery : priceQuery;
  const previewValue = canPreview
    ? calculationField === "price"
      ? yieldQuery.data?.toString()
      : priceQuery.data?.clean_price
    : undefined;
  const previewError = canPreview && previewQuery.isError;

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
  }, [orderType, setValue, watch]);

  return (
    <Card>
      <CardTitle className="mb-3">Create Order</CardTitle>
      <form
        className="grid grid-cols-2 gap-3"
        onSubmit={handleCreateOrder}
        noValidate
      >
        <div>
          <label className={labelClass} htmlFor="side">
            Side
          </label>
          <select id="side" className={inputClass} {...register("side")}>
            <option value="bid">Bid</option>
            <option value="ask">Ask</option>
          </select>
          {errors.side && <p className={errorClass}>{errors.side.message}</p>}
        </div>
        <div>
          <label className={labelClass} htmlFor="orderType">
            Type
          </label>
          <select
            id="orderType"
            className={inputClass}
            {...register("orderType", { onChange: clearCalculation })}
          >
            <option value="limit">Limit</option>
            <option value="market">Market</option>
          </select>
          {errors.orderType && (
            <p className={errorClass}>{errors.orderType.message}</p>
          )}
        </div>
        {orderType === "limit" && (
          <>
            <div className="col-span-2">
              <label className={labelClass} htmlFor="calculationField">
                Enter by
              </label>
              <select
                id="calculationField"
                className={inputClass}
                {...register("calculationField", {
                  onChange: clearCalculation,
                })}
              >
                {calculationFields.map((field) => (
                  <option key={field} value={field}>
                    {calculationLabels[field]}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelClass} htmlFor="calculationValue">
                {calculationLabels[calculationField]}
              </label>
              <input
                id="calculationValue"
                type="number"
                step="any"
                min="0"
                className={inputClass}
                {...register("calculationValue")}
              />
              {errors.calculationValue && (
                <p className={errorClass}>{errors.calculationValue.message}</p>
              )}
            </div>
            <div>
              <label className={labelClass} htmlFor="calculationPreview">
                {calculationLabels[derivedField]} (calculated)
              </label>
              <input
                id="calculationPreview"
                className={`${inputClass} text-fg-muted`}
                value={previewValue ?? ""}
                placeholder={
                  canPreview && previewQuery.isFetching ? "Calculating…" : "—"
                }
                readOnly
              />
            </div>
          </>
        )}

        {previewError && (
          <p className="col-span-2 text-sm text-danger" role="alert">
            Could not preview the calculation. You can still submit the order
            using the value you entered.
          </p>
        )}

        <div>
          <label className={labelClass} htmlFor="quantity">
            Quantity
          </label>
          <input
            id="quantity"
            type="number"
            min="1"
            step="1"
            className={inputClass}
            {...register("quantity")}
          />
          {errors.quantity && (
            <p className={errorClass}>{errors.quantity.message}</p>
          )}
        </div>
        {orderType == "limit" && (
          <>
            <div className="col-span-2">
              <div className="flex items-center gap-2">
                <input
                  id="expirationEnabled"
                  type="checkbox"
                  className="h-5 w-5 cursor-pointer rounded border-border-strong bg-surface-hover text-accent-strong focus:ring-2 focus:ring-accent-strong"
                  onChange={(e) => setDefaultExpiration(e.target.checked)}
                  checked={defaultExpiration}
                />
                <label
                  className="text-sm text-fg-muted"
                  htmlFor="expirationEnabled"
                >
                  day expiration
                </label>
              </div>
            </div>
            <div className="col-span-2 grid grid-cols-2 gap-3">
              <div>
                <label className={labelClass} htmlFor="expirationValue">
                  Expires in
                </label>
                <input
                  id="expirationValue"
                  type="number"
                  min="1"
                  step="1"
                  disabled={defaultExpiration}
                  className={`${inputClass} disabled:opacity-60`}
                  {...register("expirationValue")}
                />
                {errors.expirationValue && (
                  <p className={errorClass}>{errors.expirationValue.message}</p>
                )}
              </div>
              <div>
                <label className={labelClass} htmlFor="expirationUnit">
                  Unit
                </label>
                <select
                  id="expirationUnit"
                  className={`${inputClass} disabled:opacity-60`}
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
            className="w-full cursor-pointer rounded-md bg-accent-strong px-4 py-2.5 text-base text-fg-strong disabled:opacity-60"
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
