import { For, Show } from "solid-js";
import { formatDate, formatLastMessage } from "~/shared/utils";
import type { Topic } from "~/types";
import "./TopicList.css";

interface TopicListProps {
  topics: Topic[];
  onSelect: (topic: Topic) => void;
}

// Цвета иконок тем задаёт Telegram числом; эмодзи-иконки пока не грузим.
const ICON_COLORS: Record<number, string> = {
  0x6fb9f0: "#6fb9f0",
  0xffd67e: "#ffd67e",
  0xcb86db: "#cb86db",
  0x8eee98: "#8eee98",
  0xff93b2: "#ff93b2",
  0xfb6f5f: "#fb6f5f",
};

function colorOf(topic: Topic): string {
  return ICON_COLORS[topic.iconColor ?? 0] ?? "#6fb9f0";
}

export function TopicList(props: TopicListProps) {
  return (
    <div class="topic-list">
      <For each={props.topics}>
        {(topic) => (
          <div class="topic-card" onClick={() => props.onSelect(topic)}>
            <span class="topic-card__icon" style={{ background: colorOf(topic) }}>
              {topic.title.trim().slice(0, 1).toUpperCase()}
            </span>

            <div class="topic-card__content">
              <div class="topic-card__header">
                <span class="topic-card__title">
                  <Show when={topic.pinned}>
                    <span class="topic-card__flag" title="Закреплена">📌</span>
                  </Show>
                  <Show when={topic.closed}>
                    <span class="topic-card__flag" title="Закрыта">🔒</span>
                  </Show>
                  {topic.title}
                </span>
                <Show when={topic.lastMessage}>
                  {(msg) => <span class="topic-card__time">{formatDate(msg().createdAt)}</span>}
                </Show>
              </div>

              <div class="topic-card__preview">
                <span class="topic-card__message">
                  <Show when={topic.lastMessage} fallback="Нет сообщений">
                    {(msg) => (
                      <>
                        <Show when={msg().senderName && !msg().out}>
                          <span class="topic-card__sender">{msg().senderName}: </span>
                        </Show>
                        {formatLastMessage(msg().text || "вложение", 48)}
                      </>
                    )}
                  </Show>
                </span>
                <Show when={topic.unreadCount > 0}>
                  <span class="topic-card__badge">{topic.unreadCount}</span>
                </Show>
              </div>
            </div>
          </div>
        )}
      </For>

      <Show when={props.topics.length === 0}>
        <div class="topic-list__empty">В этой супергруппе пока нет тем</div>
      </Show>
    </div>
  );
}
