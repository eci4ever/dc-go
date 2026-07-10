import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { Building2Icon, ChevronsUpDownIcon } from "lucide-react";
import * as api from "@/lib/api";
import type { SessionData } from "@/lib/api";
import { sessionQueryKey } from "@/lib/session";

export function OrganizationSwitcher({ session }: { session: SessionData }) {
  const { isMobile } = useSidebar();
  const queryClient = useQueryClient();
  const organizations = useQuery({
    queryKey: ["organizations"],
    queryFn: async () => {
      const response = await api.listOrganizations();
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to load organizations");
      }
      return response.data;
    },
  });
  const selectOrganization = useMutation({
    mutationFn: async (organizationId: string) => {
      const response = await api.setActiveOrganization(organizationId);
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to switch organization");
      }
      return response.data;
    },
    onSuccess: (data) => queryClient.setQueryData(sessionQueryKey, data),
  });

  const activeOrganization = organizations.data?.find(
    (organization) => organization.id === session.session.activeOrganizationId,
  );
  const role = session.session.activeOrganizationRole;

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <Building2Icon />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">
                  {activeOrganization?.name ?? "Select organization"}
                </span>
                <span className="truncate text-xs">{role ?? "No active organization"}</span>
              </div>
              <ChevronsUpDownIcon className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-60 p-2"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="px-2 py-2 text-xs text-muted-foreground">
              Organizations
            </DropdownMenuLabel>
            <DropdownMenuGroup>
              {organizations.data?.map((organization) => (
                <DropdownMenuItem
                  key={organization.id}
                  disabled={selectOrganization.isPending}
                  onSelect={() => selectOrganization.mutate(organization.id)}
                >
                  <Building2Icon />
                  <span>{organization.name}</span>
                </DropdownMenuItem>
              ))}
              {!organizations.isPending && organizations.data?.length === 0 && (
                <DropdownMenuItem disabled>No organizations</DropdownMenuItem>
              )}
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
