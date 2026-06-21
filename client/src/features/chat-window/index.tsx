import { Show, For, onMount, createEffect } from "solid-js";
import { useChatWindow } from "./store";
import { MessageItem } from "./components/MessageItem";
import { MessageInput } from "./components/MessageInput";
import { Loader } from "../../shared/components";
import "./ChatWindow.css";

interface ChatWindowProps {
  chatId: number;
  chatTitle: string;
  onClose: () => void;
}

export function ChatWindow(props: ChatWindowProps) {
  const { messages, isLoading, isSending, openChat, sendMessage } =
    useChatWindow();

  let scrollRef!: HTMLDivElement;

  onMount(() => {
    openChat(props.chatId, props.chatTitle);
  });

  createEffect(() => {
    messages();
    if (scrollRef) {
      scrollRef.scrollTop = scrollRef.scrollHeight;
    }
  });

  return (
    <div class="chat-window">
      <div class="chat-window__header">
        <button class="chat-window__back" onClick={props.onClose}>
          ←
        </button>
        <h3 class="chat-window__title">{props.chatTitle}</h3>
      </div>

      <div class="chat-window__messages" ref={scrollRef}>
        <Show
          when={!isLoading()}
          fallback={
            <div class="chat-window__loader">
              <Loader />
            </div>
          }
        >
          <For each={messages()}>
            {(msg) => <MessageItem message={msg} isOwn={msg.senderId === 0} />}
          </For>
          <Show when={messages().length === 0}>
            <div class="chat-window__empty">Нет сообщений</div>
          </Show>
        </Show>
      </div>

      <MessageInput onSend={sendMessage} isSending={isSending()} />
    </div>
  );
}
