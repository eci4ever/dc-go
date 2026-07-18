import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type ColumnFiltersState,
  type SortingState,
} from "@tanstack/react-table";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { AlertCircleIcon, ArrowUpDownIcon, UsersIcon } from "lucide-react";

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
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
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
import type { User, UserRole } from "@/lib/api";
import { sessionQueryOptions } from "@/lib/session";

const usersQueryKey = ["admin", "users"] as const;

export const Route = createFileRoute("/_protected/admin/users")({
  beforeLoad: async ({ context }) => {
    const session = await context.queryClient.ensureQueryData(sessionQueryOptions);
    if (session?.user.role !== "admin") throw redirect({ to: "/dashboard", replace: true });
  },
  component: UsersPage,
});

function UsersPage() {
  const { user: currentUser } = useAuth();
  const queryClient = useQueryClient();
  const [pendingChange, setPendingChange] = useState<{ user: User; role: UserRole } | null>(null);
  const users = useQuery({
    queryKey: usersQueryKey,
    queryFn: async () => {
      const response = await api.listUsers();
      if (!response.success || !response.data)
        throw new Error(response.message ?? "Unable to load users");
      return response.data;
    },
  });
  const updateRole = useMutation({
    mutationFn: async ({ user, role }: { user: User; role: UserRole }) => {
      const response = await api.updateUserRole(user.id, role);
      if (!response.success || !response.data)
        throw new Error(response.message ?? "Unable to update role");
      return response.data;
    },
    onSuccess: (updated) => {
      queryClient.setQueryData<User[]>(usersQueryKey, (current = []) =>
        current.map((user) => (user.id === updated.id ? updated : user)),
      );
    },
  });

  const columns = useMemo<ColumnDef<User>[]>(
    () => [
      {
        id: "user",
        accessorFn: (user) => `${user.name} ${user.email}`,
        header: ({ column }) => (
          <Button
            variant="ghost"
            className="-ml-2"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            User
            <ArrowUpDownIcon data-icon="inline-end" />
          </Button>
        ),
        cell: ({ row }) => {
          const user = row.original;
          return (
            <div className="flex min-w-56 items-center gap-3">
              <Avatar>
                <AvatarImage src={user.image ?? undefined} alt={user.name} />
                <AvatarFallback>{initials(user.name)}</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="truncate font-medium">{user.name}</span>
                  {user.id === currentUser?.id && <Badge variant="outline">You</Badge>}
                </div>
                <p className="truncate text-sm text-muted-foreground">{user.email}</p>
              </div>
            </div>
          );
        },
      },
      {
        accessorKey: "banned",
        header: "Status",
        cell: ({ row }) => (
          <Badge variant={row.original.banned ? "destructive" : "secondary"}>
            {row.original.banned ? "Banned" : "Active"}
          </Badge>
        ),
      },
      {
        accessorKey: "createdAt",
        header: ({ column }) => (
          <Button
            variant="ghost"
            className="-ml-2"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Joined
            <ArrowUpDownIcon data-icon="inline-end" />
          </Button>
        ),
        cell: ({ row }) => (
          <span className="text-muted-foreground">{formatDate(row.original.createdAt)}</span>
        ),
      },
      {
        accessorKey: "role",
        header: () => <div className="text-right">Global role</div>,
        cell: ({ row }) => {
          const user = row.original;
          return (
            <div className="flex justify-end">
              <Select
                value={user.role}
                disabled={user.id === currentUser?.id || updateRole.isPending}
                onValueChange={(role: UserRole) => {
                  if (role !== user.role) setPendingChange({ user, role });
                }}
              >
                <SelectTrigger size="sm" aria-label={`Role for ${user.name}`}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="user">User</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          );
        },
      },
    ],
    [currentUser?.id, updateRole.isPending],
  );

  const confirmChange = () => {
    if (!pendingChange) return;
    updateRole.mutate(pendingChange);
    setPendingChange(null);
  };

  return (
    <div className="flex w-full min-w-0 max-w-full flex-col gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
        <p className="text-sm text-muted-foreground">Manage global application roles.</p>
      </div>

      {users.error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Unable to load users</AlertTitle>
          <AlertDescription>{users.error.message}</AlertDescription>
        </Alert>
      )}
      {updateRole.error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Role update failed</AlertTitle>
          <AlertDescription>{updateRole.error.message}</AlertDescription>
        </Alert>
      )}

      {users.isPending ? (
        <Card className="min-w-0">
          <CardHeader>
            <CardTitle>Application users</CardTitle>
            <CardDescription>Global roles do not grant access to organizations.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <Skeleton className="h-9 w-full max-w-sm" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </CardContent>
        </Card>
      ) : users.data?.length === 0 ? (
        <Card className="min-w-0">
          <CardContent>
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <UsersIcon />
                </EmptyMedia>
                <EmptyTitle>No users</EmptyTitle>
                <EmptyDescription>No application users were found.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </CardContent>
        </Card>
      ) : users.data ? (
        <UsersDataTable columns={columns} data={users.data} />
      ) : null}

      <AlertDialog
        open={pendingChange !== null}
        onOpenChange={(open) => !open && setPendingChange(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Change global role?</AlertDialogTitle>
            <AlertDialogDescription>
              {pendingChange
                ? `${pendingChange.user.name} will become ${pendingChange.role}. This changes platform-level access only.`
                : "Confirm the global role change."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmChange}>Confirm</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function UsersDataTable({ columns, data }: { columns: ColumnDef<User>[]; data: User[] }) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { sorting, columnFilters },
  });

  return (
    <Card className="min-w-0">
      <CardHeader>
        <CardTitle>Application users</CardTitle>
        <CardDescription>
          Global roles do not grant access to individual organizations.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex min-w-0 flex-col gap-4">
        <Input
          className="w-full sm:max-w-sm"
          placeholder="Search users..."
          aria-label="Search users"
          value={(table.getColumn("user")?.getFilterValue() as string) ?? ""}
          onChange={(event) => table.getColumn("user")?.setFilterValue(event.target.value)}
        />
        <div className="min-w-0 overflow-hidden rounded-md border">
          <Table className="min-w-176">
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    No users match your search.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-3 border-t sm:flex-row sm:justify-between">
        <p className="text-sm text-muted-foreground">
          {table.getFilteredRowModel().rows.length} user
          {table.getFilteredRowModel().rows.length === 1 ? "" : "s"}
        </p>
        <div className="flex w-full gap-2 sm:w-auto">
          <Button
            variant="outline"
            size="sm"
            className="flex-1 sm:flex-none"
            disabled={!table.getCanPreviousPage()}
            onClick={() => table.previousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="flex-1 sm:flex-none"
            disabled={!table.getCanNextPage()}
            onClick={() => table.nextPage()}
          >
            Next
          </Button>
        </div>
      </CardFooter>
    </Card>
  );
}

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Unknown";
  return new Intl.DateTimeFormat("en-MY", { dateStyle: "medium" }).format(date);
}
