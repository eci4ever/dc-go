const BASE = "/api/v1";

interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
}
type UserRole = "user" | "admin";
type OrganizationRole = "owner" | "admin" | "member";

interface User {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string | null;
  createdAt: string;
  updatedAt: string;
  role: UserRole;
  banned: boolean;
  banReason: string | null;
  banExpires: string | null;
  twoFactorEnabled: boolean;
}
interface Session {
  id: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
  ipAddress: string | null;
  userAgent: string | null;
  userId: string;
  impersonatedBy: string | null;
  activeOrganizationId: string | null;
  activeOrganizationRole: OrganizationRole | null;
  activeTeamId: string | null;
}
interface SessionData {
  user: User;
  session: Session;
}
interface Organization {
  id: string;
  name: string;
  slug: string;
  logo?: string | null;
  created_at: string;
}
interface ManagedSession {
  id: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
  ipAddress: string | null;
  userAgent: string | null;
  current: boolean;
}
interface AcademicStudent {
  id: string;
  organizationId: string;
  studentNo: string;
  name: string;
  email: string | null;
  program: string;
  intake: string;
  status: "active" | "graduated" | "inactive";
  createdAt: string;
}
interface AcademicSemester {
  id: string;
  organizationId: string;
  code: string;
  name: string;
  academicYear: number;
  sequence: number;
  status: "planned" | "active" | "closed";
}
interface AcademicCourse {
  id: string;
  organizationId: string;
  code: string;
  name: string;
  credits: number;
  active: boolean;
}
interface GradeScale {
  letter: string;
  minScore: number;
  maxScore: number;
  gradePoint: number;
  passing: boolean;
}
interface AcademicResult {
  id: string;
  studentId: string;
  semesterId: string;
  courseId: string;
  courseCode?: string;
  courseName?: string;
  score: number;
  grade: string;
  gradePoint: number;
  credits: number;
  qualityPoint: number;
}
interface SemesterResult {
  semesterId: string;
  semesterCode: string;
  semesterName: string;
  academicYear: number;
  sequence: number;
  results: AcademicResult[];
  totalCredits: number;
  totalPoints: number;
  gpa: number;
  cgpa: number;
}
interface Transcript {
  student: AcademicStudent;
  semesters: SemesterResult[];
  cgpa: number;
}

function csrfToken() {
  return document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token="))
    ?.split("=")[1];
}
let refreshing: Promise<boolean> | null = null;

async function refresh(): Promise<boolean> {
  if (!refreshing)
    refreshing = fetch(`${BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "X-CSRF-Token": csrfToken() ?? "" },
    })
      .then((r) => r.ok)
      .catch(() => false)
      .finally(() => {
        refreshing = null;
      });
  return refreshing;
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retryAuth = true,
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = { ...(options.headers as Record<string, string>) };
  const isFormData = typeof FormData !== "undefined" && options.body instanceof FormData;
  if (options.body && !isFormData) headers["Content-Type"] = "application/json";
  if (!["GET", "HEAD", "OPTIONS"].includes((options.method ?? "GET").toUpperCase()))
    headers["X-CSRF-Token"] = csrfToken() ?? "";
  try {
    const res = await fetch(`${BASE}${path}`, { ...options, headers, credentials: "include" });
    if (res.status === 401 && retryAuth && path !== "/auth/refresh" && (await refresh()))
      return request(path, options, false);
    const body = await res.json().catch(() => null);
    if (body) return body as ApiResponse<T>;
    return { success: res.ok, message: res.ok ? undefined : `Request failed (${res.status})` };
  } catch {
    return { success: false, message: "Unable to reach the server. Please try again." };
  }
}

export const login = (email: string, password: string) =>
  request<SessionData>(
    "/auth/login",
    {
      method: "POST",
      body: JSON.stringify({ email, password }),
    },
    false,
  );
export const register = (name: string, email: string, password: string) =>
  request<SessionData>(
    "/auth/register",
    {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    },
    false,
  );
export async function logout() {
  return request("/auth/logout", { method: "POST" });
}
export const getSession = () => request<SessionData>("/auth/session");
export const updateProfile = (name: string) =>
  request<User>("/users/me", {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
export const uploadAvatar = (file: File) => {
  const body = new FormData();
  body.append("avatar", file);
  return request<User>("/users/me/avatar", { method: "PUT", body });
};
export const removeAvatar = () => request<User>("/users/me/avatar", { method: "DELETE" });
export const changePassword = (currentPassword: string, newPassword: string) =>
  request<void>("/auth/password", {
    method: "PUT",
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
export const listSessions = () => request<ManagedSession[]>("/auth/sessions");
export const revokeSession = (id: string) =>
  request<void>(`/auth/sessions/${id}`, { method: "DELETE" });
export const listOrganizations = () => request<Organization[]>("/organizations");
export const listAdminOrganizations = () => request<Organization[]>("/admin/organizations");
export const createAdminOrganization = (name: string, slug: string) =>
  request<Organization>("/admin/organizations", {
    method: "POST",
    body: JSON.stringify({ name, slug }),
  });
export const updateAdminOrganization = (id: string, name: string, slug: string) =>
  request<Organization>(`/admin/organizations/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name, slug }),
  });
export const uploadAdminOrganizationLogo = (id: string, file: File) => {
  const body = new FormData();
  body.append("logo", file);
  return request<Organization>(`/admin/organizations/${id}/logo`, { method: "PUT", body });
};
export const deleteAdminOrganization = (id: string) =>
  request<void>(`/admin/organizations/${id}`, { method: "DELETE" });
export const setActiveOrganization = (organizationId: string) =>
  request<SessionData>("/auth/session/active-organization", {
    method: "PUT",
    body: JSON.stringify({ organization_id: organizationId }),
  });
export const listUsers = () => request<User[]>("/users");
export const updateUserRole = (id: string, role: UserRole) =>
  request<User>(`/users/${id}/role`, {
    method: "PUT",
    body: JSON.stringify({ role }),
  });
export const listAcademicStudents = (organizationId: string) =>
  request<AcademicStudent[]>(`/organizations/${organizationId}/academic/students`);
export const createAcademicStudent = (
  organizationId: string,
  data: Pick<AcademicStudent, "studentNo" | "name" | "program" | "intake"> & { email: string },
) =>
  request<AcademicStudent>(`/organizations/${organizationId}/academic/students`, {
    method: "POST",
    body: JSON.stringify(data),
  });
export const listAcademicSemesters = (organizationId: string) =>
  request<AcademicSemester[]>(`/organizations/${organizationId}/academic/semesters`);
export const createAcademicSemester = (
  organizationId: string,
  data: Omit<AcademicSemester, "id" | "organizationId">,
) =>
  request<AcademicSemester>(`/organizations/${organizationId}/academic/semesters`, {
    method: "POST",
    body: JSON.stringify(data),
  });
export const listAcademicCourses = (organizationId: string) =>
  request<AcademicCourse[]>(`/organizations/${organizationId}/academic/courses`);
export const createAcademicCourse = (
  organizationId: string,
  data: Pick<AcademicCourse, "code" | "name" | "credits">,
) =>
  request<AcademicCourse>(`/organizations/${organizationId}/academic/courses`, {
    method: "POST",
    body: JSON.stringify(data),
  });
export const listGradeScale = (organizationId: string) =>
  request<GradeScale[]>(`/organizations/${organizationId}/academic/grade-scale`);
export const upsertAcademicResult = (
  organizationId: string,
  data: { studentId: string; semesterId: string; courseId: string; score: number },
) =>
  request<AcademicResult>(`/organizations/${organizationId}/academic/results`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
export const getTranscript = (organizationId: string, studentId: string) =>
  request<Transcript>(`/organizations/${organizationId}/academic/students/${studentId}/transcript`);
export type {
  UserRole,
  OrganizationRole,
  User,
  Session,
  SessionData,
  ManagedSession,
  Organization,
  ApiResponse,
  AcademicStudent,
  AcademicSemester,
  AcademicCourse,
  GradeScale,
  AcademicResult,
  SemesterResult,
  Transcript,
};
