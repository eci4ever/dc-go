// @ts-nocheck
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Link } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Check } from "lucide-react";

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
    <main className="flex min-h-svh items-center justify-center bg-linear-to-b from-background to-muted/40 p-6">
      <section className="w-full max-w-4xl space-y-10">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
          <div className="space-y-2">
            <Badge variant="secondary" className="w-fit rounded-full px-3 py-1">
              DC Express Platform
            </Badge>
            <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
              Your services, at a glance.
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
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">Live service status</p>
              <p className="text-sm text-muted-foreground">
                Automatically refreshed every 5 seconds
              </p>
            </div>
            <Badge
              variant={isError ? "destructive" : "outline"}
              className={
                !isError && !isLoading
                  ? "gap-2 rounded-full border-0 bg-muted px-3 py-1 text-foreground"
                  : "gap-2 rounded-full px-3 py-1"
              }
            >
              {!isError && !isLoading && (
                <span className="flex size-5 items-center justify-center rounded-full bg-emerald-600 text-white">
                  <Check className="size-3.5" />
                </span>
              )}
              {isLoading ? "Checking" : isError ? "Offline" : "Operational"}
            </Badge>
          </div>
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
    <Card className="border-border/60 bg-card/80 shadow-sm transition-shadow hover:shadow-md">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
      </CardHeader>
      <CardContent className="flex items-end justify-between gap-3">
        <Badge
          variant={
            value === "Unavailable"
              ? "destructive"
              : value === "Checking…"
                ? "outline"
                : "secondary"
          }
          className={
            value !== "Unavailable" && value !== "Checking…"
              ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
              : undefined
          }
        >
          {value !== "Unavailable" && value !== "Checking…" && (
            <span className="flex size-4 items-center justify-center rounded-full bg-emerald-600 text-white">
              <Check className="size-3" />
            </span>
          )}
          {value}
        </Badge>
        <CardDescription className="text-right">Live reading</CardDescription>
      </CardContent>
    </Card>
  );
}
