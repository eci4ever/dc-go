// @ts-nocheck
import { useState } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { SignupForm } from "@/components/signup-form";
import { useAuth } from "@/hooks/use-auth";
import { GalleryVerticalEnd } from "lucide-react";

export const Route = createFileRoute("/signup")({ component: SignupPage });

function SignupPage() {
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const { register } = useAuth();

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    const form = new FormData(e.currentTarget);
    if (form.get("password") !== form.get("confirm-password")) {
      setError("Passwords do not match");
      return;
    }
    setSubmitting(true);
    const err = await register(
      String(form.get("name") ?? ""),
      String(form.get("email") ?? ""),
      String(form.get("password") ?? ""),
    );
    setSubmitting(false);
    if (err) setError(err);
    else window.location.href = "/";
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
        <SignupForm onSubmit={handleSubmit} error={error} submitting={submitting} />
      </div>
    </div>
  );
}
