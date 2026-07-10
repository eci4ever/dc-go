import { useEffect, useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  BadgeCheckIcon,
  CircleAlertIcon,
  CircleCheckIcon,
  KeyRoundIcon,
  LaptopIcon,
  LogOutIcon,
  ShieldCheckIcon,
  ShieldPlusIcon,
  SmartphoneIcon,
  TerminalIcon,
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
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/hooks/use-auth";
import * as api from "@/lib/api";
import type { ManagedSession, SessionData } from "@/lib/api";
import { sessionQueryKey } from "@/lib/session";

const sessionsQueryKey = ["auth", "sessions"] as const;

export const Route = createFileRoute("/_protected/account")({ component: AccountPage });

function AccountPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [name, setName] = useState(user?.name ?? "");
  const [image, setImage] = useState(user?.image ?? "");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [sessionToRevoke, setSessionToRevoke] = useState<ManagedSession | null>(null);

  useEffect(() => {
    if (!user) return;
    setName(user.name);
    setImage(user.image ?? "");
  }, [user]);

  const updateProfile = useMutation({
    mutationFn: async () => {
      const response = await api.updateProfile(name.trim(), image.trim() || null);
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to update your profile");
      }
      return response.data;
    },
    onSuccess: (updatedUser) => {
      queryClient.setQueryData<SessionData | null>(sessionQueryKey, (current) =>
        current ? { ...current, user: updatedUser } : current,
      );
    },
  });

  const changePassword = useMutation({
    mutationFn: async () => {
      const response = await api.changePassword(currentPassword, newPassword);
      if (!response.success) throw new Error(response.message ?? "Unable to change password");
    },
    onSuccess: () => {
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setPasswordError(null);
    },
  });

  const sessions = useQuery({
    queryKey: sessionsQueryKey,
    queryFn: async () => {
      const response = await api.listSessions();
      if (!response.success || !response.data) {
        throw new Error(response.message ?? "Unable to load sessions");
      }
      return response.data;
    },
  });

  const revokeSession = useMutation({
    mutationFn: async (sessionId: string) => {
      const response = await api.revokeSession(sessionId);
      if (!response.success) throw new Error(response.message ?? "Unable to revoke session");
      return sessionId;
    },
    onSuccess: (sessionId) => {
      queryClient.setQueryData<ManagedSession[]>(sessionsQueryKey, (current = []) =>
        current.filter((session) => session.id !== sessionId),
      );
    },
  });

  if (!user) return null;

  const submitProfile = (event: FormEvent) => {
    event.preventDefault();
    updateProfile.mutate();
  };

  const submitPassword = (event: FormEvent) => {
    event.preventDefault();
    setPasswordError(null);
    changePassword.reset();
    if (newPassword !== confirmPassword) {
      setPasswordError("New passwords do not match.");
      return;
    }
    if (newPassword === currentPassword) {
      setPasswordError("New password must be different from your current password.");
      return;
    }
    if (newPassword.length < 8) {
      setPasswordError("New password must contain at least 8 characters.");
      return;
    }
    changePassword.mutate();
  };

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Account</h1>
        <p className="text-sm text-muted-foreground">
          Manage your personal details and account security.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Profile information</CardTitle>
          <CardDescription>
            Update how your profile appears throughout the application.
          </CardDescription>
        </CardHeader>
        <form onSubmit={submitProfile}>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-[auto_minmax(0,1fr)]">
              <Avatar className="rounded-xl data-[size=lg]:size-20" size="lg">
                <AvatarImage src={image.trim() || undefined} alt={name || user.name} />
                <AvatarFallback className="rounded-xl text-lg">
                  {(name || user.name).slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <FieldGroup className="gap-5">
                <Field>
                  <FieldLabel htmlFor="account-name">Name</FieldLabel>
                  <Input
                    id="account-name"
                    value={name}
                    required
                    maxLength={100}
                    autoComplete="name"
                    onChange={(event) => {
                      updateProfile.reset();
                      setName(event.target.value);
                    }}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="account-avatar">Avatar URL</FieldLabel>
                  <Input
                    id="account-avatar"
                    type="url"
                    value={image}
                    maxLength={2048}
                    placeholder="https://example.com/avatar.jpg"
                    onChange={(event) => {
                      updateProfile.reset();
                      setImage(event.target.value);
                    }}
                  />
                  <FieldDescription>Use a secure, publicly accessible image URL.</FieldDescription>
                </Field>
                <Field data-disabled>
                  <div className="flex items-center justify-between gap-3">
                    <FieldLabel htmlFor="account-email">Email address</FieldLabel>
                    <Badge variant={user.emailVerified ? "secondary" : "outline"}>
                      {user.emailVerified ? <BadgeCheckIcon /> : <CircleAlertIcon />}
                      {user.emailVerified ? "Verified" : "Unverified"}
                    </Badge>
                  </div>
                  <Input id="account-email" value={user.email} disabled readOnly />
                  <FieldDescription>Your email address cannot be changed.</FieldDescription>
                </Field>
              </FieldGroup>
            </div>
            {updateProfile.error && (
              <Alert variant="destructive" className="mt-5">
                <CircleAlertIcon />
                <AlertTitle>Profile update failed</AlertTitle>
                <AlertDescription>{updateProfile.error.message}</AlertDescription>
              </Alert>
            )}
            {updateProfile.isSuccess && (
              <Alert className="mt-5">
                <CircleCheckIcon />
                <AlertTitle>Profile updated</AlertTitle>
                <AlertDescription>Your profile changes have been saved.</AlertDescription>
              </Alert>
            )}
          </CardContent>
          <CardFooter className="justify-end">
            <Button type="submit" disabled={updateProfile.isPending || !name.trim()}>
              {updateProfile.isPending && <Spinner data-icon="inline-start" />}
              Save changes
            </Button>
          </CardFooter>
        </form>
      </Card>

      <div className="grid items-stretch gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                <KeyRoundIcon />
              </div>
              <div>
                <CardTitle>Change password</CardTitle>
                <CardDescription>Use at least 8 characters for your new password.</CardDescription>
              </div>
            </div>
          </CardHeader>
          <form className="flex flex-1 flex-col" onSubmit={submitPassword}>
            <CardContent className="flex-1">
              <FieldGroup className="gap-5">
                <Field>
                  <FieldLabel htmlFor="current-password">Current password</FieldLabel>
                  <Input
                    id="current-password"
                    type="password"
                    value={currentPassword}
                    required
                    autoComplete="current-password"
                    onChange={(event) => setCurrentPassword(event.target.value)}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="new-password">New password</FieldLabel>
                  <Input
                    id="new-password"
                    type="password"
                    value={newPassword}
                    required
                    minLength={8}
                    maxLength={72}
                    autoComplete="new-password"
                    onChange={(event) => setNewPassword(event.target.value)}
                  />
                </Field>
                <Field data-invalid={passwordError !== null}>
                  <FieldLabel htmlFor="confirm-password">Confirm new password</FieldLabel>
                  <Input
                    id="confirm-password"
                    type="password"
                    value={confirmPassword}
                    required
                    minLength={8}
                    maxLength={72}
                    autoComplete="new-password"
                    aria-invalid={passwordError !== null}
                    onChange={(event) => setConfirmPassword(event.target.value)}
                  />
                  {passwordError && <FieldError>{passwordError}</FieldError>}
                </Field>
              </FieldGroup>
              {changePassword.error && (
                <Alert variant="destructive" className="mt-5">
                  <CircleAlertIcon />
                  <AlertTitle>Password change failed</AlertTitle>
                  <AlertDescription>{changePassword.error.message}</AlertDescription>
                </Alert>
              )}
              {changePassword.isSuccess && (
                <Alert className="mt-5">
                  <CircleCheckIcon />
                  <AlertTitle>Password changed</AlertTitle>
                  <AlertDescription>Your new password is active.</AlertDescription>
                </Alert>
              )}
            </CardContent>
            <CardFooter className="justify-end">
              <Button type="submit" disabled={changePassword.isPending}>
                {changePassword.isPending && <Spinner data-icon="inline-start" />}
                Update password
              </Button>
            </CardFooter>
          </form>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                <ShieldCheckIcon />
              </div>
              <div>
                <CardTitle>Two-factor authentication</CardTitle>
                <CardDescription>Add an extra layer of protection to your account.</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="flex flex-1 flex-col gap-4">
            <div className="flex items-center justify-between gap-4 rounded-lg bg-muted/50 p-4">
              <div>
                <p className="font-medium">Authenticator app</p>
                <p className="text-sm text-muted-foreground">
                  {user.twoFactorEnabled
                    ? "Two-factor authentication is active."
                    : "Two-factor authentication is not configured."}
                </p>
              </div>
              <Badge variant={user.twoFactorEnabled ? "secondary" : "outline"}>
                {user.twoFactorEnabled ? "Enabled" : "Disabled"}
              </Badge>
            </div>
            <Alert>
              <ShieldPlusIcon />
              <AlertTitle>Enrollment coming soon</AlertTitle>
              <AlertDescription>
                Secure authenticator enrollment and recovery codes are not available yet.
              </AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter className="justify-end">
            <Button variant="outline" disabled>
              <ShieldPlusIcon data-icon="inline-start" />
              Set up two-factor
            </Button>
          </CardFooter>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Active sessions</CardTitle>
          <CardDescription>
            Review devices signed in to your account and revoke any session you do not recognize.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {sessions.isPending ? (
            <div className="flex flex-col gap-3">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : sessions.error ? (
            <Alert variant="destructive">
              <CircleAlertIcon />
              <AlertTitle>Unable to load sessions</AlertTitle>
              <AlertDescription>{sessions.error.message}</AlertDescription>
            </Alert>
          ) : sessions.data?.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <LaptopIcon />
                </EmptyMedia>
                <EmptyTitle>No active sessions</EmptyTitle>
                <EmptyDescription>
                  No sessions are currently associated with this account.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="flex flex-col">
              {sessions.data?.map((session, index) => {
                const presentation = sessionPresentation(session);
                const DeviceIcon = presentation.icon;
                return (
                  <div key={session.id}>
                    {index > 0 && <Separator />}
                    <div className="flex flex-col gap-4 py-4 first:pt-0 last:pb-0 sm:flex-row sm:items-center">
                      <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted">
                        <DeviceIcon />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="font-medium">{presentation.label}</p>
                          {session.current && <Badge variant="secondary">Current</Badge>}
                        </div>
                        <p
                          className="truncate text-sm text-muted-foreground"
                          title={session.userAgent ?? undefined}
                        >
                          {session.ipAddress ?? "Unknown IP"} · Last active{" "}
                          {formatSessionDate(session.updatedAt)}
                        </p>
                      </div>
                      {!session.current && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="self-end sm:self-auto"
                          disabled={revokeSession.isPending}
                          onClick={() => setSessionToRevoke(session)}
                        >
                          <LogOutIcon data-icon="inline-start" />
                          Revoke
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
          {revokeSession.error && (
            <Alert variant="destructive" className="mt-5">
              <CircleAlertIcon />
              <AlertTitle>Session revocation failed</AlertTitle>
              <AlertDescription>{revokeSession.error.message}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      <AlertDialog
        open={sessionToRevoke !== null}
        onOpenChange={(open) => !open && setSessionToRevoke(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke this session?</AlertDialogTitle>
            <AlertDialogDescription>
              This device will lose access when its current access token expires and will need to
              sign in again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                if (sessionToRevoke) revokeSession.mutate(sessionToRevoke.id);
                setSessionToRevoke(null);
              }}
            >
              Revoke session
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function sessionPresentation(session: ManagedSession) {
  const userAgent = session.userAgent?.toLowerCase() ?? "";
  if (
    userAgent.includes("iphone") ||
    userAgent.includes("android") ||
    userAgent.includes("mobile")
  ) {
    return { label: "Mobile device", icon: SmartphoneIcon };
  }
  if (userAgent.includes("curl") || userAgent.includes("postman")) {
    return { label: "API client", icon: TerminalIcon };
  }
  return { label: "Web browser", icon: LaptopIcon };
}

function formatSessionDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
