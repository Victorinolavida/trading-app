import { Button } from "../components/ui/button";
import { useForm } from "react-hook-form";
import { PasswordInput } from "../components/PasswordInput";
import { useLogin } from "../features/auth/hooks";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link } from "react-router";
import { useRedirectTo } from "../hooks/useRedirectTo";

const loginSchema = z.object({
  email: z.email("Invalid email"),
  password: z
    .string()
    .min(8, "At least 8 characters")
    .max(23, "Password cannot exceed 23 characters."),
});

type LoginValues = z.infer<typeof loginSchema>;

export function LoginPage() {
  const login = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginValues>({
    defaultValues: { email: "", password: "" },
    resolver: zodResolver(loginSchema),
  });
  const { goToRedirect } = useRedirectTo();

  const onSubmit = (values: LoginValues) =>
    login.mutate(values, { onSuccess: () => goToRedirect() });

  return (
    <div className="ui-panel w-full max-w-md p-6 sm:p-8">
      <h1 className="ui-title mb-5">Sign in</h1>
      <form className="flex flex-col gap-1.5" onSubmit={handleSubmit(onSubmit)}>
        <label className="mt-2 text-sm text-fg-muted" htmlFor="email">
          Email
        </label>
        <input
          className="ui-field"
          id="email"
          type="email"
          autoComplete="email"
          {...register("email", { required: "Email is required" })}
        />
        {errors.email && (
          <p className="mt-1 text-sm text-danger">{errors.email.message}</p>
        )}

        <label className="mt-2 text-sm text-fg-muted" htmlFor="password">
          Password
        </label>
        <PasswordInput
          id="password"
          autoComplete="current-password"
          registration={register("password", {
            required: "Password is required",
          })}
        />
        {errors.password && (
          <p className="mt-1 text-sm text-danger">{errors.password.message}</p>
        )}

        <Button
          variant="primary"
          className="mt-4"
          type="submit"
          disabled={login.isPending}
        >
          {login.isPending ? "Signing in..." : "Sign in"}
        </Button>
      </form>
      <p className="mt-5 text-center text-sm text-fg-muted">
        No account?{" "}
        <Link className="ui-link" type="button" to="/register">
          Create one
        </Link>
      </p>
    </div>
  );
}
