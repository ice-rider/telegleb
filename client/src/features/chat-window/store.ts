import { createMemo, createSignal } from "solid-js";
import { api } from "~/core/api";
import { errorMessage, formatDayLabel } from "~/shared/utils";
import type { Chat, Message } from "~/types";

// Активный чат хранится ровно в одном месте. Раньше он дублировался в App и в
// сторе, и две копии расходились.
const [activeChat, setActiveChat] = createSignal<Chat | null>(null);
const [messages, setMessages] = createSignal<Message[]>([]);
const [isLoading, setIsLoading] = createSignal(false);
const [isSending, setIsSending] = createSignal(false);
const [error, setError] = createSignal<string | null>(null);

function randomId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

export function useChatWindow() {
  /**
   * Загрузка привязана к явному открытию чата, а не к монтированию компонента:
   * не-keyed <Show> не пересоздаёт потомка при переходе между чатами, поэтому
   * onMount второй раз не срабатывал и в новом чате оставались старые сообщения.
   */
  async function openChat(chat: Chat) {
    setActiveChat(chat);
    setMessages([]);
    setError(null);
    setIsLoading(true);
    try {
      const res = await api.history(chat);
      // Ответ мог прийти после того, как пользователь открыл другой чат.
      if (activeChat()?.ref !== chat.ref) return;
      setMessages(res.messages);
    } catch (err) {
      if (activeChat()?.ref !== chat.ref) return;
      setError(errorMessage(err));
    } finally {
      if (activeChat()?.ref === chat.ref) setIsLoading(false);
    }
  }

  function closeChat() {
    setActiveChat(null);
    setMessages([]);
    setError(null);
  }

  async function sendMessage(text: string) {
    const chat = activeChat();
    const trimmed = text.trim();
    if (!chat || !trimmed) return;

    setIsSending(true);
    try {
      const message = await api.sendMessage(chat, trimmed, randomId());
      if (activeChat()?.ref !== chat.ref) return;
      setMessages((prev) => [...prev, message]);
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setIsSending(false);
    }
  }

  const groupedMessages = createMemo(() => {
    const groups: { date: string; messages: Message[] }[] = [];
    for (const msg of messages()) {
      const label = formatDayLabel(msg.createdAt);
      const last = groups[groups.length - 1];
      if (last && last.date === label) {
        last.messages.push(msg);
      } else {
        groups.push({ date: label, messages: [msg] });
      }
    }
    return groups;
  });

  return {
    activeChat,
    messages,
    groupedMessages,
    isLoading,
    isSending,
    error,
    openChat,
    closeChat,
    sendMessage,
  };
}
