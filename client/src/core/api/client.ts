import axios, { AxiosError } from "axios";
import { log } from "~/core/log";
import type { Chat, Dashboard, Me, Message, NextStep, Topic } from "~/types";

const TOKEN_STORAGE_KEY = "telegleb.sessionToken";

/** Ошибка контракта: код машиночитаемый, ветвиться нужно по нему, не по тексту. */
export class ApiError extends Error {
  readonly code: string;
  readonly retryAfter?: number;

  constructor(code: string, message: string, retryAfter?: number) {
    super(message);
    this.code = code;
    this.retryAfter = retryAfter;
  }
}

let sessionToken: string | null = null;
try {
  sessionToken = localStorage.getItem(TOKEN_STORAGE_KEY);
} catch {
  sessionToken = null;
}

let onSessionExpired: (() => void) | null = null;

export function setSessionExpiredHandler(handler: () => void) {
  onSessionExpired = handler;
}

export function getToken(): string | null {
  return sessionToken;
}

export function setToken(token: string | null) {
  sessionToken = token;
  try {
    if (token) localStorage.setItem(TOKEN_STORAGE_KEY, token);
    else localStorage.removeItem(TOKEN_STORAGE_KEY);
  } catch {
    // Приватный режим и заблокированное хранилище не должны ронять вход.
  }
}

const http = axios.create({
  baseURL: "/api/v1",
  headers: { "Content-Type": "application/json" },
});

// Токен ходит только заголовком: в query он оседает в access-логах, в теле
// дублирует то, что уже есть в заголовке.
http.interceptors.request.use((config) => {
  if (sessionToken) {
    config.headers.Authorization = `Bearer ${sessionToken}`;
  }
  return config;
});

http.interceptors.response.use(
  (res) => res,
  (err: AxiosError<{ error?: { code: string; message: string; retryAfter?: number } }>) => {
    const payload = err.response?.data?.error;
    const apiError = new ApiError(
      payload?.code ?? "NETWORK_ERROR",
      payload?.message ?? err.message ?? "Не удалось выполнить запрос",
      payload?.retryAfter,
    );

    // Каждая ошибка API попадает в консоль с полным контекстом — методом,
    // адресом, статусом и кодом. Это единственное место, где видно все
    // сетевые сбои сразу.
    log.error("api", `${apiError.code} — ${apiError.message}`, {
      method: err.config?.method?.toUpperCase(),
      url: err.config?.url,
      params: err.config?.params,
      status: err.response?.status,
      retryAfter: apiError.retryAfter,
      body: err.response?.data,
    });

    // Разлогин живёт здесь, а не в каждом вызове: иначе клиент с протухшим
    // токеном остаётся в интерфейсе и просто показывает ошибки.
    if (apiError.code === "SESSION_EXPIRED") {
      setToken(null);
      onSessionExpired?.();
    }

    return Promise.reject(apiError);
  },
);

export interface RequestCodeResponse {
  sessionToken: string;
  nextStep: NextStep;
  codeType: string;
  timeout?: number;
}

export interface AuthStepResponse {
  nextStep: NextStep;
  me?: Me;
}

export interface HistoryResponse {
  messages: Message[];
  nextBeforeId?: number;
}

/**
 * Сводка загрузки дашборда. Отвечает на вопрос «почему в папке мало чатов»
 * числами: сколько диалогов реально пришло, сколько сервер отбросил и
 * сколько участников получила каждая папка.
 */
function logDashboard(data: Dashboard) {
  log.group("dashboard", `чатов ${data.chats.length}, папок ${data.folders.length}`, () => {
    const archived = data.chats.filter((c) => c.archived).length;
    console.info("всего:", data.chats.length, "| в архиве:", archived, "| форумов:", data.chats.filter((c) => c.isForum).length);
    console.info("сервер:", data.stats);

    const rows = data.folders.map((f) => ({
      папка: f.title,
      id: f.id,
      чатов: data.chats.filter((c) => c.folderIds.includes(f.id)).length,
    }));
    rows.push({ папка: "Все (без архива)", id: 0, чатов: data.chats.length - archived });
    rows.push({ папка: "Архив", id: -1, чатов: archived });
    console.table(rows);

    const orphans = data.chats.filter((c) => !c.archived && c.folderIds.length === 0).length;
    if (orphans) console.info("чатов вне пользовательских папок:", orphans);
  });

  if (data.stats.skippedUnknownPeer > 0) {
    log.warn("dashboard", `сервер отбросил ${data.stats.skippedUnknownPeer} диалогов: пир не найден в словарях ответа Telegram`);
  }
  if (data.truncated) {
    log.warn("dashboard", "выборка диалогов упёрлась в потолок — раскладка по папкам может быть неполной");
  }
}

export const api = {
  async requestCode(phone: string): Promise<RequestCodeResponse> {
    const { data } = await http.post<RequestCodeResponse>("/auth/request-code", { phone });
    setToken(data.sessionToken);
    return data;
  },

  async verifyCode(code: string): Promise<AuthStepResponse> {
    const { data } = await http.post<AuthStepResponse>("/auth/verify-code", { code });
    return data;
  },

  async verifyPassword(password: string): Promise<AuthStepResponse> {
    const { data } = await http.post<AuthStepResponse>("/auth/verify-password", { password });
    return data;
  },

  async session(): Promise<Me> {
    const { data } = await http.get<{ status: string; me: Me }>("/auth/session");
    return data.me;
  },

  async logout(): Promise<void> {
    try {
      await http.post("/auth/logout");
    } finally {
      setToken(null);
    }
  },

  /** Чаты, папки и профиль приезжают одним ответом — по контракту это атомарно. */
  async loadDashboard(): Promise<Dashboard> {
    const { data } = await http.get<Dashboard>("/chats");
    logDashboard(data);
    return data;
  },

  async topics(chat: Chat): Promise<Topic[]> {
    const { data } = await http.get<{ topics: Topic[] }>(
      `/chats/${encodeURIComponent(chat.ref)}/topics`,
    );
    log.info("topics", `${chat.title}: тем ${data.topics.length}`);
    return data.topics;
  },

  async history(chat: Chat, opts: { limit?: number; beforeId?: number; topicId?: number } = {}): Promise<HistoryResponse> {
    const { data } = await http.get<HistoryResponse>(
      `/chats/${encodeURIComponent(chat.ref)}/messages`,
      { params: { limit: opts.limit ?? 50, beforeId: opts.beforeId, topicId: opts.topicId } },
    );
    return data;
  },

  async sendMessage(chat: Chat, text: string, randomId: string, topicId?: number): Promise<Message> {
    const { data } = await http.post<{ message: Message }>(
      `/chats/${encodeURIComponent(chat.ref)}/messages`,
      { text, randomId, topicId },
    );
    return data.message;
  },

  /**
   * Медиа скачивается через fetch, а не через <img src>: адрес требует
   * заголовка Authorization, а тег его отправить не может.
   */
  async mediaObjectUrl(ref: string): Promise<string> {
    const res = await fetch(`/api/v1/media/${encodeURIComponent(ref)}`, {
      headers: sessionToken ? { Authorization: `Bearer ${sessionToken}` } : {},
    });
    if (!res.ok) {
      throw new ApiError("MEDIA_NOT_FOUND", "Не удалось загрузить вложение");
    }
    return URL.createObjectURL(await res.blob());
  },
};
