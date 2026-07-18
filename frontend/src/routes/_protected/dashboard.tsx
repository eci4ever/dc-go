import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  ArrowRightIcon,
  BadgeCheckIcon,
  CalendarDaysIcon,
  KeyRoundIcon,
  LaptopIcon,
  MailCheckIcon,
  ShieldCheckIcon,
  UserRoundIcon,
} from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/hooks/use-auth";
import * as api from "@/lib/api";

const sessionsQueryKey = ["auth", "sessions"] as const;

export const Route = createFileRoute("/_protected/dashboard")({ component: DashboardPage });

function DashboardPage() {
  const { user, session } = useAuth();
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

  if (!user || !session) return null;

  const firstName = user.name.trim().split(/\s+/)[0] || "there";
  const securityChecks = [!user.banned, user.emailVerified, user.twoFactorEnabled];
  const completedSecurityChecks = securityChecks.filter(Boolean).length;
  const activities = [
    {
      title: "Current session started",
      description: session.session.ipAddress ?? "IP address unavailable",
      date: session.session.createdAt,
      icon: LaptopIcon,
    },
    {
      title: "Profile last updated",
      description: "Your latest account details were saved",
      date: user.updatedAt,
      icon: UserRoundIcon,
    },
    {
      title: "Account created",
      description: "Your DC GO account was registered",
      date: user.createdAt,
      icon: CalendarDaysIcon,
    },
  ].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());

  return (
    <div className="flex w-full min-w-0 max-w-full flex-1 flex-col gap-4">
      <Card className="min-w-0">
        <CardHeader>
          <div className="flex min-w-0 items-center gap-4">
            <Avatar className="size-12" size="lg">
              <AvatarImage src={user.image ?? undefined} alt={user.name} />
              <AvatarFallback>{initials(user.name)}</AvatarFallback>
            </Avatar>
            <div className="flex min-w-0 flex-col gap-1">
              <div className="flex flex-wrap items-center gap-2">
                <CardTitle className="text-xl text-balance sm:text-2xl">
                  Welcome back, {firstName}
                </CardTitle>
                <Badge variant="secondary">{user.role === "admin" ? "Admin" : "User"}</Badge>
              </div>
              <CardDescription className="mt-1 text-pretty">
                Here is an overview of your account and security.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <SummaryCard
          label="Account status"
          value={user.banned ? "Restricted" : "Active"}
          description={
            user.banned ? "Contact an administrator" : "Your account is in good standing"
          }
          icon={BadgeCheckIcon}
        />
        <SummaryCard
          label="Email"
          value={user.emailVerified ? "Verified" : "Pending"}
          description={user.email}
          icon={MailCheckIcon}
        />
        <SummaryCard
          label="Two-factor auth"
          value={user.twoFactorEnabled ? "Enabled" : "Disabled"}
          description={
            user.twoFactorEnabled ? "Extra protection is active" : "Set up additional protection"
          }
          icon={ShieldCheckIcon}
        />
        <Card size="sm" className="min-w-0">
          <CardHeader>
            <CardTitle className="text-sm text-muted-foreground">Active sessions</CardTitle>
            <CardAction>
              <LaptopIcon className="size-5 text-muted-foreground" aria-hidden="true" />
            </CardAction>
          </CardHeader>
          <CardContent className="flex min-w-0 flex-col gap-1">
            {sessions.isPending ? (
              <Skeleton className="h-7 w-16" />
            ) : (
              <p className="text-2xl font-semibold tracking-tight">
                {sessions.isError ? "—" : sessions.data.length}
              </p>
            )}
            <p className="truncate text-xs text-muted-foreground">
              {sessions.isError ? "Unable to load sessions" : "Devices currently signed in"}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)]">
        <Card className="min-w-0">
          <CardHeader>
            <CardTitle>Recent activity</CardTitle>
            <CardDescription>Important events from your account.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col">
            {activities.map((activity, index) => {
              const ActivityIcon = activity.icon;
              return (
                <div key={activity.title}>
                  {index > 0 && <Separator />}
                  <div className="flex min-w-0 items-start gap-3 py-4 first:pt-0 last:pb-0">
                    <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                      <ActivityIcon className="size-4" aria-hidden="true" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="font-medium">{activity.title}</p>
                      <p className="truncate text-sm text-muted-foreground">
                        {activity.description}
                      </p>
                    </div>
                    <time
                      className="hidden shrink-0 text-xs text-muted-foreground sm:block"
                      dateTime={activity.date}
                    >
                      {formatDate(activity.date)}
                    </time>
                  </div>
                </div>
              );
            })}
          </CardContent>
        </Card>

        <div className="flex min-w-0 flex-col gap-4">
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Security overview</CardTitle>
              <CardDescription>
                {completedSecurityChecks} of {securityChecks.length} checks complete.
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              <SecurityItem label="Account active" complete={!user.banned} />
              <SecurityItem label="Email verified" complete={user.emailVerified} />
              <SecurityItem label="Two-factor enabled" complete={user.twoFactorEnabled} />
            </CardContent>
          </Card>

          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Quick actions</CardTitle>
              <CardDescription>Common account settings in one place.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              <QuickAction label="Edit profile" hash="profile" icon={UserRoundIcon} />
              <QuickAction label="Change password" hash="password" icon={KeyRoundIcon} />
              <QuickAction label="Review sessions" hash="sessions" icon={LaptopIcon} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function SummaryCard({
  label,
  value,
  description,
  icon: Icon,
}: {
  label: string;
  value: string;
  description: string;
  icon: typeof BadgeCheckIcon;
}) {
  return (
    <Card size="sm" className="min-w-0">
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{label}</CardTitle>
        <CardAction>
          <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
        </CardAction>
      </CardHeader>
      <CardContent className="flex min-w-0 flex-col gap-1">
        <p className="text-2xl font-semibold tracking-tight">{value}</p>
        <p className="truncate text-xs text-muted-foreground" title={description}>
          {description}
        </p>
      </CardContent>
    </Card>
  );
}

function SecurityItem({ label, complete }: { label: string; complete: boolean }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-sm">{label}</span>
      <Badge variant={complete ? "secondary" : "outline"}>
        {complete ? "Complete" : "Action needed"}
      </Badge>
    </div>
  );
}

function QuickAction({
  label,
  hash,
  icon: Icon,
}: {
  label: string;
  hash: string;
  icon: typeof UserRoundIcon;
}) {
  return (
    <Button asChild variant="outline" className="w-full justify-between">
      <Link to="/account" hash={hash}>
        <span className="flex items-center gap-2">
          <Icon data-icon="inline-start" />
          {label}
        </span>
        <ArrowRightIcon data-icon="inline-end" />
      </Link>
    </Button>
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
