"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getUser, isAuthenticated, type User } from "@/lib/auth";

/**
 * Client-side auth guard. Redirects to /login if the user has no session.
 * Returns the cached user object while authenticated.
 */
export function useAuth() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
      return;
    }
    setUser(getUser());
    setLoading(false);
  }, [router]);

  return { user, loading };
}
