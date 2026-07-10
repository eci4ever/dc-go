import { createContext, useCallback, useContext, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import * as api from "@/lib/api";
import { sessionQueryKey, sessionQueryOptions } from "@/lib/session";
import type { SessionData, User } from "@/lib/api";

interface AuthContextValue {
  user: User | null;
  session: SessionData | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<string | null>;
  register: (name: string, email: string, password: string) => Promise<string | null>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const session = useQuery(sessionQueryOptions);

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      api.login(email, password),
  });
  const registerMutation = useMutation({
    mutationFn: ({ name, email, password }: { name: string; email: string; password: string }) =>
      api.register(name, email, password),
  });
  const logoutMutation = useMutation({ mutationFn: api.logout });

  const login = useCallback(
    async (email: string, password: string) => {
      const response = await loginMutation.mutateAsync({ email, password });
      if (!response.success || !response.data) return response.message ?? "Login failed";
      queryClient.setQueryData(sessionQueryKey, response.data);
      return null;
    },
    [loginMutation, queryClient],
  );

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      const response = await registerMutation.mutateAsync({ name, email, password });
      if (!response.success || !response.data) return response.message ?? "Registration failed";
      queryClient.setQueryData(sessionQueryKey, response.data);
      return null;
    },
    [registerMutation, queryClient],
  );

  const logout = useCallback(async () => {
    await logoutMutation.mutateAsync();
    queryClient.setQueryData(sessionQueryKey, null);
    queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== "auth" });
    await navigate({ to: "/login", replace: true });
  }, [logoutMutation, navigate, queryClient]);

  return (
    <AuthContext.Provider
      value={{
        user: session.data?.user ?? null,
        session: session.data ?? null,
        loading: session.isPending,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
