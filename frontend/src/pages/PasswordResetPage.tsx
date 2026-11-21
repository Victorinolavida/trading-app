import { useEffect, useState } from "react";
import { Link } from "react-router";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "../components/ui/button";
import { PasswordInput } from "../components/PasswordInput";
import { authApi } from "../features/auth/api";
import { getErrorMessage } from "../lib/apiError";

const emailSchema = z.object({ email: z.email("Invalid email") });
const passwordSchema = z
  .object({
    password: z
      .string()
      .min(8, "At least 8 characters")
      .max(23, "Password cannot exceed 23 characters."),
    confirmPassword: z.string().min(1, "Please confirm your password"),
  })
  .refine((values) => values.password === values.confirmPassword, {
    path: ["confirmPassword"],
    message: "Passwords do not match",
  });

export function ForgotPasswordPage() {
  const request = useMutation({ mutationFn: authApi.requestPasswordReset });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<z.infer<typeof emailSchema>>({
    defaultValues: { email: "" },
    resolver: zodResolver(emailSchema),
  });
  return (
    <div className="ui-panel w-full max-w-md p-6 sm:p-8">
      <h1 className="ui-title mb-5">Reset password</h1>
      {request.isSuccess ? (
        <p className="text-sm text-fg-muted" role="status">
          If an account exists for that email, you’ll receive a reset link.
        </p>
      ) : (
        <form
          className="flex flex-col gap-1.5"
          onSubmit={handleSubmit(({ email }) => request.mutate(email))}
          noValidate
        >
          <label className="mt-2 text-sm text-fg-muted" htmlFor="email">
            Email
          </label>
          <input
            className="ui-field"
            id="email"
            type="email"
            autoComplete="email"
            {...register("email")}
          />
          {errors.email && (
            <p className="mt-1 text-sm text-danger">{errors.email.message}</p>
          )}
          {request.error && (
            <p className="mt-1 text-sm text-danger" role="alert">
              {getErrorMessage(request.error)}
            </p>
          )}
          <Button
            variant="primary"
            className="mt-4"
            type="submit"
            disabled={request.isPending}
          >
            {request.isPending ? "Sending..." : "Send reset link"}
          </Button>
        </form>
      )}
      <p className="mt-5 text-center text-sm">
        <Link className="ui-link" to="/login">
          Back to sign in
        </Link>
      </p>
    </div>
  );
}

export function ResetPasswordPage() {
  const [token] = useState(
    () => new URLSearchParams(window.location.search).get("token") ?? "",
  );
  const queryClient = useQueryClient();
  useEffect(() => {
    const url = new URL(window.location.href);
    url.searchParams.delete("token");
    window.history.replaceState(
      window.history.state,
      "",
      url.pathname + url.search + url.hash,
    );
  }, []);
  const reset = useMutation({
    mutationFn: (password: string) =>
      authApi.confirmPasswordReset(token, password),
    onSuccess: () => queryClient.clear(),
  });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<z.infer<typeof passwordSchema>>({
    defaultValues: { password: "", confirmPassword: "" },
    resolver: zodResolver(passwordSchema),
  });
  return (
    <div className="ui-panel w-full max-w-md p-6 sm:p-8">
      <h1 className="ui-title mb-5">Choose a new password</h1>
      {reset.isSuccess ? (
        <p className="text-sm text-fg-muted" role="status">
          Your password has been changed. Sign in with your new password.
        </p>
      ) : !token ? (
        <p className="text-sm text-danger" role="alert">
          This reset link is missing a token. Request a new link.
        </p>
      ) : (
        <form
          className="flex flex-col gap-1.5"
          onSubmit={handleSubmit(({ password }) => reset.mutate(password))}
          noValidate
        >
          <label className="mt-2 text-sm text-fg-muted" htmlFor="password">
            New password
          </label>
          <PasswordInput
            id="password"
            autoComplete="new-password"
            registration={register("password")}
          />
          {errors.password && (
            <p className="mt-1 text-sm text-danger">
              {errors.password.message}
            </p>
          )}
          <label
            className="mt-2 text-sm text-fg-muted"
            htmlFor="confirmPassword"
          >
            Confirm password
          </label>
          <PasswordInput
            id="confirmPassword"
            autoComplete="new-password"
            registration={register("confirmPassword")}
          />
          {errors.confirmPassword && (
            <p className="mt-1 text-sm text-danger">
              {errors.confirmPassword.message}
            </p>
          )}
          {reset.error && (
            <p className="mt-1 text-sm text-danger" role="alert">
              {getErrorMessage(reset.error)}
            </p>
          )}
          <Button
            variant="primary"
            className="mt-4"
            type="submit"
            disabled={reset.isPending}
          >
            {reset.isPending ? "Saving..." : "Change password"}
          </Button>
        </form>
      )}
      <p className="mt-5 text-center text-sm">
        <Link
          className="ui-link"
          to={reset.isSuccess ? "/login" : "/forgot-password"}
        >
          {reset.isSuccess ? "Sign in" : "Request a new link"}
        </Link>
      </p>
    </div>
  );
}
