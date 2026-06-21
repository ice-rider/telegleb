import { createSignal } from "solid-js";
import { api } from "../../core/api";
import type { Message } from "../../types";

const [messages, setMessages] = createSignal<Message[]>([]);
const [activeChatId, setActiveChatId] = createSignal<number | null>(null);
const [chatTitle, setChatTitle] = createSignal("");
const [isLoading, setIsLoading] = createSignal(false);
const [isSending, setIsSending] = createSignal(false);
const [error, setError] = createSignal<string | null>(null);

export const useChatWindow = () => {
  async function openChat(chatId: number, title: string) {
    setActiveChatId(chatId);
    setChatTitle(title);
    setMessages([]);
    setIsLoading(true);
    setError(null);
    try {
      const msgs = await api.openChat(chatId, 50, 0);
      setMessages(msgs);
    } catch (e: any) {
      setError(e.message || "Ошибка загрузки сообщений");
    } finally {
      setIsLoading(false);
    }
  }

  async function sendMessage(content: string) {
    const chatId = activeChatId();
    if (!chatId || !content.trim()) return;
    setIsSending(true);
    try {
      const msg = await api.sendMessage(chatId, content.trim());
      setMessages((prev) => [...prev, msg]);
    } catch (e: any) {
      setError(e.message || "Ошибка отправки");
    } finally {
      setIsSending(false);
    }
  }

  function closeChat() {
    setActiveChatId(null);
    setMessages([]);
    setChatTitle("");
  }

  return {
    messages,
    activeChatId,
    chatTitle,
    isLoading,
    isSending,
    error,
    openChat,
    sendMessage,
    closeChat,
  };
};
