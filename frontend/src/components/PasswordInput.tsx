import { useState } from "react";
import type { UseFormRegisterReturn } from "react-hook-form";

const inputClass =
  "w-full rounded-md border border-border bg-surface py-2.5 pl-3 pr-11 text-base text-inherit focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30";

interface PasswordInputProps {
  id: string;
  autoComplete: string;
  registration: UseFormRegisterReturn;
}

export function PasswordInput({
  id,
  autoComplete,
  registration,
}: PasswordInputProps) {
  const [visible, setVisible] = useState(false);
  const label = visible ? "Hide password" : "Show password";

  return (
    <div className="relative">
      <input
        {...registration}
        id={id}
        type={visible ? "text" : "password"}
        autoComplete={autoComplete}
        className={inputClass}
      />
      <button
        type="button"
        className="absolute inset-y-0 right-0 flex w-11 cursor-pointer items-center justify-center text-fg-muted hover:text-fg"
        onClick={() => setVisible((value) => !value)}
        aria-label={label}
        aria-pressed={visible}
        title={label}
      >
        {visible ? (
          <svg
            aria-hidden="true"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
            <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
            <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24" />
            <line x1="1" y1="1" x2="23" y2="23" />
          </svg>
        ) : (
          <svg
            aria-hidden="true"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8Z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        )}
      </button>
    </div>
  );
}
