import type { Message } from "../../../types";
import { formatTime } from "../../../shared/utils";
import "./MessageItem.css";

interface MessageItemProps {
  message: Message;
  isOwn: boolean;
}

export function MessageItem(props: MessageItemProps) {
  return (
    <div class={`message ${props.isOwn ? "message--own" : ""}`}>
      <div class="message__bubble">
        <p class="message__text">{props.message.text}</p>
        <span class="message__time">{formatTime(props.message.createdAt)}</span>
      </div>
    </div>
  );
}
