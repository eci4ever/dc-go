import type { ComponentProps } from "react";
import { LayoutDashboardIcon, UsersIcon } from "lucide-react";

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
    title: "User management",
    url: "/admin/users" as const,
    icon: UsersIcon,
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
        {user.role === "admin" && <NavMain label="Administration" items={adminItems} />}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} onLogout={onLogout} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
