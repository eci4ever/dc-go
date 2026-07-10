// @ts-nocheck
import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import * as api from "@/lib/api";
import type { User } from "@/lib/api";

interface AuthState {
  user: User | null;
  loading: boolean;
}

interface AuthContext extends AuthState {
  login: (email: string, password: string) => Promise<string | null>;
  register: (name: string, email: string, password: string) => Promise<string | null>;
  logout: () => Promise<void>;
}

const AuthCtx = createContext<AuthContext | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ user: null, loading: true });

  const fetchSession = useCallback(async () => {
    const res = await api.getSession();
    if (res.success && res.data) {
      setState({ user: res.data.user, loading: false });
    } else {
      setState({ user: null, loading: false });
    }
  }, []);

  useEffect(() => {
    fetchSession();
  }, [fetchSession]);

  const login = useCallback(async (email: string, password: string): Promise<string | null> => {
    const res = await api.login(email, password);
    if (res.success && res.data) {
      setState({ user: res.data.user, loading: false });
      return null;
    }
    return res.message ?? "Login failed";
  }, []);

  const register = useCallback(
    async (name: string, email: string, password: string): Promise<string | null> => {
      const res = await api.register(name, email, password);
      if (res.success && res.data) {
        setState({ user: res.data.user, loading: false });
        return null;
      }
      return res.message ?? "Registration failed";
    },
    [],
  );

  const logout = useCallback(async () => {
    await api.logout();
    setState({ user: null, loading: false });
    window.location.href = "/login";
  }, []);

  return <AuthCtx value={{ ...state, login, register, logout }}>{children}</AuthCtx>;
}

export function useAuth(): AuthContext {
  const ctx = useContext(AuthCtx);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
