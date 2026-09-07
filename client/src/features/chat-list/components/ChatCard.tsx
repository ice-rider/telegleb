import { Show } from "solid-js";
import { Avatar, Icon } from "~/shared/components";
import type { IconName } from "~/shared/components";
import { formatDate, formatLastMessage } from "~/shared/utils";
import type { Chat, MediaKind } from "~/types";
import "./ChatCard.css";

interface ChatCardProps {
  chat: Chat;
  isActive: boolean;
  onClick: () => void;
}

const MEDIA_LABEL: Record<MediaKind, { icon: IconName; text: string }> = {
  photo: { icon: "photo", text: "Фото" },
  video: { icon: "video", text: "Видео" },
  voice: { icon: "mic", text: "Голосовое сообщение" },
  audio: { icon: "music", text: "Аудио" },
  sticker: { icon: "photo", text: "Стикер" },
  gif: { icon: "video", text: "GIF" },
  document: { icon: "file", text: "Файл" },
};

export function ChatCard(props: ChatCardProps) {
  const media = () => {
    const msg = props.chat.lastMessage;
    return msg && !msg.text && msg.media ? MEDIA_LABEL[msg.media.kind] : null;
  };

  const preview = () => {
    const msg = props.chat.lastMessage;
    if (!msg) return "Нет сообщений";
    if (media()) return media()!.text;
    return (msg.out ? "Вы: " : "") + formatLastMessage(msg.text);
  };

  const typeIcon = (): IconName | null => {
    if (props.chat.isForum) return "forum";
    if (props.chat.type === "group") return "group";
    if (props.chat.type === "channel") return "channel";
    return null;
  };

  return (
    <div
      class={`chat-card ${props.isActive ? "chat-card--active" : ""}`}
      onClick={props.onClick}
    >
      <Avatar name={props.chat.title} />

      <div class="chat-card__content">
        <div class="chat-card__header">
          <Show when={props.chat.pinned}>
            <Icon name="pin" size={13} class="chat-card__flag" title="Закреплён" />
          </Show>
          <Show when={typeIcon()}>
            {(name) => <Icon name={name()} size={14} class="chat-card__flag" />}
          </Show>
          <span class="chat-card__title">{props.chat.title}</span>
          <Show when={props.chat.lastMessage}>
            {(msg) => <span class="chat-card__time mono">{formatDate(msg().createdAt)}</span>}
          </Show>
        </div>

        <div class="chat-card__preview">
          <span class="chat-card__message">
            <Show when={media()}>
              {(m) => <Icon name={m().icon} size={13} class="chat-card__media-icon" />}
            </Show>
            {preview()}
          </span>
          <Show when={props.chat.muted}>
            <Icon name="muted" size={13} class="chat-card__flag" title="Уведомления выключены" />
          </Show>
          <Show when={props.chat.unreadCount > 0 || props.chat.markedUnread}>
            <span
              class={`chat-card__badge mono ${props.chat.muted ? "chat-card__badge--muted" : ""}`}
            >
              {props.chat.unreadCount > 0 ? props.chat.unreadCount : ""}
            </span>
          </Show>
        </div>
      </div>
    </div>
  );
}
