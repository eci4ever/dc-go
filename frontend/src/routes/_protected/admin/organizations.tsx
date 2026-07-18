import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type FormEvent,
} from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, redirect } from "@tanstack/react-router";
import {
  AlertCircleIcon,
  Building2Icon,
  PencilIcon,
  PlusIcon,
  SearchIcon,
  Trash2Icon,
  UploadIcon,
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
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
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
import type { Organization } from "@/lib/api";
import { sessionQueryOptions } from "@/lib/session";

const organizationsQueryKey = ["admin", "organizations"] as const;
const pageSize = 10;
const maxLogoBytes = 2 * 1024 * 1024;

const emptyForm = { id: "", name: "", slug: "", logo: null as string | null };

export const Route = createFileRoute("/_protected/admin/organizations")({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(sessionQueryOptions);
    if (session?.user.role !== "admin") throw redirect({ to: "/dashboard", replace: true });
  },
  component: OrganizationsPage,
});

function OrganizationsPage() {
  const { session } = useAuth();
  const queryClient = useQueryClient();
  const [form, setForm] = useState(emptyForm);
  const logoInputRef = useRef<HTMLInputElement>(null);
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [logoError, setLogoError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(0);
  const [organizationToDelete, setOrganizationToDelete] = useState<Organization | null>(null);
  const organizations = useQuery({
    queryKey: organizationsQueryKey,
    queryFn: async () => {
      const response = await api.listAdminOrganizations();
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to load organizations");
      }
      return response.data;
    },
  });

  useEffect(
    () => () => {
      if (logoPreview) URL.revokeObjectURL(logoPreview);
    },
    [logoPreview],
  );

  const saveOrganization = useMutation({
    mutationFn: async () => {
      const response = form.id
        ? await api.updateAdminOrganization(form.id, form.name.trim(), form.slug.trim())
        : await api.createAdminOrganization(form.name.trim(), form.slug.trim());
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to save organization");
      }
      if (!logoFile) return response.data;

      const logoResponse = await api.uploadAdminOrganizationLogo(response.data.id, logoFile);
      if (!logoResponse.success || !logoResponse.data) {
        throw new Error(
          logoResponse.message ?? "Organization was saved, but its logo could not be uploaded",
        );
      }
      return logoResponse.data;
    },
    onSuccess: (saved) => {
      queryClient.setQueryData<Organization[]>(organizationsQueryKey, (current = []) => {
        const exists = current.some((organization) => organization.id === saved.id);
        const next = exists
          ? current.map((organization) => (organization.id === saved.id ? saved : organization))
          : [...current, saved];
        return next.sort((a, b) => a.name.localeCompare(b.name));
      });
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      resetForm();
      setDialogOpen(false);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: organizationsQueryKey }),
  });
  const deleteOrganization = useMutation({
    mutationFn: async (id: string) => {
      const response = await api.deleteAdminOrganization(id);
      if (!response.success) {
        throw new Error(response.message ?? "Unable to delete organization");
      }
      return id;
    },
    onSuccess: (id) => {
      queryClient.setQueryData<Organization[]>(organizationsQueryKey, (current = []) =>
        current.filter((organization) => organization.id !== id),
      );
      queryClient.invalidateQueries({ queryKey: ["organizations"] });
      setOrganizationToDelete(null);
    },
  });

  const filteredOrganizations = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return organizations.data ?? [];
    return (organizations.data ?? []).filter(
      (organization) =>
        organization.name.toLowerCase().includes(term) ||
        organization.slug.toLowerCase().includes(term),
    );
  }, [organizations.data, search]);
  const pageCount = Math.max(1, Math.ceil(filteredOrganizations.length / pageSize));
  const currentPage = Math.min(page, pageCount - 1);
  const pageOrganizations = filteredOrganizations.slice(
    currentPage * pageSize,
    currentPage * pageSize + pageSize,
  );

  const submit = (event: FormEvent) => {
    event.preventDefault();
    saveOrganization.mutate();
  };

  const openCreateDialog = () => {
    saveOrganization.reset();
    resetForm();
    setDialogOpen(true);
  };

  const openEditDialog = (organization: Organization) => {
    saveOrganization.reset();
    resetLogoSelection();
    setForm({
      id: organization.id,
      name: organization.name,
      slug: organization.slug,
      logo: organization.logo ?? null,
    });
    setDialogOpen(true);
  };

  const resetLogoSelection = () => {
    setLogoFile(null);
    setLogoPreview(null);
    setLogoError(null);
    if (logoInputRef.current) logoInputRef.current.value = "";
  };

  const resetForm = () => {
    setForm(emptyForm);
    resetLogoSelection();
  };

  const selectLogo = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!(["image/jpeg", "image/png"] as string[]).includes(file.type)) {
      setLogoFile(null);
      setLogoPreview(null);
      setLogoError("Choose a valid PNG or JPEG image.");
      event.target.value = "";
      return;
    }
    if (file.size > maxLogoBytes) {
      setLogoFile(null);
      setLogoPreview(null);
      setLogoError("Logo must be 2 MB or smaller.");
      event.target.value = "";
      return;
    }
    setLogoError(null);
    setLogoFile(file);
    setLogoPreview(URL.createObjectURL(file));
  };

  return (
    <div className="flex w-full min-w-0 max-w-full flex-col gap-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-balance">Organizations</h1>
          <p className="text-sm text-pretty text-muted-foreground">
            Manage ADTEC JTM institutes available across the platform.
          </p>
        </div>
        <Button className="w-full sm:w-auto" onClick={openCreateDialog}>
          <PlusIcon data-icon="inline-start" />
          Add organization
        </Button>
      </div>

      {(organizations.error || deleteOrganization.error) && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Organization operation failed</AlertTitle>
          <AlertDescription>
            {(organizations.error ?? deleteOrganization.error)?.message}
          </AlertDescription>
        </Alert>
      )}

      <Card className="min-w-0">
        <CardHeader>
          <CardTitle>Institute organizations</CardTitle>
          <CardDescription>
            {organizations.data?.length ?? 0} organizations configured.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex min-w-0 flex-col gap-4">
          <div className="relative w-full sm:max-w-sm">
            <SearchIcon
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              className="pl-9"
              placeholder="Search organizations..."
              aria-label="Search organizations"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(0);
              }}
            />
          </div>

          {organizations.isPending ? (
            <div className="flex flex-col gap-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : pageOrganizations.length ? (
            <div className="min-w-0 overflow-hidden rounded-md border">
              <Table className="min-w-176">
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-16">Logo</TableHead>
                    <TableHead>Organization</TableHead>
                    <TableHead>Slug</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pageOrganizations.map((organization) => {
                    const isActive = organization.id === session?.session.activeOrganizationId;
                    return (
                      <TableRow key={organization.id}>
                        <TableCell>
                          <Avatar size="lg">
                            {organization.logo && (
                              <AvatarImage src={organization.logo} alt={`${organization.name} logo`} />
                            )}
                            <AvatarFallback>
                              <Building2Icon className="size-4" aria-hidden="true" />
                            </AvatarFallback>
                          </Avatar>
                        </TableCell>
                        <TableCell>
                          <div className="flex min-w-64 items-center gap-2">
                            <p className="truncate font-medium">{organization.name}</p>
                            {isActive && <Badge variant="secondary">Active</Badge>}
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{organization.slug}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatDate(organization.created_at)}
                        </TableCell>
                        <TableCell>
                          <div className="flex justify-end gap-2">
                            <Button
                              variant="outline"
                              size="icon-sm"
                              aria-label={`Edit ${organization.name}`}
                              onClick={() => openEditDialog(organization)}
                            >
                              <PencilIcon />
                            </Button>
                            <Button
                              variant="destructive"
                              size="icon-sm"
                              aria-label={`Delete ${organization.name}`}
                              disabled={isActive || deleteOrganization.isPending}
                              onClick={() => setOrganizationToDelete(organization)}
                            >
                              <Trash2Icon />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          ) : (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Building2Icon />
                </EmptyMedia>
                <EmptyTitle>No organizations found</EmptyTitle>
                <EmptyDescription>Try a different search term.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
        <CardFooter className="flex-col gap-3 border-t sm:flex-row sm:justify-between">
          <p className="text-sm text-muted-foreground">
            {filteredOrganizations.length} organization
            {filteredOrganizations.length === 1 ? "" : "s"}
          </p>
          <div className="flex w-full gap-2 sm:w-auto">
            <Button
              variant="outline"
              size="sm"
              className="flex-1 sm:flex-none"
              disabled={currentPage === 0}
              onClick={() => setPage((current) => Math.max(0, current - 1))}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="flex-1 sm:flex-none"
              disabled={currentPage >= pageCount - 1}
              onClick={() => setPage((current) => Math.min(pageCount - 1, current + 1))}
            >
              Next
            </Button>
          </div>
        </CardFooter>
      </Card>

      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (saveOrganization.isPending) return;
          setDialogOpen(open);
          if (!open) resetForm();
        }}
      >
        <DialogContent className="max-h-[calc(100svh-2rem)] overflow-y-auto sm:max-w-lg">
          <form className="flex flex-col gap-6" onSubmit={submit}>
            <DialogHeader>
              <DialogTitle>{form.id ? "Edit organization" : "Add organization"}</DialogTitle>
              <DialogDescription>
                {form.id
                  ? "Update the institute name and unique URL slug."
                  : "Create a new institute organization on the platform."}
              </DialogDescription>
            </DialogHeader>
            <FieldGroup>
              <Field data-invalid={Boolean(logoError)}>
                <FieldLabel htmlFor="organization-logo">Organization logo</FieldLabel>
                <div className="flex flex-col gap-4 rounded-lg border p-4 sm:flex-row sm:items-center">
                  <div className="flex min-w-0 flex-1 items-center gap-4">
                    <Avatar size="lg">
                      {(logoPreview || form.logo) && (
                        <AvatarImage
                          src={logoPreview ?? form.logo ?? undefined}
                          alt="Organization logo preview"
                        />
                      )}
                      <AvatarFallback>
                        <Building2Icon aria-hidden="true" />
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex min-w-0 flex-1 flex-col gap-1">
                      <p className="truncate text-sm font-medium">
                        {logoFile?.name ??
                          (form.logo ? "Current organization logo" : "No logo selected")}
                      </p>
                      <p className="text-xs text-muted-foreground">PNG or JPEG, up to 2 MB.</p>
                    </div>
                  </div>
                  <Input
                    ref={logoInputRef}
                    id="organization-logo"
                    type="file"
                    accept="image/png,image/jpeg"
                    className="sr-only"
                    aria-invalid={Boolean(logoError)}
                    onChange={selectLogo}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="w-full sm:w-auto"
                    disabled={saveOrganization.isPending}
                    onClick={() => logoInputRef.current?.click()}
                  >
                    <UploadIcon data-icon="inline-start" />
                    {logoPreview || form.logo ? "Replace" : "Upload"}
                  </Button>
                </div>
                <FieldDescription>
                  A square image works best and will appear in the organization table.
                </FieldDescription>
                {logoError && <FieldError>{logoError}</FieldError>}
              </Field>
              <Field>
                <FieldLabel htmlFor="organization-name">Organization name</FieldLabel>
                <Input
                  id="organization-name"
                  autoFocus
                  required
                  maxLength={100}
                  placeholder="e.g. ADTEC JTM Campus"
                  value={form.name}
                  onChange={(event) => {
                    const name = event.target.value;
                    setForm((current) => ({
                      ...current,
                      name,
                      slug: current.id ? current.slug : slugify(name),
                    }));
                  }}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="organization-slug">URL slug</FieldLabel>
                <Input
                  id="organization-slug"
                  required
                  maxLength={50}
                  pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
                  placeholder="adtec-jtm-campus"
                  value={form.slug}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, slug: slugify(event.target.value) }))
                  }
                />
              </Field>
            </FieldGroup>
            {saveOrganization.error && (
              <Alert variant="destructive">
                <AlertCircleIcon />
                <AlertTitle>Unable to save organization</AlertTitle>
                <AlertDescription>{saveOrganization.error.message}</AlertDescription>
              </Alert>
            )}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={saveOrganization.isPending}
                onClick={() => setDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={saveOrganization.isPending || !form.name.trim() || !form.slug.trim()}
              >
                {saveOrganization.isPending ? (
                  <Spinner data-icon="inline-start" />
                ) : form.id ? (
                  <PencilIcon data-icon="inline-start" />
                ) : (
                  <PlusIcon data-icon="inline-start" />
                )}
                {form.id ? "Save changes" : "Create organization"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={organizationToDelete !== null}
        onOpenChange={(open) => !open && setOrganizationToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this organization?</AlertDialogTitle>
            <AlertDialogDescription>
              {organizationToDelete
                ? `${organizationToDelete.name} and all associated academic records will be permanently deleted.`
                : "This organization will be permanently deleted."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteOrganization.isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteOrganization.isPending}
              onClick={() => {
                if (organizationToDelete) deleteOrganization.mutate(organizationToDelete.id);
              }}
            >
              {deleteOrganization.isPending && <Spinner data-icon="inline-start" />}
              Delete organization
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Unknown";
  return new Intl.DateTimeFormat("en-MY", { dateStyle: "medium" }).format(date);
}
