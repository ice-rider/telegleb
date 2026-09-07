import { createSignal } from "solid-js";
import { api, getToken, setSessionExpiredHandler, setToken } from "~/core/api";
import { errorMessage } from "~/shared/utils";
import type { AuthStatus, Me } from "~/types";

const [status, setStatus] = createSignal<AuthStatus>("checking");
const [me, setMe] = createSignal<Me | null>(null);
const [isLoading, setIsLoading] = createSignal(false);
const [error, setError] = createSignal<string | null>(null);
const [codeType, setCodeType] = createSignal<string>("");

// Один обработчик на всё приложение: перехватчик ответов зовёт его, когда
// сервер сообщил SESSION_EXPIRED.
setSessionExpiredHandler(() => {
  setMe(null);
  setStatus("awaitingPhone");
  setError("Сессия истекла, войдите заново");
});

const CODE_HINTS: Record<string, string> = {
  app: "Код отправлен в приложение Telegram",
  sms: "Код отправлен в SMS",
  call: "Вам позвонят и продиктуют код",
  flashCall: "Введите последние цифры входящего номера",
  missedCall: "Введите последние цифры пропущенного вызова",
  fragmentSms: "Код отправлен через Fragment",
  email: "Код отправлен на почту",
};

async function run<T>(action: () => Promise<T>): Promise<T | null> {
  setIsLoading(true);
  setError(null);
  try {
    return await action();
  } catch (err) {
    setError(errorMessage(err));
    return null;
  } finally {
    setIsLoading(false);
  }
}

/**
 * Проверяет токен на старте. Без этого клиент, восстановивший протухший токен
 * из localStorage, попадает в интерфейс и падает на первом же запросе.
 */
async function bootstrap() {
  if (!getToken()) {
    setStatus("awaitingPhone");
    return;
  }
  try {
    setMe(await api.session());
    setStatus("authorized");
  } catch {
    setToken(null);
    setStatus("awaitingPhone");
  }
}

export function useAuth() {
  async function requestCode(phone: string) {
    const res = await run(() => api.requestCode(phone));
    if (res) {
      setCodeType(res.codeType);
      setStatus("awaitingCode");
    }
  }

  async function verifyCode(code: string) {
    const res = await run(() => api.verifyCode(code));
    if (!res) return;
    if (res.nextStep === "password") {
      setStatus("awaitingPassword");
      return;
    }
    setMe(res.me ?? null);
    setStatus("authorized");
  }

  async function verifyPassword(password: string) {
    const res = await run(() => api.verifyPassword(password));
    if (!res) return;
    setMe(res.me ?? null);
    setStatus("authorized");
  }

  async function logout() {
    try {
      await api.logout();
    } catch {
      // Сервер мог уже забыть сессию — локальный выход всё равно выполняем.
    }
    setMe(null);
    setError(null);
    setStatus("awaitingPhone");
  }

  return {
    status,
    me,
    isLoading,
    error,
    codeHint: () => CODE_HINTS[codeType()] ?? "Введите код подтверждения",
    bootstrap,
    requestCode,
    verifyCode,
    verifyPassword,
    logout,
  };
}
