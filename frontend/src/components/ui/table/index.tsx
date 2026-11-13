import type { ComponentProps } from "react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs));

export function Table({ className, ...props }: ComponentProps<"table">) {
  return (
    <div className="relative overflow-x-auto bg-surface-raised shadow-xs rounded-base border border-border">
      <table
        className={cn(
          "w-full text-sm text-left rtl:text-right text-fg",
          className,
        )}
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
        "text-sm text-fg bg-surface-hover border-y border-border-strong",
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
        "bg-surface-raised border-b border-border last:border-b-0",
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
      className={cn("px-6 py-3 font-medium", className)}
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
        "px-6 py-4 font-medium text-fg-strong whitespace-nowrap",
        className,
      )}
      {...props}
    />
  );
}

export function TableCell({ className, ...props }: ComponentProps<"td">) {
  return <td className={cn("px-6 py-4", className)} {...props} />;
}
