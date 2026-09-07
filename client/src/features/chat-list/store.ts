import { createMemo, createSignal } from "solid-js";
import { api } from "~/core/api";
import { errorMessage } from "~/shared/utils";
import type { Chat, Folder, Me } from "~/types";

export type Tab =
  | { kind: "all" }
  | { kind: "folder"; id: number }
  | { kind: "archive" };

/**
 * Состояние дашборда — ровно одно значение, а не отдельные сигналы под чаты и
 * папки. Так наблюдатель не может увидеть чаты без их раскладки: именно этот
 * промежуточный кадр в официальном клиенте показывает архивные и чужие для
 * вкладки чаты, которые потом исчезают.
 */
type DashboardState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; chats: Chat[]; folders: Folder[]; me: Me; nextCursor?: string }
  | { status: "error"; message: string };

const [state, setState] = createSignal<DashboardState>({ status: "idle" });
const [tab, setTab] = createSignal<Tab>({ kind: "all" });
const [query, setQuery] = createSignal("");

export function useChatList() {
  async function load() {
    setState({ status: "loading" });
    try {
      const dashboard = await api.loadDashboard();
      // Одно присваивание — чаты и папки становятся видимыми одновременно.
      setState({
        status: "ready",
        chats: dashboard.chats,
        folders: dashboard.folders,
        me: dashboard.me,
        nextCursor: dashboard.nextCursor,
      });
    } catch (err) {
      setState({ status: "error", message: errorMessage(err) });
    }
  }

  const isReady = createMemo(() => state().status === "ready");

  const folders = createMemo(() => {
    const s = state();
    return s.status === "ready" ? [...s.folders].sort((a, b) => a.order - b.order) : [];
  });

  const archivedCount = createMemo(() => {
    const s = state();
    return s.status === "ready" ? s.chats.filter((c) => c.archived).length : 0;
  });

  /**
   * Пока раскладка неизвестна, список пуст — рисовать нечего. Принадлежность
   * читается из полей самого чата, никакого соединения с отдельным списком
   * папок здесь нет и быть не должно.
   */
  const visibleChats = createMemo<Chat[]>(() => {
    const s = state();
    if (s.status !== "ready") return [];

    const current = tab();
    let list: Chat[];
    switch (current.kind) {
      case "archive":
        list = s.chats.filter((c) => c.archived);
        break;
      case "folder":
        list = s.chats.filter((c) => c.folderIds.includes(current.id));
        break;
      default:
        list = s.chats.filter((c) => !c.archived);
    }

    const q = query().trim().toLowerCase();
    if (q) {
      list = list.filter((c) => c.title.toLowerCase().includes(q));
    }

    // Порядок задаёт сервер: Telegram уже учёл закреплённые чаты и время.
    return [...list].sort((a, b) => a.order - b.order);
  });

  return {
    state,
    isReady,
    folders,
    archivedCount,
    visibleChats,
    tab,
    setTab,
    query,
    setQuery,
    load,
  };
}
