import { queryOptions } from "@tanstack/react-query";
import { getSession } from "@/lib/api";

export const sessionQueryKey = ["auth", "session"] as const;

export const sessionQueryOptions = queryOptions({
  queryKey: sessionQueryKey,
  queryFn: async () => {
    const response = await getSession();
    return response.success && response.data ? response.data : null;
  },
  staleTime: 30_000,
  retry: false,
});
