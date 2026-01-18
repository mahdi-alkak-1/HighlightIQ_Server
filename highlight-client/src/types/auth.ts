export interface RegisterPayload {
  name: string;
  email: string;
  password: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface AuthUser {
  id: string;
  name: string;
  email: string;
}

export interface AuthResponse {
  user: AuthUser;
  access_token: string;
  token_type: string;
}

export type AuthErrorMap = Record<string, string>;
