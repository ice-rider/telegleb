/**
 * Единая точка логирования на клиенте.
 *
 * Всё, что пошло не так, обязано оказаться в консоли со своим контекстом:
 * без этого «в папке 2 чата вместо 13» остаётся догадкой. Обычные console.*
 * раскиданные по коду для этого не годятся — теряется и структура, и место,
 * где ошибка возникла.
 */

const PREFIX = "[telegleb]";

function stamp(): string {
  return new Date().toISOString().slice(11, 23);
}

export const log = {
  debug(scope: string, message: string, data?: unknown) {
    if (data !== undefined) console.debug(`${PREFIX} ${stamp()} ${scope}: ${message}`, data);
    else console.debug(`${PREFIX} ${stamp()} ${scope}: ${message}`);
  },

  info(scope: string, message: string, data?: unknown) {
    if (data !== undefined) console.info(`${PREFIX} ${stamp()} ${scope}: ${message}`, data);
    else console.info(`${PREFIX} ${stamp()} ${scope}: ${message}`);
  },

  warn(scope: string, message: string, data?: unknown) {
    if (data !== undefined) console.warn(`${PREFIX} ${stamp()} ${scope}: ${message}`, data);
    else console.warn(`${PREFIX} ${stamp()} ${scope}: ${message}`);
  },

  error(scope: string, message: string, data?: unknown) {
    if (data !== undefined) console.error(`${PREFIX} ${stamp()} ${scope}: ${message}`, data);
    else console.error(`${PREFIX} ${stamp()} ${scope}: ${message}`);
  },

  /** Раскрывает группу в консоли; используется для сводок загрузки. */
  group(scope: string, title: string, body: () => void) {
    console.groupCollapsed(`${PREFIX} ${stamp()} ${scope}: ${title}`);
    try {
      body();
    } finally {
      console.groupEnd();
    }
  },
};

/**
 * Ошибки, до которых не дотянулись try/catch: исключения в обработчиках
 * событий и отвалившиеся промисы. Без этого они уходят в консоль без всякой
 * привязки к приложению, а часть браузеров показывает их невнятно.
 */
export function installGlobalErrorLogging() {
  window.addEventListener("error", (event) => {
    log.error("window", "необработанная ошибка", {
      message: event.message,
      source: `${event.filename}:${event.lineno}:${event.colno}`,
      error: event.error,
    });
  });

  window.addEventListener("unhandledrejection", (event) => {
    log.error("window", "промис отклонён и не обработан", {
      reason: event.reason,
    });
  });

  log.info("boot", "логирование ошибок включено");
}
