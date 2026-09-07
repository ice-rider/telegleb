import { For, Show } from "solid-js";
import { Icon } from "~/shared/components";
import { formatDate, formatLastMessage } from "~/shared/utils";
import type { Topic } from "~/types";
import "./TopicList.css";

interface TopicListProps {
  topics: Topic[];
  onSelect: (topic: Topic) => void;
}

/** Цвета иконок тем задаёт Telegram числом из фиксированного набора. */
const ICON_COLORS: Record<number, string> = {
  0x6fb9f0: "#6fb9f0",
  0xffd67e: "#ffd67e",
  0xcb86db: "#cb86db",
  0x8eee98: "#8eee98",
  0xff93b2: "#ff93b2",
  0xfb6f5f: "#fb6f5f",
};

const FALLBACK = ["#c9f31d", "#7dd3fc", "#fb923c", "#e879f9", "#a3e635", "#64748b"];

function colorOf(topic: Topic): string {
  return ICON_COLORS[topic.iconColor ?? 0] ?? FALLBACK[topic.id % FALLBACK.length];
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
                <Show when={topic.pinned}>
                  <Icon name="pin" size={13} class="topic-card__flag" title="Закреплена" />
                </Show>
                <Show when={topic.closed}>
                  <Icon name="lock" size={13} class="topic-card__flag" title="Закрыта" />
                </Show>
                <span class="topic-card__title">{topic.title}</span>
                <Show when={topic.lastMessage}>
                  {(msg) => <span class="topic-card__time mono">{formatDate(msg().createdAt)}</span>}
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
                        {formatLastMessage(msg().text || "вложение", 52)}
                      </>
                    )}
                  </Show>
                </span>
                <Show when={topic.unreadCount > 0}>
                  <span class="topic-card__badge mono">{topic.unreadCount}</span>
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
