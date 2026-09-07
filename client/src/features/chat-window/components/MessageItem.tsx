import { Show } from "solid-js";
import { formatTime } from "~/shared/utils";
import type { Message } from "~/types";
import { MessageMedia } from "./MessageMedia";
import { MessageText } from "./MessageText";
import "./MessageItem.css";

interface MessageItemProps {
  message: Message;
  showSender: boolean;
}

export function MessageItem(props: MessageItemProps) {
  // Своё сообщение определяется флагом out. Сравнение senderId с id профиля
  // ломается в каналах и до загрузки профиля.
  return (
    <div class={`message ${props.message.out ? "message--own" : ""}`}>
      <div class="message__bubble">
        <Show when={props.showSender && !props.message.out && props.message.senderName}>
          {(name) => <div class="message__sender">{name()}</div>}
        </Show>

        <Show when={props.message.forwardedFrom}>
          {(from) => <div class="message__forwarded">Переслано от {from()}</div>}
        </Show>

        <Show when={props.message.replyTo}>
          {(reply) => (
            <div class="message__reply">
              <span class="message__reply-author">{reply().senderName || "Сообщение"}</span>
              <span class="message__reply-text">{reply().text}</span>
            </div>
          )}
        </Show>

        <Show when={props.message.media}>
          {(media) => <MessageMedia media={media()} />}
        </Show>

        <Show when={props.message.text}>
          <MessageText text={props.message.text} entities={props.message.entities} />
        </Show>

        <span class="message__time">
          <Show when={props.message.editedAt}>
            <span class="message__edited">изм. </span>
          </Show>
          {formatTime(props.message.createdAt)}
        </span>
      </div>
    </div>
  );
}
