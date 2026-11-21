import type { ComponentProps } from "react";
import { cn } from "../card";

export function Table({
  className,
  containerClassName,
  ...props
}: ComponentProps<"table"> & { containerClassName?: string }) {
  return (
    <div className={cn("ui-panel overflow-auto", containerClassName)}>
      <table
        className={cn("w-full border-collapse text-left text-xs", className)}
        {...props}
      />
    </div>
  );
}

export function TableCaption({
  className,
  ...props
}: ComponentProps<"caption">) {
  return (
    <caption
      className={cn(
        "p-5 text-lg font-medium text-left rtl:text-right text-fg-strong",
        className,
      )}
      {...props}
    />
  );
}

export function TableHeader({ className, ...props }: ComponentProps<"thead">) {
  return (
    <thead
      className={cn(
        "border-b border-border-strong bg-surface-raised",
        className,
      )}
      {...props}
    />
  );
}

export function TableBody(props: ComponentProps<"tbody">) {
  return <tbody {...props} />;
}

export function TableRow({ className, ...props }: ComponentProps<"tr">) {
  return (
    <tr
      className={cn(
        "border-b border-border/70 transition-colors last:border-b-0 hover:bg-surface-hover/40",
        className,
      )}
      {...props}
    />
  );
}

// Column header (<th scope="col">)
export function TableHead({ className, ...props }: ComponentProps<"th">) {
  return (
    <th
      scope="col"
      className={cn(
        "whitespace-nowrap px-2 py-1.5 text-[10px] font-medium uppercase tracking-wider text-fg-muted",
        className,
      )}
      {...props}
    />
  );
}

// First cell of a row (<th scope="row">), used for the row title
export function TableRowHeader({ className, ...props }: ComponentProps<"th">) {
  return (
    <th
      scope="row"
      className={cn(
        "whitespace-nowrap px-2 py-1.5 font-medium text-fg-strong",
        className,
      )}
      {...props}
    />
  );
}

export function TableCell({ className, ...props }: ComponentProps<"td">) {
  return (
    <td className={cn("whitespace-nowrap px-2 py-1.5", className)} {...props} />
  );
}
