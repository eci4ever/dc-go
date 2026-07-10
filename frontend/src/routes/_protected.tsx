import { createFileRoute, Outlet, redirect, useLocation } from "@tanstack/react-router";
import { AppSidebar } from "@/components/app-sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import { sessionQueryOptions } from "@/lib/session";

export const Route = createFileRoute("/_protected")({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(sessionQueryOptions);
    if (!session) throw redirect({ to: "/login", replace: true });
  },
  component: ProtectedLayout,
});

function ProtectedLayout() {
  const { user, session, logout } = useAuth();
  const pathname = useLocation({ select: (location) => location.pathname });
  if (!user || !session) return null;

  const isUserManagement = pathname.startsWith("/admin/users");

  return (
    <SidebarProvider>
      <AppSidebar user={user} session={session} onLogout={logout} />
      <SidebarInset>
        <div className="sticky top-0 z-10 shrink-0 bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/80">
          <header className="flex h-14 items-center gap-3 px-4 md:px-6">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="data-vertical:h-4 data-vertical:self-center"
            />
            <Breadcrumb>
              <BreadcrumbList>
                {isUserManagement && (
                  <>
                    <BreadcrumbItem className="hidden md:inline-flex">
                      <span>Administration</span>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator className="hidden md:list-item" />
                  </>
                )}
                <BreadcrumbItem>
                  <BreadcrumbPage>
                    {isUserManagement ? "Users" : "Dashboard"}
                  </BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </header>
          <Separator />
        </div>
        <main className="flex flex-1 flex-col p-4 md:p-6">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
