import type { ComponentProps } from "react";
import { useQuery } from "@tanstack/react-query";
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
import * as api from "@/lib/api";

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

const organizationItem = {
  title: "Organization",
  url: "/organization" as const,
  icon: Building2Icon,
};

const academicItem = {
  title: "Academic",
  url: "/academic" as const,
  icon: GraduationCapIcon,
};

export function AppSidebar({ user, session, onLogout, ...props }: AppSidebarProps) {
  const ownedOrganizations = useQuery({
    queryKey: ["organizations", "owned"],
    queryFn: async () => {
      const response = await api.listOwnedOrganizations();
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to load owned organizations");
      }
      return response.data;
    },
  });
  const ownsOrganization = (ownedOrganizations.data?.length ?? 0) > 0;
  const activeMembership = useQuery({
    queryKey: ["organization", session.session.activeOrganizationId, "member", "me"],
    enabled: Boolean(session.session.activeOrganizationId),
    queryFn: async () => {
      const response = await api.getCurrentOrganizationMember(
        session.session.activeOrganizationId!,
      );
      if (!response.success || !response.data) return null;
      return response.data;
    },
  });
  const hasAcademicPermission = activeMembership.data?.permissions.some((permission) =>
    permission.startsWith("academic."),
  );
  const canManageAcademic =
    Boolean(session.session.activeOrganizationId) &&
    (["owner", "admin"].includes(session.session.activeOrganizationRole ?? "") ||
      hasAcademicPermission);
  const instituteItems = [
    ...(ownsOrganization ? [organizationItem] : []),
    ...(canManageAcademic ? [academicItem] : []),
  ];

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
        {instituteItems.length > 0 && <NavMain label="Institute" items={instituteItems} />}
        {user.role === "admin" && <NavMain label="Administration" items={adminItems} />}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} onLogout={onLogout} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
