import { api, tokens } from "./api";

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface User {
  id: string;
  email: string;
  name: string;
}

const USER_KEY = "credit_catch_user";

function persistUser(user: User) {
  if (typeof window !== "undefined") {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

function clearUser() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(USER_KEY);
  }
}

/** Returns the cached user from localStorage, or null if not logged in. */
export function getUser(): User | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

/** Returns true if the user appears to be logged in (has tokens + user). */
export function isAuthenticated(): boolean {
  return tokens.access !== null || tokens.getRefresh() !== null;
}

export async function login(email: string, password: string): Promise<User> {
  const data = await api.post<AuthResponse>("/api/v1/auth/login", {
    body: { email, password },
  });
  tokens.set(data.access_token, data.refresh_token);
  persistUser(data.user);
  return data.user;
}

export async function signup(
  name: string,
  email: string,
  password: string,
): Promise<User> {
  const data = await api.post<AuthResponse>("/api/v1/auth/signup", {
    body: { name, email, password },
  });
  tokens.set(data.access_token, data.refresh_token);
  persistUser(data.user);
  return data.user;
}

export async function logout(): Promise<void> {
  const refreshToken = tokens.getRefresh();
  try {
    if (refreshToken) {
      await api.post("/api/v1/auth/logout", {
        body: { refresh_token: refreshToken },
      });
    }
  } finally {
    tokens.clear();
    clearUser();
  }
}
