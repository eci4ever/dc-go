import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  AlertCircleIcon,
  Building2Icon,
  CheckCircle2Icon,
  HistoryIcon,
  KeyRoundIcon,
  MailPlusIcon,
  SearchIcon,
  ShieldCheckIcon,
  Trash2Icon,
  UploadIcon,
  UsersIcon,
} from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuth } from "@/hooks/use-auth";
import * as api from "@/lib/api";
import { sessionQueryKey } from "@/lib/session";

export const Route = createFileRoute("/_protected/organization")({
  component: OrganizationManagementPage,
});

type ManagedRole = "admin" | "member";

const permissionOptions: Array<{
  value: api.OrganizationPermission;
  label: string;
  description: string;
}> = [
  {
    value: "members.manage",
    label: "Members",
    description: "Invite members and cancel invitations",
  },
  {
    value: "academic.students.manage",
    label: "Students",
    description: "View and create student records",
  },
  {
    value: "academic.structure.manage",
    label: "Academic structure",
    description: "Manage semesters, courses, and grade scale",
  },
  {
    value: "academic.results.manage",
    label: "Results",
    description: "Enter results and view transcripts",
  },
  {
    value: "audit.view",
    label: "Audit trail",
    description: "Review organization administrative activity",
  },
];

function OrganizationManagementPage() {
  const { session } = useAuth();
  const queryClient = useQueryClient();
  const activeOrganizationId = session?.session.activeOrganizationId;
  const [selectedOrganizationId, setSelectedOrganizationId] = useState<string | null>(null);
  const ownedOrganizations = useQuery({
    queryKey: ["organizations", "owned"],
    queryFn: () => load(api.listOwnedOrganizations()),
  });
  const defaultOrganizationId = ownedOrganizations.data?.some(
    (item) => item.id === activeOrganizationId,
  )
    ? activeOrganizationId
    : ownedOrganizations.data?.[0]?.id;
  const organizationId = selectedOrganizationId ?? defaultOrganizationId;
  const organizationKey = ["organization", organizationId] as const;
  const membersKey = ["organization", organizationId, "members"] as const;
  const invitationsKey = ["organization", organizationId, "invitations"] as const;
  const auditKey = ["organization", organizationId, "audit-logs"] as const;

  const [details, setDetails] = useState({ name: "", slug: "" });
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [memberSearch, setMemberSearch] = useState("");
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteForm, setInviteForm] = useState<{ email: string; role: ManagedRole }>({
    email: "",
    role: "member",
  });
  const [memberToRemove, setMemberToRemove] = useState<api.OrganizationMember | null>(null);
  const [permissionMember, setPermissionMember] = useState<api.OrganizationMember | null>(null);
  const [selectedPermissions, setSelectedPermissions] = useState<api.OrganizationPermission[]>([]);
  const [notice, setNotice] = useState<string | null>(null);

  const organization = useQuery({
    queryKey: organizationKey,
    enabled: Boolean(organizationId),
    queryFn: () => load(api.getOrganization(organizationId!)),
  });
  const members = useQuery({
    queryKey: membersKey,
    enabled: Boolean(organizationId),
    queryFn: () => load(api.listOrganizationMembers(organizationId!)),
  });
  const invitations = useQuery({
    queryKey: invitationsKey,
    enabled: Boolean(organizationId),
    queryFn: () => load(api.listOrganizationInvitations(organizationId!)),
  });
  const auditLogs = useQuery({
    queryKey: auditKey,
    enabled: Boolean(organizationId),
    queryFn: () => load(api.listOrganizationAuditLogs(organizationId!)),
  });

  useEffect(() => {
    if (organization.data) {
      setDetails({ name: organization.data.name, slug: organization.data.slug });
    }
  }, [organization.data]);

  useEffect(() => {
    if (!logoFile) {
      setLogoPreview(null);
      return;
    }
    const preview = URL.createObjectURL(logoFile);
    setLogoPreview(preview);
    return () => URL.revokeObjectURL(preview);
  }, [logoFile]);

  const updateDetails = useMutation({
    mutationFn: () =>
      load(api.updateOrganization(organizationId!, details.name.trim(), details.slug.trim())),
    onSuccess: (saved) => {
      queryClient.setQueryData(organizationKey, saved);
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      queryClient.invalidateQueries({ queryKey: ["admin", "organizations"] });
      queryClient.invalidateQueries({ queryKey: auditKey });
      setNotice("Organization details updated successfully.");
    },
  });
  const uploadLogo = useMutation({
    mutationFn: () => load(api.uploadOrganizationLogo(organizationId!, logoFile!)),
    onSuccess: (saved) => {
      queryClient.setQueryData(organizationKey, saved);
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      queryClient.invalidateQueries({ queryKey: ["admin", "organizations"] });
      queryClient.invalidateQueries({ queryKey: auditKey });
      setLogoFile(null);
      setNotice("Organization logo updated successfully.");
    },
  });
  const changeRole = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: ManagedRole }) =>
      complete(api.updateOrganizationMemberRole(organizationId!, userId, role)),
    onSuccess: (_, variables) => {
      queryClient.setQueryData<api.OrganizationMember[]>(membersKey, (current = []) =>
        current.map((member) =>
          member.user_id === variables.userId ? { ...member, role: variables.role } : member,
        ),
      );
      queryClient.invalidateQueries({ queryKey: auditKey });
      setNotice("Member role updated successfully.");
    },
  });
  const removeMember = useMutation({
    mutationFn: (userId: string) => complete(api.removeOrganizationMember(organizationId!, userId)),
    onSuccess: (_, userId) => {
      queryClient.setQueryData<api.OrganizationMember[]>(membersKey, (current = []) =>
        current.filter((member) => member.user_id !== userId),
      );
      setMemberToRemove(null);
      queryClient.invalidateQueries({ queryKey: auditKey });
      setNotice("Member removed from the organization.");
    },
  });
  const sendInvitation = useMutation({
    mutationFn: () =>
      load(api.inviteOrganizationMember(organizationId!, inviteForm.email.trim(), inviteForm.role)),
    onSuccess: (invitation) => {
      queryClient.setQueryData<api.OrganizationInvitation[]>(invitationsKey, (current = []) => [
        invitation,
        ...current,
      ]);
      setInviteOpen(false);
      setInviteForm({ email: "", role: "member" });
      setNotice(`Invitation sent to ${invitation.email}.`);
      queryClient.invalidateQueries({ queryKey: auditKey });
    },
  });
  const cancelInvitation = useMutation({
    mutationFn: (id: string) => complete(api.cancelOrganizationInvitation(id)),
    onSuccess: (_, id) => {
      queryClient.setQueryData<api.OrganizationInvitation[]>(invitationsKey, (current = []) =>
        current.filter((invitation) => invitation.id !== id),
      );
      setNotice("Invitation cancelled.");
      queryClient.invalidateQueries({ queryKey: auditKey });
    },
  });
  const updatePermissions = useMutation({
    mutationFn: () =>
      load(
        api.updateOrganizationMemberPermissions(
          organizationId!,
          permissionMember!.user_id,
          selectedPermissions,
        ),
      ),
    onSuccess: (updated) => {
      queryClient.setQueryData<api.OrganizationMember[]>(membersKey, (current = []) =>
        current.map((member) =>
          member.user_id === updated.user_id
            ? { ...member, permissions: updated.permissions }
            : member,
        ),
      );
      queryClient.invalidateQueries({ queryKey: auditKey });
      setPermissionMember(null);
      setNotice("Member permissions updated successfully.");
    },
  });
  const setActiveOrganization = useMutation({
    mutationFn: () => load(api.setActiveOrganization(organizationId!)),
    onSuccess: (updatedSession) => {
      queryClient.setQueryData(sessionQueryKey, updatedSession);
      setNotice(`${organization.data?.name ?? "Organization"} is now active.`);
    },
  });

  const filteredMembers = useMemo(() => {
    const term = memberSearch.trim().toLowerCase();
    if (!term) return members.data ?? [];
    return (members.data ?? []).filter(
      (member) =>
        member.user.name.toLowerCase().includes(term) ||
        member.user.email.toLowerCase().includes(term) ||
        member.role.includes(term),
    );
  }, [memberSearch, members.data]);

  if (ownedOrganizations.isPending) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-32 w-full" />
        <div className="grid gap-4 sm:grid-cols-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      </div>
    );
  }

  if (ownedOrganizations.isError) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>Unable to verify organization ownership</AlertTitle>
        <AlertDescription>{ownedOrganizations.error.message}</AlertDescription>
      </Alert>
    );
  }

  if (!organizationId) {
    return (
      <Empty className="flex-1">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Building2Icon />
          </EmptyMedia>
          <EmptyTitle>No owned organization</EmptyTitle>
          <EmptyDescription>
            A platform admin must assign you as an organization owner before you can manage it.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  const queryError = organization.error ?? members.error ?? invitations.error ?? auditLogs.error;
  const mutationError =
    updateDetails.error ??
    uploadLogo.error ??
    changeRole.error ??
    removeMember.error ??
    sendInvitation.error ??
    cancelInvitation.error ??
    updatePermissions.error ??
    setActiveOrganization.error;
  const pendingInvitations =
    invitations.data?.filter((invitation) => invitation.status === "pending").length ?? 0;
  const admins = members.data?.filter((member) => member.role === "admin").length ?? 0;
  const organizationLocked = organization.data?.status !== "active";

  return (
    <div className="flex w-full min-w-0 max-w-full flex-1 flex-col gap-4">
      <Card className="min-w-0">
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center">
            <Avatar className="size-16 rounded-xl" size="lg">
              <AvatarImage
                src={organization.data?.logo ?? undefined}
                alt={organization.data?.name ?? "Organization"}
                className="object-contain"
              />
              <AvatarFallback className="rounded-xl">
                {initials(organization.data?.name ?? "Organization")}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              {organization.isPending ? (
                <div className="flex flex-col gap-2">
                  <Skeleton className="h-7 w-56 max-w-full" />
                  <Skeleton className="h-4 w-72 max-w-full" />
                </div>
              ) : (
                <>
                  <div className="flex flex-wrap items-center gap-2">
                    <CardTitle className="truncate text-xl sm:text-2xl">
                      {organization.data?.name}
                    </CardTitle>
                    <Badge variant="secondary">
                      <ShieldCheckIcon data-icon="inline-start" />
                      Owner
                    </Badge>
                    <Badge
                      variant={organization.data?.status === "active" ? "outline" : "destructive"}
                    >
                      {capitalize(organization.data?.status ?? "active")}
                    </Badge>
                  </div>
                  <CardDescription className="mt-1">
                    Manage your organization profile, members, roles, and invitations.
                  </CardDescription>
                </>
              )}
            </div>
            <div className="flex shrink-0 flex-col gap-2 sm:items-end">
              {(ownedOrganizations.data?.length ?? 0) > 1 && (
                <Select
                  value={organizationId}
                  onValueChange={(value) => {
                    setSelectedOrganizationId(value);
                    setNotice(null);
                  }}
                >
                  <SelectTrigger className="w-full sm:w-64" aria-label="Organization to manage">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {ownedOrganizations.data?.map((item) => (
                        <SelectItem key={item.id} value={item.id}>
                          {item.name}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
              )}
              {activeOrganizationId !== organizationId ? (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={setActiveOrganization.isPending}
                  onClick={() => {
                    setNotice(null);
                    setActiveOrganization.mutate();
                  }}
                >
                  {setActiveOrganization.isPending && <Spinner data-icon="inline-start" />}
                  Set as active
                </Button>
              ) : (
                <Badge variant="outline">Active organization</Badge>
              )}
            </div>
          </div>
        </CardHeader>
      </Card>

      <div className="grid min-w-0 gap-4 sm:grid-cols-3">
        <SummaryCard label="Members" value={members.data?.length ?? 0} icon={UsersIcon} />
        <SummaryCard label="Administrators" value={admins} icon={ShieldCheckIcon} />
        <SummaryCard label="Pending invitations" value={pendingInvitations} icon={MailPlusIcon} />
      </div>

      {organizationLocked && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Organization is {organization.data?.status}</AlertTitle>
          <AlertDescription>
            Records remain available for review, but administrative and academic changes are locked
            until a platform admin restores Active status.
          </AlertDescription>
        </Alert>
      )}

      {notice && (
        <Alert>
          <CheckCircle2Icon />
          <AlertTitle>Changes saved</AlertTitle>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {(queryError || mutationError) && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Unable to complete the request</AlertTitle>
          <AlertDescription>{(queryError ?? mutationError)?.message}</AlertDescription>
        </Alert>
      )}

      <Tabs defaultValue="overview" className="min-w-0 gap-4">
        <TabsList className="grid w-full grid-cols-4 sm:w-fit">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="members">Members</TabsTrigger>
          <TabsTrigger value="invitations">Invitations</TabsTrigger>
          <TabsTrigger value="audit">Audit</TabsTrigger>
        </TabsList>

        <TabsContent value="overview">
          <div className="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,0.7fr)_minmax(0,1.3fr)]">
            <Card className="min-w-0">
              <CardHeader>
                <CardTitle>Organization logo</CardTitle>
                <CardDescription>Use a square JPEG or PNG image up to 2 MB.</CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-4">
                <div className="flex items-center gap-4">
                  <Avatar className="size-20 rounded-xl" size="lg">
                    <AvatarImage
                      src={logoPreview ?? organization.data?.logo ?? undefined}
                      alt="Organization logo preview"
                      className="object-contain"
                    />
                    <AvatarFallback className="rounded-xl">
                      {initials(organization.data?.name ?? "Organization")}
                    </AvatarFallback>
                  </Avatar>
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">
                      {logoFile?.name ?? "Current organization logo"}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      The logo appears across organization records.
                    </p>
                  </div>
                </div>
                <Field>
                  <FieldLabel htmlFor="organization-logo">Choose logo</FieldLabel>
                  <Input
                    id="organization-logo"
                    type="file"
                    accept="image/png,image/jpeg"
                    disabled={organizationLocked}
                    onChange={(event) => {
                      setNotice(null);
                      setLogoFile(event.target.files?.[0] ?? null);
                    }}
                  />
                </Field>
              </CardContent>
              <CardFooter className="justify-end">
                <Button
                  type="button"
                  disabled={organizationLocked || !logoFile || uploadLogo.isPending}
                  onClick={() => uploadLogo.mutate()}
                >
                  {uploadLogo.isPending ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <UploadIcon data-icon="inline-start" />
                  )}
                  Upload logo
                </Button>
              </CardFooter>
            </Card>

            <Card className="min-w-0">
              <form
                onSubmit={(event: FormEvent) => {
                  event.preventDefault();
                  setNotice(null);
                  updateDetails.mutate();
                }}
              >
                <CardHeader>
                  <CardTitle>General information</CardTitle>
                  <CardDescription>
                    Update the organization name and unique URL slug.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <FieldGroup>
                    <Field>
                      <FieldLabel htmlFor="organization-name">Organization name</FieldLabel>
                      <Input
                        id="organization-name"
                        value={details.name}
                        maxLength={100}
                        required
                        disabled={organizationLocked}
                        onChange={(event) =>
                          setDetails((current) => ({ ...current, name: event.target.value }))
                        }
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="organization-slug">URL slug</FieldLabel>
                      <Input
                        id="organization-slug"
                        value={details.slug}
                        maxLength={50}
                        required
                        disabled={organizationLocked}
                        onChange={(event) =>
                          setDetails((current) => ({
                            ...current,
                            slug: slugify(event.target.value),
                          }))
                        }
                      />
                      <FieldDescription>
                        Lowercase letters, numbers, and hyphens only.
                      </FieldDescription>
                    </Field>
                  </FieldGroup>
                </CardContent>
                <CardFooter className="justify-end">
                  <Button
                    type="submit"
                    disabled={
                      organizationLocked ||
                      !details.name.trim() ||
                      !details.slug.trim() ||
                      updateDetails.isPending
                    }
                  >
                    {updateDetails.isPending && <Spinner data-icon="inline-start" />}
                    Save changes
                  </Button>
                </CardFooter>
              </form>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="members">
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Organization members</CardTitle>
              <CardDescription>
                Manage access for {members.data?.length ?? 0} organization members.
              </CardDescription>
              <CardAction>
                <Button size="sm" disabled={organizationLocked} onClick={() => setInviteOpen(true)}>
                  <MailPlusIcon data-icon="inline-start" />
                  Invite member
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent className="flex min-w-0 flex-col gap-4">
              <div className="relative max-w-sm">
                <SearchIcon
                  className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  className="pl-9"
                  placeholder="Search members..."
                  aria-label="Search organization members"
                  value={memberSearch}
                  onChange={(event) => setMemberSearch(event.target.value)}
                />
              </div>
              {members.isPending ? (
                <TableSkeleton />
              ) : filteredMembers.length ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Member</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead className="hidden lg:table-cell">Permissions</TableHead>
                      <TableHead className="hidden md:table-cell">Joined</TableHead>
                      <TableHead className="w-24">
                        <span className="sr-only">Actions</span>
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredMembers.map((member) => (
                      <TableRow key={member.id}>
                        <TableCell>
                          <div className="flex min-w-0 items-center gap-3">
                            <Avatar>
                              <AvatarImage
                                src={member.user.image ?? undefined}
                                alt={member.user.name}
                              />
                              <AvatarFallback>{initials(member.user.name)}</AvatarFallback>
                            </Avatar>
                            <div className="min-w-0">
                              <p className="truncate font-medium">{member.user.name}</p>
                              <p className="truncate text-xs text-muted-foreground">
                                {member.user.email}
                              </p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          {member.role === "owner" ? (
                            <Badge variant="secondary">Owner</Badge>
                          ) : (
                            <Select
                              value={member.role}
                              disabled={organizationLocked || changeRole.isPending}
                              onValueChange={(value) => {
                                setNotice(null);
                                changeRole.mutate({
                                  userId: member.user_id,
                                  role: value as ManagedRole,
                                });
                              }}
                            >
                              <SelectTrigger
                                className="w-28"
                                aria-label={`Role for ${member.user.name}`}
                              >
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectGroup>
                                  <SelectItem value="admin">Admin</SelectItem>
                                  <SelectItem value="member">Member</SelectItem>
                                </SelectGroup>
                              </SelectContent>
                            </Select>
                          )}
                        </TableCell>
                        <TableCell className="hidden lg:table-cell">
                          {member.role === "owner" || member.role === "admin" ? (
                            <Badge variant="outline">Full role access</Badge>
                          ) : member.permissions.length ? (
                            <Badge variant="secondary">
                              {member.permissions.length} permission
                              {member.permissions.length === 1 ? "" : "s"}
                            </Badge>
                          ) : (
                            <Badge variant="outline">No elevated access</Badge>
                          )}
                        </TableCell>
                        <TableCell className="hidden text-muted-foreground md:table-cell">
                          {formatDate(member.created_at)}
                        </TableCell>
                        <TableCell>
                          {member.role !== "owner" && (
                            <div className="flex items-center gap-1">
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label={`Configure permissions for ${member.user.name}`}
                                disabled={organizationLocked}
                                onClick={() => {
                                  setPermissionMember(member);
                                  setSelectedPermissions(member.permissions);
                                }}
                              >
                                <KeyRoundIcon />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label={`Remove ${member.user.name}`}
                                disabled={organizationLocked}
                                onClick={() => setMemberToRemove(member)}
                              >
                                <Trash2Icon />
                              </Button>
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Empty className="border">
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <UsersIcon />
                    </EmptyMedia>
                    <EmptyTitle>No members found</EmptyTitle>
                    <EmptyDescription>
                      {memberSearch
                        ? "Try a different name, email, or role."
                        : "Invite a member to start building your organization team."}
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="invitations">
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Invitations</CardTitle>
              <CardDescription>
                Track and manage access invitations sent by your team.
              </CardDescription>
              <CardAction>
                <Button size="sm" disabled={organizationLocked} onClick={() => setInviteOpen(true)}>
                  <MailPlusIcon data-icon="inline-start" />
                  New invitation
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent>
              {invitations.isPending ? (
                <TableSkeleton />
              ) : invitations.data?.length ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Email</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="hidden md:table-cell">Expires</TableHead>
                      <TableHead className="w-14">
                        <span className="sr-only">Actions</span>
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {invitations.data.map((invitation) => (
                      <TableRow key={invitation.id}>
                        <TableCell className="font-medium">{invitation.email}</TableCell>
                        <TableCell className="capitalize">{invitation.role}</TableCell>
                        <TableCell>
                          <Badge
                            variant={invitation.status === "pending" ? "secondary" : "outline"}
                          >
                            {capitalize(invitation.status)}
                          </Badge>
                        </TableCell>
                        <TableCell className="hidden text-muted-foreground md:table-cell">
                          {formatDate(invitation.expires_at)}
                        </TableCell>
                        <TableCell>
                          {invitation.status === "pending" && (
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              disabled={organizationLocked || cancelInvitation.isPending}
                              aria-label={`Cancel invitation for ${invitation.email}`}
                              onClick={() => {
                                setNotice(null);
                                cancelInvitation.mutate(invitation.id);
                              }}
                            >
                              <Trash2Icon />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Empty className="border">
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <MailPlusIcon />
                    </EmptyMedia>
                    <EmptyTitle>No invitations yet</EmptyTitle>
                    <EmptyDescription>
                      Send an invitation to add administrators or members.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="audit">
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Audit trail</CardTitle>
              <CardDescription>
                Immutable history of organization, member, permission, and invitation changes.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {auditLogs.isPending ? (
                <TableSkeleton />
              ) : auditLogs.data?.length ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Action</TableHead>
                      <TableHead>Performed by</TableHead>
                      <TableHead className="hidden lg:table-cell">Details</TableHead>
                      <TableHead className="hidden md:table-cell">Date</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {auditLogs.data.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>
                          <div className="flex min-w-44 items-center gap-2">
                            <HistoryIcon
                              className="size-4 text-muted-foreground"
                              aria-hidden="true"
                            />
                            <div className="min-w-0">
                              <p className="font-medium">{formatAuditAction(log.action)}</p>
                              <p className="truncate text-xs text-muted-foreground">
                                {log.target_type}
                              </p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="min-w-44">
                            <p className="truncate font-medium">{log.actor_name}</p>
                            <p className="truncate text-xs text-muted-foreground">
                              {log.actor_email}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell className="hidden max-w-80 lg:table-cell">
                          <p
                            className="truncate text-muted-foreground"
                            title={formatAuditDetails(log.details)}
                          >
                            {formatAuditDetails(log.details)}
                          </p>
                        </TableCell>
                        <TableCell className="hidden text-muted-foreground md:table-cell">
                          {formatDateTime(log.created_at)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Empty className="border">
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <HistoryIcon />
                    </EmptyMedia>
                    <EmptyTitle>No audit events yet</EmptyTitle>
                    <EmptyDescription>
                      Administrative changes will appear here automatically.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
        <DialogContent>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              setNotice(null);
              sendInvitation.mutate();
            }}
          >
            <DialogHeader>
              <DialogTitle>Invite organization member</DialogTitle>
              <DialogDescription>
                Send access to an administrator or member. Ownership is assigned by a platform
                admin.
              </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
              <Field>
                <FieldLabel htmlFor="invite-email">Email address</FieldLabel>
                <Input
                  id="invite-email"
                  type="email"
                  autoComplete="email"
                  placeholder="name@example.com"
                  required
                  disabled={organizationLocked}
                  value={inviteForm.email}
                  onChange={(event) =>
                    setInviteForm((current) => ({ ...current, email: event.target.value }))
                  }
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="invite-role">Organization role</FieldLabel>
                <Select
                  value={inviteForm.role}
                  disabled={organizationLocked}
                  onValueChange={(role) =>
                    setInviteForm((current) => ({ ...current, role: role as ManagedRole }))
                  }
                >
                  <SelectTrigger id="invite-role" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="member">Member</SelectItem>
                      <SelectItem value="admin">Admin</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
                <FieldDescription>
                  Admins can manage academic records; members have standard access.
                </FieldDescription>
              </Field>
            </FieldGroup>
            {sendInvitation.error && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircleIcon />
                <AlertTitle>Invitation failed</AlertTitle>
                <AlertDescription>{sendInvitation.error.message}</AlertDescription>
              </Alert>
            )}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setInviteOpen(false)}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={
                  organizationLocked || !inviteForm.email.trim() || sendInvitation.isPending
                }
              >
                {sendInvitation.isPending && <Spinner data-icon="inline-start" />}
                Send invitation
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog
        open={permissionMember !== null}
        onOpenChange={(open) => !open && setPermissionMember(null)}
      >
        <DialogContent>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              setNotice(null);
              updatePermissions.mutate();
            }}
          >
            <DialogHeader>
              <DialogTitle>Configure member permissions</DialogTitle>
              <DialogDescription>
                {permissionMember
                  ? `Choose the capabilities available to ${permissionMember.user.name}.`
                  : "Choose the capabilities available to this member."}
              </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
              {permissionMember?.role === "admin" && (
                <Alert>
                  <ShieldCheckIcon />
                  <AlertTitle>Administrator has full access</AlertTitle>
                  <AlertDescription>
                    These custom permissions take effect if this account is changed to Member.
                  </AlertDescription>
                </Alert>
              )}
              <Field>
                <FieldLabel>Granted capabilities</FieldLabel>
                <ToggleGroup
                  type="multiple"
                  variant="outline"
                  value={selectedPermissions}
                  className="grid w-full grid-cols-1 sm:grid-cols-2"
                  disabled={organizationLocked}
                  onValueChange={(permissions) =>
                    setSelectedPermissions(permissions as api.OrganizationPermission[])
                  }
                >
                  {permissionOptions.map((permission) => (
                    <ToggleGroupItem
                      key={permission.value}
                      value={permission.value}
                      aria-label={permission.label}
                      className="h-auto min-h-14 justify-start px-3 text-left whitespace-normal"
                    >
                      <span className="flex min-w-0 flex-col items-start gap-0.5">
                        <span>{permission.label}</span>
                        <span className="text-xs text-muted-foreground">
                          {permission.description}
                        </span>
                      </span>
                    </ToggleGroupItem>
                  ))}
                </ToggleGroup>
                <FieldDescription>
                  Owner and Admin roles retain full access regardless of custom permissions.
                </FieldDescription>
              </Field>
            </FieldGroup>
            {updatePermissions.error && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircleIcon />
                <AlertTitle>Unable to save permissions</AlertTitle>
                <AlertDescription>{updatePermissions.error.message}</AlertDescription>
              </Alert>
            )}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setPermissionMember(null)}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={organizationLocked || !permissionMember || updatePermissions.isPending}
              >
                {updatePermissions.isPending && <Spinner data-icon="inline-start" />}
                Save permissions
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={memberToRemove !== null}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove this member?</AlertDialogTitle>
            <AlertDialogDescription>
              {memberToRemove
                ? `${memberToRemove.user.name} will lose access to this organization and its records.`
                : "This member will lose organization access."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={removeMember.isPending}
              onClick={() => memberToRemove && removeMember.mutate(memberToRemove.user_id)}
            >
              {removeMember.isPending && <Spinner data-icon="inline-start" />}
              Remove member
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function SummaryCard({
  label,
  value,
  icon: Icon,
}: {
  label: string;
  value: number;
  icon: typeof UsersIcon;
}) {
  return (
    <Card size="sm" className="min-w-0">
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{label}</CardTitle>
        <CardAction>
          <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
        </CardAction>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold tracking-tight">{value}</p>
      </CardContent>
    </Card>
  );
}

function TableSkeleton() {
  return (
    <div className="flex flex-col gap-3" aria-label="Loading records">
      {[0, 1, 2].map((row) => (
        <Skeleton key={row} className="h-12 w-full" />
      ))}
    </div>
  );
}

async function load<T>(promise: Promise<api.ApiResponse<T>>): Promise<T> {
  const response = await promise;
  if (!response.success || response.data === undefined) {
    throw new Error(response.message ?? "Unable to complete the request");
  }
  return response.data;
}

async function complete(promise: Promise<api.ApiResponse<void>>): Promise<void> {
  const response = await promise;
  if (!response.success) {
    throw new Error(response.message ?? "Unable to complete the request");
  }
}

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function initials(value: string) {
  return (
    value
      .trim()
      .split(/\s+/)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join("") || "OR"
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-MY", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-MY", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatAuditAction(value: string) {
  return value
    .split(".")
    .map((part) => capitalize(part))
    .join(" ");
}

function formatAuditDetails(details: Record<string, unknown>) {
  const entries = Object.entries(details);
  if (!entries.length) return "No additional details";
  return entries
    .map(([key, value]) => `${key.replace(/_/g, " ")}: ${formatDetail(value)}`)
    .join(" · ");
}

function formatDetail(value: unknown) {
  if (Array.isArray(value)) return value.join(", ") || "None";
  if (value === null || value === undefined || value === "") return "None";
  return String(value);
}

function capitalize(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}
