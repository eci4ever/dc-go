// @ts-nocheck
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Link } from "@tanstack/react-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface HealthResponse {
  success: boolean;
  data?: {
    status: string;
    db: string;
    latency: number;
  };
}

export const Route = createFileRoute("/")({ component: LandingPage });

function LandingPage() {
  const { data, isLoading, isError } = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/health").then((response) => response.json()),
    refetchInterval: 5000,
  });

  const health = data?.data;
  const value = (current: string | number | undefined) =>
    isLoading ? "Checking…" : isError ? "Unavailable" : String(current ?? "Unknown");

  return (
    <main className="flex min-h-svh items-center justify-center bg-gradient-to-b from-background to-muted/40 p-6">
      <section className="w-full max-w-4xl space-y-10">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium uppercase tracking-[0.2em] text-primary">
              DC Express
            </p>
            <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
              Everything is running smoothly.
            </h1>
            <p className="max-w-xl text-muted-foreground">
              A simple, reliable foundation for your next project.
            </p>
          </div>
          <div className="flex gap-3">
            <Button variant="outline" asChild>
              <Link to="/login">Log in</Link>
            </Button>
            <Button asChild>
              <Link to="/signup">Create account</Link>
            </Button>
          </div>
        </div>
        <div className="space-y-3">
          <p className="text-sm font-medium text-muted-foreground">Live service status</p>
          <div className="grid gap-4 md:grid-cols-3">
            <StatusCard label="API status" value={value(health?.status)} />
            <StatusCard label="Database status" value={value(health?.db)} />
            <StatusCard
              label="Database latency"
              value={isLoading || isError ? value(undefined) : `${health?.latency ?? "—"} ms`}
            />
          </div>
        </div>
      </section>
    </main>
  );
}

function StatusCard({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <span
            className={`size-2 rounded-full ${value === "Unavailable" ? "bg-destructive" : "bg-emerald-500"}`}
          />
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold capitalize">{value}</p>
      </CardContent>
    </Card>
  );
}
