import type { ComponentProps, ReactNode } from "react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs));

export function Card({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "mb-6 rounded-lg border border-border bg-surface-raised p-6",
        className,
      )}
      {...props}
    />
  );
}

export function CardTitle({ className, ...props }: ComponentProps<"h1">) {
  return (
    <h1 className={cn("mb-2 text-2xl font-semibold", className)} {...props} />
  );
}

export function CardSubtitle({ className, ...props }: ComponentProps<"p">) {
  return <p className={cn("mb-4 text-fg-muted", className)} {...props} />;
}

export function DetailList({ className, ...props }: ComponentProps<"dl">) {
  return (
    <dl
      className={cn("grid grid-cols-2 gap-4 text-sm", className)}
      {...props}
    />
  );
}

export function Section({
  title,
  className,
  children,
}: {
  title: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <section
      className={cn(
        "border-t border-border pt-5 mt-5 first:border-t-0 first:pt-0 first:mt-0",
        className,
      )}
    >
      <h2 className="mb-3 text-xs font-medium uppercase tracking-wider text-fg-subtle">
        {title}
      </h2>
      <dl className="grid grid-cols-1 gap-x-8 gap-y-4 sm:grid-cols-2">
        {children}
      </dl>
    </section>
  );
}

export function DetailItem({
  label,
  children,
  mono,
  className,
}: {
  label: string;
  children: ReactNode;
  mono?: boolean;
  className?: string;
}) {
  return (
    <div className={className}>
      <dt className="text-xs text-fg-muted">{label}</dt>
      <dd
        className={cn(
          "mt-0.5 text-sm text-fg-strong break-all",
          mono && "font-mono",
        )}
      >
        {children}
      </dd>
    </div>
  );
}

export function Badge({ active }: { active: boolean }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        active
          ? "bg-success/15 text-success"
          : "bg-surface-hover/40 text-fg-muted",
      )}
    >
      {active ? "Enabled" : "Disabled"}
    </span>
  );
}
