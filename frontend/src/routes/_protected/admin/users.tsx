import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { AlertCircleIcon, UsersIcon } from "lucide-react";

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
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
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

  const confirmChange = () => {
    if (!pendingChange) return;
    updateRole.mutate(pendingChange);
    setPendingChange(null);
  };

  return (
    <div className="flex flex-col gap-4">
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

      <Card>
        <CardHeader>
          <CardTitle>Application users</CardTitle>
          <CardDescription>
            Global roles do not grant access to individual organizations.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {users.isPending ? (
            <div className="flex flex-col gap-3">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : users.data?.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <UsersIcon />
                </EmptyMedia>
                <EmptyTitle>No users</EmptyTitle>
                <EmptyDescription>No application users were found.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : users.data ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Global role</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.data.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-medium">{user.name}</span>
                        <span className="text-sm text-muted-foreground">{user.email}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={user.banned ? "destructive" : "secondary"}>
                        {user.banned ? "Banned" : "Active"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end">
                        <Select
                          value={user.role}
                          disabled={user.id === currentUser?.id || updateRole.isPending}
                          onValueChange={(role: UserRole) => setPendingChange({ user, role })}
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
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </CardContent>
      </Card>

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
