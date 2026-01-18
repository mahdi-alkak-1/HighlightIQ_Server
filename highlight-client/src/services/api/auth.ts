import { request } from "@/services/api/client";
import { AuthResponse, LoginPayload, RegisterPayload } from "@/types/auth";

export const registerUser = async (payload: RegisterPayload) => {
  return request<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
};

export const loginUser = async (payload: LoginPayload) => {
  return request<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
};
