import { useForm } from "react-hook-form";
import { PasswordInput } from "../components/PasswordInput";
import { useLogin } from "../hooks/useAuth";
import { ApiError } from "../lib/api";

interface FormValues {
  email: string;
  password: string;
}

export function LoginPage({ onSwitch }: { onSwitch: () => void }) {
  const login = useLogin();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ defaultValues: { email: "", password: "" } });

  const onSubmit = (values: FormValues) => login.mutate(values);

  const serverError =
    login.error instanceof ApiError ? login.error.message : null;

  return (
    <div className="w-full max-w-sm rounded-lg border border-zinc-800 bg-zinc-900 p-8">
      <h1 className="mb-5 text-2xl font-semibold">Sign in</h1>
      <form
        className="flex flex-col gap-1.5"
        onSubmit={handleSubmit(onSubmit)}
        noValidate
      >
        <label className="mt-2 text-sm text-gray-400" htmlFor="email">
          Email
        </label>
        <input
          className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
          id="email"
          type="email"
          autoComplete="email"
          {...register("email", { required: "Email is required" })}
        />
        {errors.email && (
          <p className="mt-1 text-sm text-red-400">{errors.email.message}</p>
        )}

        <label className="mt-2 text-sm text-gray-400" htmlFor="password">
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
          <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>
        )}

        {serverError && (
          <p className="mt-1 text-sm text-red-400">{serverError}</p>
        )}

        <button
          className="mt-4 cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:cursor-default disabled:opacity-60"
          type="submit"
          disabled={login.isPending}
        >
          {login.isPending ? "Signing in..." : "Sign in"}
        </button>
      </form>
      <p className="mt-5 text-center text-sm text-gray-400">
        No account?{" "}
        <button
          className="cursor-pointer text-blue-400"
          type="button"
          onClick={onSwitch}
        >
          Create one
        </button>
      </p>
    </div>
  );
}
