import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AuthCard from "@/components/auth/AuthCard";
import AuthHero from "@/components/auth/AuthHero";
import AuthSwitchLink from "@/components/auth/AuthSwitchLink";
import BrandLogo from "@/components/auth/BrandLogo";
import PrimaryButton from "@/components/buttons/PrimaryButton";
import TextField from "@/components/form/TextField";
import { loginCopy } from "@/data/authCopy";
import { useLogin } from "@/hooks/useLogin";
import AuthLayout from "@/layouts/AuthLayout";

const LoginPage = () => {
  const navigate = useNavigate();
  const [formState, setFormState] = useState({
    email: "",
    password: "",
  });
  const { login, isLoading, errorMessage, fieldErrors } = useLogin();

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const response = await login(formState);
    if (response) {
      navigate("/dashboard");
    }
  };

  return (
    <AuthLayout
      left={
        <AuthHero
          title={loginCopy.heroTitle}
          subtitle={loginCopy.heroSubtitle}
          imageSrc={loginCopy.heroImage}
        />
      }
      right={
        <AuthCard>
          <div className="space-y-6">
            <BrandLogo />
            <div className="space-y-1">
              <h2 className="text-xl font-semibold text-white">
                {loginCopy.formTitle}
              </h2>
              <p className="text-sm text-white/60">{loginCopy.formSubtitle}</p>
            </div>

            {errorMessage && (
              <div className="rounded-lg border border-red-500/40 bg-red-500/10 px-3 py-2 text-xs text-red-300">
                {errorMessage}
              </div>
            )}

            <form className="space-y-4" onSubmit={handleSubmit}>
              <TextField
                label={loginCopy.fields.email.label}
                placeholder={loginCopy.fields.email.placeholder}
                name="email"
                type="email"
                value={formState.email}
                onChange={(value) =>
                  setFormState((prev) => ({ ...prev, email: value }))
                }
                error={fieldErrors.email}
                autoComplete="email"
              />
              <div className="space-y-2">
                <TextField
                  label={loginCopy.fields.password.label}
                  placeholder={loginCopy.fields.password.placeholder}
                  name="password"
                  type="password"
                  value={formState.password}
                  onChange={(value) =>
                    setFormState((prev) => ({ ...prev, password: value }))
                  }
                  error={fieldErrors.password}
                  autoComplete="current-password"
                />
                <div className="flex justify-end">
                  <Link
                    to="#"
                    className="text-xs font-semibold text-brand-link hover:underline"
                  >
                    {loginCopy.forgotPassword}
                  </Link>
                </div>
              </div>

              <PrimaryButton
                label={isLoading ? "Signing in..." : loginCopy.submitLabel}
                type="submit"
                disabled={isLoading}
              />
            </form>

            <AuthSwitchLink
              copy={loginCopy.switchCopy}
              linkText={loginCopy.switchLink}
              to="/register"
            />
          </div>
        </AuthCard>
      }
    />
  );
};

export default LoginPage;
