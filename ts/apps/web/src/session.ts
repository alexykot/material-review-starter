import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authClient, isUnauthenticated } from "./api.ts";

const meKey = ["me"] as const;

// useMe returns the signed-in user, or null when signed out.
export function useMe() {
  return useQuery({
    queryKey: meKey,
    queryFn: async () => {
      try {
        const { user } = await authClient.getMe({});
        return user ?? null;
      } catch (error) {
        if (isUnauthenticated(error)) return null;
        throw error;
      }
    },
    retry: false,
  });
}

export function useSignIn() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { email: string; password: string }) => authClient.signIn(input),
    onSuccess: ({ user }) => queryClient.setQueryData(meKey, user ?? null),
  });
}

export function useSignOut() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => authClient.signOut({}),
    onSuccess: () => queryClient.setQueryData(meKey, null),
  });
}
