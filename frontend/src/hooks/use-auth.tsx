// @ts-nocheck
import { createContext, useContext, useCallback, type ReactNode } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import * as api from "@/lib/api";
import type { User } from "@/lib/api";

interface AuthContext {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<string | null>;
  register: (name: string, email: string, password: string) => Promise<string | null>;
  logout: () => Promise<void>;
}

const AuthCtx = createContext<AuthContext | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const session = useQuery({
    queryKey: ["session"],
    queryFn: async () => {
      const response = await api.getSession();
      return response.success && response.data ? response.data : null;
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  const login = useCallback(
    async (email: string, password: string) => {
      const response = await api.login(email, password);
      if (response.success && response.data) {
        queryClient.setQueryData(["session"], response.data);
        return null;
      }
      return response.message ?? "Login failed";
    },
    [queryClient],
  );

  const register = useCallback(
    async (name: string, email: string, password: string) => {
      const response = await api.register(name, email, password);
      if (response.success && response.data) {
        queryClient.setQueryData(["session"], response.data);
        return null;
      }
      return response.message ?? "Registration failed";
    },
    [queryClient],
  );

  const logout = useCallback(async () => {
    await api.logout();
    queryClient.clear();
    await navigate({ to: "/login", replace: true });
  }, [navigate, queryClient]);

  return (
    <AuthCtx
      value={{
        user: session.data?.user ?? null,
        loading: session.isLoading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthCtx>
  );
}

export function useAuth(): AuthContext {
  const context = useContext(AuthCtx);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
