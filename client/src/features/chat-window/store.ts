import { createMemo, createSignal } from "solid-js";
import { api } from "~/core/api";
import { log } from "~/core/log";
import { errorMessage, formatDayLabel } from "~/shared/utils";
import type { Chat, Message, Topic } from "~/types";

// Активный чат хранится ровно в одном месте. Раньше он дублировался в App и в
// сторе, и две копии расходились.
const [activeChat, setActiveChat] = createSignal<Chat | null>(null);
// Активная тема форума. null означает «показываем список тем», а не «General»:
// у форума нет плоской истории, поэтому чат сам по себе не открывается.
const [activeTopic, setActiveTopic] = createSignal<Topic | null>(null);
const [topics, setTopics] = createSignal<Topic[]>([]);
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
    setActiveTopic(null);
    setTopics([]);
    setMessages([]);
    setError(null);
    setIsLoading(true);

    try {
      if (chat.isForum) {
        const list = await api.topics(chat);
        if (activeChat()?.ref !== chat.ref) return;
        setTopics(list);
        if (list.length === 0) {
          log.warn("chat", `форум ${chat.title} вернул ноль тем`);
        }
        return;
      }

      const res = await api.history(chat);
      if (activeChat()?.ref !== chat.ref) return;
      setMessages(res.messages);
    } catch (err) {
      if (activeChat()?.ref !== chat.ref) return;
      log.error("chat", `не удалось открыть ${chat.title}`, { ref: chat.ref, isForum: chat.isForum, err });
      setError(errorMessage(err));
    } finally {
      if (activeChat()?.ref === chat.ref) setIsLoading(false);
    }
  }

  async function openTopic(topic: Topic) {
    const chat = activeChat();
    if (!chat) return;

    setActiveTopic(topic);
    setMessages([]);
    setError(null);
    setIsLoading(true);
    try {
      const res = await api.history(chat, { topicId: topic.id });
      if (activeTopic()?.id !== topic.id) return;
      setMessages(res.messages);
    } catch (err) {
      if (activeTopic()?.id !== topic.id) return;
      log.error("topic", `не удалось открыть тему ${topic.title}`, { chat: chat.ref, topicId: topic.id, err });
      setError(errorMessage(err));
    } finally {
      if (activeTopic()?.id === topic.id) setIsLoading(false);
    }
  }

  /** Возврат из темы к списку тем форума. */
  function closeTopic() {
    setActiveTopic(null);
    setMessages([]);
    setError(null);
  }

  function closeChat() {
    setActiveChat(null);
    setActiveTopic(null);
    setTopics([]);
    setMessages([]);
    setError(null);
  }

  async function sendMessage(text: string) {
    const chat = activeChat();
    const trimmed = text.trim();
    if (!chat || !trimmed) return;

    const topic = activeTopic();
    setIsSending(true);
    try {
      // Без topicId сообщение в форуме уедет в General, а не в открытую тему.
      const message = await api.sendMessage(chat, trimmed, randomId(), topic?.id);
      if (activeChat()?.ref !== chat.ref || activeTopic()?.id !== topic?.id) return;
      setMessages((prev) => [...prev, message]);
    } catch (err) {
      log.error("send", "не удалось отправить сообщение", { chat: chat.ref, topicId: topic?.id, err });
      setError(errorMessage(err));
    } finally {
      setIsSending(false);
    }
  }

  const sortedTopics = createMemo(() =>
    [...topics()].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return a.order - b.order;
    }),
  );

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

  /** Форум показывает список тем, пока конкретная тема не выбрана. */
  const showingTopicList = createMemo(() => Boolean(activeChat()?.isForum) && activeTopic() === null);

  return {
    activeChat,
    activeTopic,
    topics: sortedTopics,
    showingTopicList,
    messages,
    groupedMessages,
    isLoading,
    isSending,
    error,
    openChat,
    openTopic,
    closeTopic,
    closeChat,
    sendMessage,
  };
}
