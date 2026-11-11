import { useForm } from "react-hook-form";
import { PasswordInput } from "../components/PasswordInput";
import { useRegister } from "../features/auth/hooks";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

const registerSchema = z.object({
  email: z.email("Invalid email"),
  password: z.string().min(8, "At least 8 characters").max(23, "Password cannot exceed 23 characters."),
  firstName: z
  .string()
  .min(3, "First name must be at least 3 characters long.")
  .max(50, "First name cannot exceed 50 characters."),
  lastName: z
  .string()
  .min(3, "First name must be at least 3 characters long.")
  .max(50, "First name cannot exceed 50 characters.")
});


type RegisterValues = z.infer<typeof registerSchema>;

export function RegisterPage({ onSwitch }: { onSwitch: () => void }) {
  const registerMutation = useRegister();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterValues>(
    { 
      defaultValues: { email: "", firstName: "", lastName: "", password: "" },
      resolver: zodResolver(registerSchema)
    }
  );

  const onSubmit = (values: RegisterValues) => registerMutation.mutate(values);

  return (
    <div className="w-full max-w-sm rounded-lg border border-zinc-800 bg-zinc-900 p-8">
      <h1 className="mb-5 text-2xl font-semibold">Create account</h1>
      <form className="flex flex-col gap-1.5" onSubmit={handleSubmit(onSubmit)} noValidate>
        <label className="mt-2 text-sm text-gray-400" htmlFor="email">Email</label>
        <input
          className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
          id="email"
          type="email"
          autoComplete="email"
          {...register("email", {
            required: "Email is required",
          })}
        />
        {errors.email && <p className="mt-1 text-sm text-red-400">{errors.email.message}</p>}

        <label className="mt-2 text-sm text-gray-400" htmlFor="firstName">First name</label>
        <input
          className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
          id="firstName"
          autoComplete="given-name"
          {...register("firstName", { required: "First name is required" })}
        />
        {errors.firstName && <p className="mt-1 text-sm text-red-400">{errors.firstName.message}</p>}

        <label className="mt-2 text-sm text-gray-400" htmlFor="lastName">Last name</label>
        <input
          className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2.5 text-base text-inherit focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
          id="lastName"
          autoComplete="family-name"
          {...register("lastName", { required: "Last name is required" })}
        />
        {errors.lastName && <p className="mt-1 text-sm text-red-400">{errors.lastName.message}</p>}

        <label className="mt-2 text-sm text-gray-400" htmlFor="password">Password</label>
        <PasswordInput
          id="password"
          autoComplete="new-password"
          registration={register("password", { required: "Password is required" })}
        />
        {errors.password && <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>}

        <button
          className="mt-4 cursor-pointer rounded-md bg-blue-600 px-4 py-2.5 text-base text-white disabled:cursor-default disabled:opacity-60"
          type="submit"
          disabled={registerMutation.isPending}
        >
          {registerMutation.isPending ? "Creating..." : "Create account"}
        </button>
      </form>
      <p className="mt-5 text-center text-sm text-gray-400">
        Already registered?{" "}
        <button className="cursor-pointer text-blue-400" type="button" onClick={onSwitch}>Sign in</button>
      </p>
    </div>
  );
}

