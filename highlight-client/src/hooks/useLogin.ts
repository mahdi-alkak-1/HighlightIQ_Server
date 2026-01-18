import { useState } from "react";
import { loginUser } from "@/services/api/auth";
import { AuthErrorMap, AuthResponse, LoginPayload } from "@/types/auth";
import { isApiError } from "@/types/api";

export const useLogin = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<AuthErrorMap>({});

  const login = async (payload: LoginPayload): Promise<AuthResponse | null> => {
    setIsLoading(true);
    setErrorMessage(null);
    setFieldErrors({});

    try {
      const response = await loginUser(payload);
      return response;
    } catch (error) {
      if (isApiError(error)) {
        const apiMessage = error.data?.message ?? "Login failed";
        setErrorMessage(apiMessage);

        if (typeof error.data?.errors === "object" && error.data?.errors) {
          setFieldErrors(error.data.errors);
        }
      } else {
        setErrorMessage("Login failed");
      }
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    login,
    isLoading,
    errorMessage,
    fieldErrors,
  };
};
