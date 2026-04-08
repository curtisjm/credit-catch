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

export async function login(email: string, password: string): Promise<User> {
  const data = await api.post<AuthResponse>("/api/auth/login", {
    body: { email, password },
  });
  tokens.set(data.access_token, data.refresh_token);
  return data.user;
}

export async function signup(
  name: string,
  email: string,
  password: string,
): Promise<User> {
  const data = await api.post<AuthResponse>("/api/auth/signup", {
    body: { name, email, password },
  });
  tokens.set(data.access_token, data.refresh_token);
  return data.user;
}

export async function logout(): Promise<void> {
  try {
    await api.post("/api/auth/logout");
  } finally {
    tokens.clear();
  }
}

export async function getMe(): Promise<User> {
  return api.get<User>("/api/auth/me");
}
