import { createFileRoute, Outlet, redirect, useLocation } from "@tanstack/react-router";
import { AppSidebar } from "@/components/app-sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
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
  const isOrganizationManagement = pathname.startsWith("/admin/organizations");
  const isAccount = pathname.startsWith("/account");
  const isAcademic = pathname.startsWith("/academic");
  const pageName = isOrganizationManagement
    ? "Organizations"
    : isUserManagement
      ? "Users"
      : isAccount
        ? "Account"
        : isAcademic
          ? "Academic"
          : "Dashboard";

  return (
    <SidebarProvider className="h-svh min-h-0 overflow-hidden">
      <AppSidebar user={user} session={session} onLogout={logout} />
      <SidebarInset className="h-svh min-w-0 overflow-hidden">
        <header className="flex h-16 shrink-0 items-center gap-2 bg-muted/40 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex min-w-0 items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-vertical:h-4 data-vertical:self-center"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbPage>{pageName}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex min-h-0 min-w-0 max-w-full flex-1 flex-col gap-4 overflow-x-hidden overflow-y-auto overscroll-y-contain p-4">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
