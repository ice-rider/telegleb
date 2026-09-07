import { ApiError } from "~/core/api";

const MESSAGES: Record<string, string> = {
  INVALID_PHONE: "Неверный формат номера",
  INVALID_CODE: "Неверный код",
  INVALID_PASSWORD: "Неверный пароль",
  INVALID_BODY: "Некорректный запрос",
  SESSION_EXPIRED: "Сессия истекла, войдите заново",
  SESSION_INVALID_STATE: "Шаг входа выполнен не по порядку",
  PEER_INVALID: "Некорректная ссылка на чат",
  PEER_NOT_FOUND: "Чат недоступен",
  MESSAGE_EMPTY: "Сообщение пустое",
  MEDIA_NOT_FOUND: "Вложение недоступно",
  TELEGRAM_ERROR: "Telegram временно недоступен",
  INTERNAL: "Внутренняя ошибка сервера",
  NETWORK_ERROR: "Нет связи с сервером",
};

/** Текст для пользователя выбирается по коду ошибки, а не по message с сервера. */
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.code === "FLOOD_WAIT") {
      const seconds = err.retryAfter ?? 0;
      return `Слишком много запросов, повторите через ${seconds} с`;
    }
    return MESSAGES[err.code] ?? "Не удалось выполнить запрос";
  }
  return "Не удалось выполнить запрос";
}
