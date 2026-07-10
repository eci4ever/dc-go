import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { sessionQueryOptions } from "@/lib/session";

export const Route = createFileRoute("/_protected")({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(sessionQueryOptions);
    if (!session) throw redirect({ to: "/login", replace: true });
  },
  component: Outlet,
});
