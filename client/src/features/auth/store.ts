import { createSignal, createMemo } from "solid-js";
import { api } from "../../core/api";
import type { AuthState } from "../../types";

const [authState, setAuthState] = createSignal<AuthState>({
  sessionToken: localStorage.getItem("sessionToken"),
  status: "AWAITING_PHONE",
  isLoading: false,
  error: null,
});

export const useAuth = () => {
  const isLoggedIn = createMemo(() => !!authState().sessionToken);

  async function requestLogin(phone: string) {
    setAuthState((s) => ({ ...s, isLoading: true, error: null }));
    try {
      const res = await api.requestLogin(phone);
      setAuthState((s) => ({
        ...s,
        sessionToken: res.sessionToken,
        status: "AWAITING_CODE",
        isLoading: false,
      }));
    } catch (e: any) {
      setAuthState((s) => ({
        ...s,
        isLoading: false,
        error: e.message || "Ошибка авторизации",
      }));
    }
  }

  async function verifyCode(code: string) {
    const token = authState().sessionToken;
    if (!token) return;
    setAuthState((s) => ({ ...s, isLoading: true, error: null }));
    try {
      const res = await api.verifyCode(token, code);
      if (res.nextStep === "AUTHORIZED") {
        localStorage.setItem("sessionToken", token);
        setAuthState((s) => ({
          ...s,
          status: "AUTHORIZED",
          isLoading: false,
        }));
      } else {
        setAuthState((s) => ({
          ...s,
          status: "AWAITING_PASSWORD",
          isLoading: false,
        }));
      }
    } catch (e: any) {
      setAuthState((s) => ({
        ...s,
        isLoading: false,
        error: e.message || "Неверный код",
      }));
    }
  }

  async function verifyPassword(password: string) {
    const token = authState().sessionToken;
    if (!token) return;
    setAuthState((s) => ({ ...s, isLoading: true, error: null }));
    try {
      await api.verifyPassword(token, password);
      localStorage.setItem("sessionToken", token);
      setAuthState((s) => ({
        ...s,
        status: "AUTHORIZED",
        isLoading: false,
      }));
    } catch (e: any) {
      setAuthState((s) => ({
        ...s,
        isLoading: false,
        error: e.message || "Неверный пароль",
      }));
    }
  }

  async function logout() {
    const token = authState().sessionToken;
    if (token) {
      try {
        await api.logout(token);
      } catch {
        // ignore
      }
    }
    localStorage.removeItem("sessionToken");
    setAuthState({
      sessionToken: null,
      status: "AWAITING_PHONE",
      isLoading: false,
      error: null,
    });
  }

  function initFromStorage() {
    const token = localStorage.getItem("sessionToken");
    if (token) {
      api.setToken(token);
      setAuthState((s) => ({
        ...s,
        sessionToken: token,
        status: "AUTHORIZED",
      }));
    }
  }

  return {
    state: authState,
    isLoggedIn,
    requestLogin,
    verifyCode,
    verifyPassword,
    logout,
    initFromStorage,
  };
};
