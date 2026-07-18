import type { ComponentProps } from "react";
import { Link } from "@tanstack/react-router";
import { Building2Icon, GraduationCapIcon, LayoutDashboardIcon, UsersIcon } from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavUser } from "@/components/nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import type { SessionData, User } from "@/lib/api";

interface AppSidebarProps extends ComponentProps<typeof Sidebar> {
  user: User;
  session: SessionData;
  onLogout: () => Promise<void>;
}

const overviewItems = [
  {
    title: "Dashboard",
    url: "/dashboard" as const,
    icon: LayoutDashboardIcon,
  },
];

const adminItems = [
  {
    title: "Organizations",
    url: "/admin/organizations" as const,
    icon: Building2Icon,
  },
  {
    title: "Users",
    url: "/admin/users" as const,
    icon: UsersIcon,
  },
];

const academicItems = [
  {
    title: "Academic",
    url: "/academic" as const,
    icon: GraduationCapIcon,
  },
];

export function AppSidebar({ user, session, onLogout, ...props }: AppSidebarProps) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" tooltip="ADTEC JTM" asChild>
              <Link to="/dashboard" aria-label="ADTEC JTM dashboard">
                <div className="flex aspect-square size-8 shrink-0 items-center justify-center rounded-md bg-white p-1">
                  <img
                    src="/branding/tms-mark.png"
                    alt="TMS"
                    className="h-auto w-full object-contain"
                  />
                </div>
                <div className="grid min-w-0 flex-1 text-left leading-tight">
                  <span className="truncate font-semibold">ADTEC JTM</span>
                  <span className="truncate text-xs text-muted-foreground">
                    Training Management System
                  </span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain label="Overview" items={overviewItems} />
        {session.session.activeOrganizationId &&
          ["owner", "admin"].includes(session.session.activeOrganizationRole ?? "") && (
            <NavMain label="Institute" items={academicItems} />
          )}
        {user.role === "admin" && <NavMain label="Administration" items={adminItems} />}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} onLogout={onLogout} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
