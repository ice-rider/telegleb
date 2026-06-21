import type { Chat } from "../../../types";
import { Avatar } from "../../../shared/components";
import { formatLastMessage, formatDate } from "../../../shared/utils";
import "./ChatCard.css";

interface ChatCardProps {
  chat: Chat;
  isActive: boolean;
  onClick: () => void;
}

export function ChatCard(props: ChatCardProps) {
  const typeIcon = () => {
    switch (props.chat.type) {
      case "group":
        return "👥";
      case "channel":
        return "📢";
      default:
        return "";
    }
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
            {typeIcon()} {props.chat.title}
          </span>
          <span class="chat-card__time">
            {formatDate(props.chat.lastMessage.createdAt)}
          </span>
        </div>
        <div class="chat-card__preview">
          <span class="chat-card__message">
            {formatLastMessage(props.chat.lastMessage.text)}
          </span>
          {props.chat.unreadCount > 0 && (
            <span class="chat-card__badge">{props.chat.unreadCount}</span>
          )}
        </div>
      </div>
    </div>
  );
}
