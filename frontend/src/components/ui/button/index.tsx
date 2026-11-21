import type { ComponentProps } from "react";
import { cn } from "../card";

const variants = {
  secondary:
    "border border-border-strong text-fg-soft enabled:hover:bg-surface-hover",
  primary:
    "border border-accent-strong bg-accent-strong text-white enabled:hover:border-accent enabled:hover:bg-accent-strong/80",
  dark:
    "border border-border-strong bg-surface-hover text-fg-strong enabled:hover:border-fg-subtle enabled:hover:bg-border-strong",
  danger:
    "border border-danger-border/70 bg-danger-surface/40 text-danger-soft enabled:hover:border-danger-strong enabled:hover:bg-danger-surface/60",
  ghost:
    "text-fg-muted enabled:hover:bg-surface-hover enabled:hover:text-fg-strong",
  plain: "text-inherit enabled:hover:text-fg-strong",
  link: "text-accent-soft enabled:hover:text-accent-subtle enabled:hover:underline",
  tab: "-mb-px rounded-none border-b-2 border-transparent uppercase tracking-wider text-fg-muted enabled:hover:text-fg-strong aria-selected:border-accent aria-selected:text-fg-strong",
};

const sizes = {
  default: "px-3 py-2 text-sm",
  sm: "px-2 py-1 text-xs",
  icon: "size-9",
  inline: "gap-1",
};

type ButtonStyleProps = {
  variant?: keyof typeof variants;
  size?: keyof typeof sizes;
  className?: string;
};

export function buttonStyles({
  variant = "secondary",
  size = "default",
  className,
}: ButtonStyleProps = {}) {
  return cn(
    "inline-flex cursor-pointer items-center justify-center gap-2 rounded font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-50",
    variants[variant],
    sizes[size],
    className,
  );
}

export function Button({
  variant,
  size,
  className,
  type = "button",
  ...props
}: ComponentProps<"button"> & ButtonStyleProps) {
  return (
    <button
      type={type}
      className={buttonStyles({ variant, size, className })}
      {...props}
    />
  );
}
