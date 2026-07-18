import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  AlertCircleIcon,
  BookOpenIcon,
  CalculatorIcon,
  CalendarDaysIcon,
  GraduationCapIcon,
  UserPlusIcon,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
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

export const Route = createFileRoute("/_protected/academic")({ component: AcademicPage });

function AcademicPage() {
  const { session } = useAuth();
  const queryClient = useQueryClient();
  const organizationId = session?.session.activeOrganizationId;
  const role = session?.session.activeOrganizationRole;
  const membership = useQuery({
    queryKey: ["organization", organizationId, "member", "me"],
    enabled: Boolean(organizationId),
    queryFn: async () => {
      const response = await api.getCurrentOrganizationMember(organizationId!);
      if (!response.success || !response.data) return null;
      return response.data;
    },
  });
  const fullAccess = role === "owner" || role === "admin";
  const permissions = membership.data?.permissions ?? [];
  const canStudents = fullAccess || permissions.includes("academic.students.manage");
  const canStructure = fullAccess || permissions.includes("academic.structure.manage");
  const canResults = fullAccess || permissions.includes("academic.results.manage");
  const canReadStudents = canStudents || canResults;
  const canReadStructure = canStructure || canResults;
  const canManage = canStudents || canStructure || canResults;
  const [selectedStudent, setSelectedStudent] = useState("");
  const [studentForm, setStudentForm] = useState({
    studentNo: "",
    name: "",
    email: "",
    program: "",
    intake: "",
  });
  const [semesterForm, setSemesterForm] = useState({
    code: "",
    name: "",
    academicYear: String(new Date().getFullYear()),
    sequence: "1",
  });
  const [courseForm, setCourseForm] = useState({ code: "", name: "", credits: "" });
  const [resultForm, setResultForm] = useState({
    studentId: "",
    semesterId: "",
    courseId: "",
    score: "",
  });

  const students = useQuery({
    queryKey: ["academic", organizationId, "students"],
    enabled: Boolean(organizationId && canReadStudents),
    queryFn: () => load(api.listAcademicStudents(organizationId!)),
  });
  const semesters = useQuery({
    queryKey: ["academic", organizationId, "semesters"],
    enabled: Boolean(organizationId && canReadStructure),
    queryFn: () => load(api.listAcademicSemesters(organizationId!)),
  });
  const courses = useQuery({
    queryKey: ["academic", organizationId, "courses"],
    enabled: Boolean(organizationId && canReadStructure),
    queryFn: () => load(api.listAcademicCourses(organizationId!)),
  });
  const gradeScale = useQuery({
    queryKey: ["academic", organizationId, "grade-scale"],
    enabled: Boolean(organizationId && canReadStructure),
    queryFn: () => load(api.listGradeScale(organizationId!)),
  });
  const transcript = useQuery({
    queryKey: ["academic", organizationId, "transcript", selectedStudent],
    enabled: Boolean(organizationId && selectedStudent && canResults),
    queryFn: () => load(api.getTranscript(organizationId!, selectedStudent)),
  });

  const createStudent = useMutation({
    mutationFn: () => load(api.createAcademicStudent(organizationId!, studentForm)),
    onSuccess: (student) => {
      queryClient.setQueryData<api.AcademicStudent[]>(
        ["academic", organizationId, "students"],
        (current = []) => [...current, student].sort((a, b) => a.name.localeCompare(b.name)),
      );
      setStudentForm({ studentNo: "", name: "", email: "", program: "", intake: "" });
    },
  });
  const createSemester = useMutation({
    mutationFn: () =>
      load(
        api.createAcademicSemester(organizationId!, {
          code: semesterForm.code,
          name: semesterForm.name,
          academicYear: Number(semesterForm.academicYear),
          sequence: Number(semesterForm.sequence),
          status: "active",
        }),
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["academic", organizationId, "semesters"] });
      setSemesterForm({
        code: "",
        name: "",
        academicYear: String(new Date().getFullYear()),
        sequence: "1",
      });
    },
  });
  const createCourse = useMutation({
    mutationFn: () =>
      load(
        api.createAcademicCourse(organizationId!, {
          code: courseForm.code,
          name: courseForm.name,
          credits: Number(courseForm.credits),
        }),
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["academic", organizationId, "courses"] });
      setCourseForm({ code: "", name: "", credits: "" });
    },
  });
  const saveResult = useMutation({
    mutationFn: () =>
      load(
        api.upsertAcademicResult(organizationId!, {
          studentId: resultForm.studentId,
          semesterId: resultForm.semesterId,
          courseId: resultForm.courseId,
          score: Number(resultForm.score),
        }),
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["academic", organizationId, "transcript", resultForm.studentId],
      });
      setSelectedStudent(resultForm.studentId);
      setResultForm((current) => ({ ...current, courseId: "", score: "" }));
    },
  });

  if (!organizationId) {
    return (
      <Empty className="flex-1">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <GraduationCapIcon />
          </EmptyMedia>
          <EmptyTitle>Select an institute</EmptyTitle>
          <EmptyDescription>
            Choose an organization from the sidebar before managing academic records.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  if (!fullAccess && membership.isPending) {
    return <Skeleton className="h-48 w-full" />;
  }

  if (!canManage) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>Access restricted</AlertTitle>
        <AlertDescription>
          Your organization role does not include academic management permissions.
        </AlertDescription>
      </Alert>
    );
  }

  const queryError = students.error ?? semesters.error ?? courses.error ?? gradeScale.error;
  const score = Number(resultForm.score);
  const gradePreview = gradeScale.data?.find(
    (item) => Number.isFinite(score) && score >= item.minScore && score <= item.maxScore,
  );

  return (
    <div className="flex w-full min-w-0 max-w-full flex-col gap-4">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-balance">Academic records</h1>
        <p className="text-sm text-pretty text-muted-foreground">
          Manage students, courses, semesters and CGPA for the active institute.
        </p>
      </div>

      {queryError && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Unable to load academic data</AlertTitle>
          <AlertDescription>{queryError.message}</AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 sm:grid-cols-3">
        <MetricCard label="Students" value={students.data?.length} icon={UserPlusIcon} />
        <MetricCard label="Courses" value={courses.data?.length} icon={BookOpenIcon} />
        <MetricCard label="Semesters" value={semesters.data?.length} icon={CalendarDaysIcon} />
      </div>

      <div className="grid min-w-0 gap-4 xl:grid-cols-3">
        {canStudents && (
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Add student</CardTitle>
              <CardDescription>Create a student record for this institute.</CardDescription>
            </CardHeader>
            <form
              className="flex flex-1 flex-col gap-(--card-spacing)"
              onSubmit={(event) => submit(event, createStudent.mutate)}
            >
              <CardContent className="flex-1">
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="student-no">Student number</FieldLabel>
                    <Input
                      id="student-no"
                      required
                      value={studentForm.studentNo}
                      onChange={(event) =>
                        setStudentForm((current) => ({ ...current, studentNo: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="student-name">Full name</FieldLabel>
                    <Input
                      id="student-name"
                      required
                      value={studentForm.name}
                      onChange={(event) =>
                        setStudentForm((current) => ({ ...current, name: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="student-email">Email (optional)</FieldLabel>
                    <Input
                      id="student-email"
                      type="email"
                      value={studentForm.email}
                      onChange={(event) =>
                        setStudentForm((current) => ({ ...current, email: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="student-program">Program</FieldLabel>
                    <Input
                      id="student-program"
                      required
                      value={studentForm.program}
                      onChange={(event) =>
                        setStudentForm((current) => ({ ...current, program: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="student-intake">Intake</FieldLabel>
                    <Input
                      id="student-intake"
                      required
                      placeholder="e.g. 1/2026"
                      value={studentForm.intake}
                      onChange={(event) =>
                        setStudentForm((current) => ({ ...current, intake: event.target.value }))
                      }
                    />
                  </Field>
                </FieldGroup>
                <MutationAlert mutation={createStudent} success="Student added." />
              </CardContent>
              <CardFooter className="border-t">
                <SubmitButton pending={createStudent.isPending}>Add student</SubmitButton>
              </CardFooter>
            </form>
          </Card>
        )}

        {canStructure && (
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Add semester</CardTitle>
              <CardDescription>Define an academic semester for results.</CardDescription>
            </CardHeader>
            <form
              className="flex flex-1 flex-col gap-(--card-spacing)"
              onSubmit={(event) => submit(event, createSemester.mutate)}
            >
              <CardContent className="flex-1">
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="semester-code">Code</FieldLabel>
                    <Input
                      id="semester-code"
                      required
                      placeholder="2026-S1"
                      value={semesterForm.code}
                      onChange={(event) =>
                        setSemesterForm((current) => ({ ...current, code: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="semester-name">Name</FieldLabel>
                    <Input
                      id="semester-name"
                      required
                      placeholder="Semester 1"
                      value={semesterForm.name}
                      onChange={(event) =>
                        setSemesterForm((current) => ({ ...current, name: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="academic-year">Academic year</FieldLabel>
                    <Input
                      id="academic-year"
                      type="number"
                      min="2000"
                      max="2200"
                      required
                      value={semesterForm.academicYear}
                      onChange={(event) =>
                        setSemesterForm((current) => ({
                          ...current,
                          academicYear: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="semester-sequence">Sequence</FieldLabel>
                    <Input
                      id="semester-sequence"
                      type="number"
                      min="1"
                      max="20"
                      required
                      value={semesterForm.sequence}
                      onChange={(event) =>
                        setSemesterForm((current) => ({ ...current, sequence: event.target.value }))
                      }
                    />
                  </Field>
                </FieldGroup>
                <MutationAlert mutation={createSemester} success="Semester added." />
              </CardContent>
              <CardFooter className="border-t">
                <SubmitButton pending={createSemester.isPending}>Add semester</SubmitButton>
              </CardFooter>
            </form>
          </Card>
        )}

        {canStructure && (
          <Card className="min-w-0">
            <CardHeader>
              <CardTitle>Add course</CardTitle>
              <CardDescription>Courses contribute credits toward GPA and CGPA.</CardDescription>
            </CardHeader>
            <form
              className="flex flex-1 flex-col gap-(--card-spacing)"
              onSubmit={(event) => submit(event, createCourse.mutate)}
            >
              <CardContent className="flex-1">
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="course-code">Course code</FieldLabel>
                    <Input
                      id="course-code"
                      required
                      placeholder="TEC101"
                      value={courseForm.code}
                      onChange={(event) =>
                        setCourseForm((current) => ({ ...current, code: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="course-name">Course name</FieldLabel>
                    <Input
                      id="course-name"
                      required
                      value={courseForm.name}
                      onChange={(event) =>
                        setCourseForm((current) => ({ ...current, name: event.target.value }))
                      }
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="course-credits">Credits</FieldLabel>
                    <Input
                      id="course-credits"
                      type="number"
                      min="0.5"
                      max="30"
                      step="0.5"
                      required
                      value={courseForm.credits}
                      onChange={(event) =>
                        setCourseForm((current) => ({ ...current, credits: event.target.value }))
                      }
                    />
                  </Field>
                </FieldGroup>
                <MutationAlert mutation={createCourse} success="Course added." />
              </CardContent>
              <CardFooter className="border-t">
                <SubmitButton pending={createCourse.isPending}>Add course</SubmitButton>
              </CardFooter>
            </form>
          </Card>
        )}
      </div>

      {canResults && (
        <Card className="min-w-0">
          <CardHeader>
            <CardTitle>Enter result</CardTitle>
            <CardDescription>
              The grade and quality points are calculated automatically from the score.
            </CardDescription>
            {gradePreview && (
              <CardAction>
                <Badge variant={gradePreview.passing ? "secondary" : "destructive"}>
                  {gradePreview.letter} · {gradePreview.gradePoint.toFixed(2)}
                </Badge>
              </CardAction>
            )}
          </CardHeader>
          <form onSubmit={(event) => submit(event, saveResult.mutate)}>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <Field>
                  <FieldLabel>Student</FieldLabel>
                  <EntitySelect
                    value={resultForm.studentId}
                    placeholder="Select student"
                    items={(students.data ?? []).map((item) => ({
                      value: item.id,
                      label: `${item.studentNo} · ${item.name}`,
                    }))}
                    onChange={(studentId) =>
                      setResultForm((current) => ({ ...current, studentId }))
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>Semester</FieldLabel>
                  <EntitySelect
                    value={resultForm.semesterId}
                    placeholder="Select semester"
                    items={(semesters.data ?? []).map((item) => ({
                      value: item.id,
                      label: `${item.code} · ${item.name}`,
                    }))}
                    onChange={(semesterId) =>
                      setResultForm((current) => ({ ...current, semesterId }))
                    }
                  />
                </Field>
                <Field>
                  <FieldLabel>Course</FieldLabel>
                  <EntitySelect
                    value={resultForm.courseId}
                    placeholder="Select course"
                    items={(courses.data ?? []).map((item) => ({
                      value: item.id,
                      label: `${item.code} · ${item.name}`,
                    }))}
                    onChange={(courseId) => setResultForm((current) => ({ ...current, courseId }))}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="result-score">Score</FieldLabel>
                  <Input
                    id="result-score"
                    type="number"
                    min="0"
                    max="100"
                    step="0.01"
                    required
                    value={resultForm.score}
                    onChange={(event) =>
                      setResultForm((current) => ({ ...current, score: event.target.value }))
                    }
                  />
                </Field>
              </div>
              <MutationAlert mutation={saveResult} success="Result saved and CGPA updated." />
            </CardContent>
            <CardFooter className="justify-end border-t">
              <SubmitButton
                pending={saveResult.isPending}
                disabled={
                  !resultForm.studentId ||
                  !resultForm.semesterId ||
                  !resultForm.courseId ||
                  resultForm.score === ""
                }
              >
                Save result
              </SubmitButton>
            </CardFooter>
          </form>
        </Card>
      )}

      {canResults && (
        <Card className="min-w-0">
          <CardHeader>
            <CardTitle>Student transcript</CardTitle>
            <CardDescription>Review semester GPA and cumulative CGPA.</CardDescription>
            {transcript.data && (
              <CardAction>
                <Badge>CGPA {transcript.data.cgpa.toFixed(2)}</Badge>
              </CardAction>
            )}
          </CardHeader>
          <CardContent className="flex min-w-0 flex-col gap-4">
            <EntitySelect
              value={selectedStudent}
              placeholder="Select a student"
              items={(students.data ?? []).map((item) => ({
                value: item.id,
                label: `${item.studentNo} · ${item.name}`,
              }))}
              onChange={setSelectedStudent}
            />
            {transcript.isPending && selectedStudent ? (
              <Skeleton className="h-40 w-full" />
            ) : transcript.error ? (
              <Alert variant="destructive">
                <AlertCircleIcon />
                <AlertTitle>Unable to load transcript</AlertTitle>
                <AlertDescription>{transcript.error.message}</AlertDescription>
              </Alert>
            ) : transcript.data?.semesters.length ? (
              <div className="flex min-w-0 flex-col gap-4">
                {transcript.data.semesters.map((semester) => (
                  <div
                    key={semester.semesterId}
                    className="min-w-0 overflow-hidden rounded-md border"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2 bg-muted/40 px-4 py-3">
                      <div>
                        <p className="font-medium">{semester.semesterName}</p>
                        <p className="text-xs text-muted-foreground">{semester.semesterCode}</p>
                      </div>
                      <div className="flex gap-2">
                        <Badge variant="outline">GPA {semester.gpa.toFixed(2)}</Badge>
                        <Badge variant="secondary">CGPA {semester.cgpa.toFixed(2)}</Badge>
                      </div>
                    </div>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Course</TableHead>
                          <TableHead className="text-right">Credits</TableHead>
                          <TableHead className="text-right">Score</TableHead>
                          <TableHead className="text-right">Grade</TableHead>
                          <TableHead className="text-right">Points</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {semester.results.map((result) => (
                          <TableRow key={result.id}>
                            <TableCell>
                              <div className="min-w-48">
                                <p className="font-medium">{result.courseCode}</p>
                                <p className="truncate text-xs text-muted-foreground">
                                  {result.courseName}
                                </p>
                              </div>
                            </TableCell>
                            <TableCell className="text-right">{result.credits}</TableCell>
                            <TableCell className="text-right">{result.score.toFixed(2)}</TableCell>
                            <TableCell className="text-right">
                              <Badge variant="outline">{result.grade}</Badge>
                            </TableCell>
                            <TableCell className="text-right">
                              {result.qualityPoint.toFixed(2)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                ))}
              </div>
            ) : selectedStudent ? (
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <CalculatorIcon />
                  </EmptyMedia>
                  <EmptyTitle>No results yet</EmptyTitle>
                  <EmptyDescription>
                    Enter the first course result for this student.
                  </EmptyDescription>
                </EmptyHeader>
              </Empty>
            ) : null}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function MetricCard({
  label,
  value,
  icon: Icon,
}: {
  label: string;
  value: number | undefined;
  icon: typeof UserPlusIcon;
}) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{label}</CardTitle>
        <CardAction>
          <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
        </CardAction>
      </CardHeader>
      <CardContent>
        {value === undefined ? (
          <Skeleton className="h-7 w-12" />
        ) : (
          <p className="text-2xl font-semibold tracking-tight">{value}</p>
        )}
      </CardContent>
    </Card>
  );
}

function EntitySelect({
  value,
  placeholder,
  items,
  onChange,
}: {
  value: string;
  placeholder: string;
  items: { value: string; label: string }[];
  onChange: (value: string) => void;
}) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}

function SubmitButton({
  pending,
  disabled,
  children,
}: {
  pending: boolean;
  disabled?: boolean;
  children: string;
}) {
  return (
    <Button type="submit" className="w-full" disabled={pending || disabled}>
      {pending && <Spinner data-icon="inline-start" />}
      {children}
    </Button>
  );
}

function MutationAlert({
  mutation,
  success,
}: {
  mutation: { error: Error | null; isSuccess: boolean };
  success: string;
}) {
  if (mutation.error) {
    return (
      <Alert variant="destructive" className="mt-4">
        <AlertCircleIcon />
        <AlertTitle>Unable to save</AlertTitle>
        <AlertDescription>{mutation.error.message}</AlertDescription>
      </Alert>
    );
  }
  if (mutation.isSuccess) {
    return (
      <Alert className="mt-4">
        <AlertTitle>Saved</AlertTitle>
        <AlertDescription>{success}</AlertDescription>
      </Alert>
    );
  }
  return null;
}

function submit(event: FormEvent, action: () => void) {
  event.preventDefault();
  action();
}

async function load<T>(response: Promise<api.ApiResponse<T>>) {
  const result = await response;
  if (!result.success || result.data === undefined) {
    throw new Error(result.message ?? "Unable to complete the request");
  }
  return result.data;
}
