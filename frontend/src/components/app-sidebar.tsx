import type { ComponentProps } from "react";
import { Building2Icon, GraduationCapIcon, LayoutDashboardIcon, UsersIcon } from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavUser } from "@/components/nav-user";
import { OrganizationSwitcher } from "@/components/team-switcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
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
        <OrganizationSwitcher session={session} />
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
