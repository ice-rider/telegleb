import { For, Show, createEffect } from "solid-js";
import { Loader } from "~/shared/components";
import type { Chat } from "~/types";
import { MessageInput } from "./components/MessageInput";
import { MessageItem } from "./components/MessageItem";
import { useChatWindow } from "./store";
import "./ChatWindow.css";

export { useChatWindow } from "./store";

interface ChatWindowProps {
  chat: Chat;
  onClose: () => void;
}

export function ChatWindow(props: ChatWindowProps) {
  const { groupedMessages, isLoading, isSending, error, sendMessage } = useChatWindow();

  let scrollRef!: HTMLDivElement;

  createEffect(() => {
    groupedMessages();
    if (scrollRef) scrollRef.scrollTop = scrollRef.scrollHeight;
  });

  // Имя отправителя нужно только там, где собеседников больше одного.
  const showSender = () => props.chat.type !== "direct";

  return (
    <div class="chat-window">
      <div class="chat-window__header">
        <button class="chat-window__back" onClick={props.onClose}>
          ←
        </button>
        <h3 class="chat-window__title">{props.chat.title}</h3>
      </div>

      <div class="chat-window__messages" ref={scrollRef}>
        <Show when={!isLoading()} fallback={<div class="chat-window__loader"><Loader /></div>}>
          <Show when={error()}>
            {(message) => <div class="chat-window__error">{message()}</div>}
          </Show>

          <For each={groupedMessages()}>
            {(group) => (
              <>
                <div class="chat-window__date-separator">
                  <span>{group.date}</span>
                </div>
                <For each={group.messages}>
                  {(msg) => <MessageItem message={msg} showSender={showSender()} />}
                </For>
              </>
            )}
          </For>

          <Show when={!error() && groupedMessages().length === 0}>
            <div class="chat-window__empty">Нет сообщений</div>
          </Show>
        </Show>
      </div>

      <MessageInput onSend={sendMessage} isSending={isSending()} />
    </div>
  );
}
