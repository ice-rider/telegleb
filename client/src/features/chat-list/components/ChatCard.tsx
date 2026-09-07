import { Show } from "solid-js";
import { Avatar } from "~/shared/components";
import { formatDate, formatLastMessage } from "~/shared/utils";
import type { Chat } from "~/types";
import "./ChatCard.css";

interface ChatCardProps {
  chat: Chat;
  isActive: boolean;
  onClick: () => void;
}

const TYPE_ICON: Record<Chat["type"], string> = {
  direct: "",
  group: "👥",
  channel: "📢",
};

export function ChatCard(props: ChatCardProps) {
  const preview = () => {
    const msg = props.chat.lastMessage;
    if (!msg) return "";
    if (!msg.text && msg.media) return mediaLabel(msg.media.kind);
    const prefix = msg.out ? "Вы: " : "";
    return prefix + formatLastMessage(msg.text);
  };

  return (
    <div
      class={`chat-card ${props.isActive ? "chat-card--active" : ""}`}
      onClick={props.onClick}
    >
      <Avatar name={props.chat.title} />
      <div class="chat-card__content">
        <div class="chat-card__header">
          <span class="chat-card__title">
            <Show when={props.chat.pinned}>
              <span class="chat-card__pin" title="Закреплён">📌</span>
            </Show>
            {TYPE_ICON[props.chat.type]} {props.chat.title}
          </span>
          <Show when={props.chat.lastMessage}>
            {(msg) => <span class="chat-card__time">{formatDate(msg().createdAt)}</span>}
          </Show>
        </div>
        <div class="chat-card__preview">
          <span class="chat-card__message">{preview()}</span>
          <Show when={props.chat.muted}>
            <span class="chat-card__muted" title="Уведомления выключены">🔕</span>
          </Show>
          <Show when={props.chat.unreadCount > 0 || props.chat.markedUnread}>
            <span class={`chat-card__badge ${props.chat.muted ? "chat-card__badge--muted" : ""}`}>
              {props.chat.unreadCount > 0 ? props.chat.unreadCount : ""}
            </span>
          </Show>
        </div>
      </div>
    </div>
  );
}

function mediaLabel(kind: string): string {
  switch (kind) {
    case "photo":
      return "📷 Фото";
    case "video":
      return "🎬 Видео";
    case "voice":
      return "🎤 Голосовое сообщение";
    case "audio":
      return "🎵 Аудио";
    case "sticker":
      return "🙂 Стикер";
    case "gif":
      return "🎞 GIF";
    default:
      return "📎 Файл";
  }
}
