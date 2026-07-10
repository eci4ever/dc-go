// @ts-nocheck
import { useState } from "react";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { LoginForm } from "@/components/login-form";
import { useAuth } from "@/hooks/use-auth";
import { GalleryVerticalEnd } from "lucide-react";

export const Route = createFileRoute("/login")({ component: LoginPage });

function LoginPage() {
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    const form = new FormData(e.currentTarget);
    const err = await login(String(form.get("email") ?? ""), String(form.get("password") ?? ""));
    setSubmitting(false);
    if (err) setError(err);
    else await navigate({ to: "/dashboard" });
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-muted p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="flex items-center gap-2 self-center font-medium">
          <div className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <GalleryVerticalEnd className="size-4" />
          </div>
          DC Express
        </Link>
        <LoginForm onSubmit={handleSubmit} error={error} submitting={submitting} />
      </div>
    </div>
  );
}
