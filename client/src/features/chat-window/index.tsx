import { Show, For, onMount, createEffect } from "solid-js";
import { useChatWindow } from "./store";
import { MessageItem } from "./components/MessageItem";
import { MessageInput } from "./components/MessageInput";
import { Loader } from "../../shared/components";
import "./ChatWindow.css";

interface ChatWindowProps {
  chatId: number;
  chatTitle: string;
  ownUserId: number;
  onClose: () => void;
}

export function ChatWindow(props: ChatWindowProps) {
  const { groupedMessages, isLoading, isSending, openChat, sendMessage, isOwnMessage } =
    useChatWindow();

  let scrollRef!: HTMLDivElement;

  onMount(() => {
    openChat(props.chatId, props.chatTitle, props.ownUserId);
  });

  createEffect(() => {
    groupedMessages();
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
          <For each={groupedMessages()}>
            {(group) => (
              <>
                <div class="chat-window__date-separator">
                  <span>{group.date}</span>
                </div>
                <For each={group.messages}>
                  {(msg) => (
                    <MessageItem
                      message={msg}
                      isOwn={isOwnMessage(msg.senderId)}
                    />
                  )}
                </For>
              </>
            )}
          </For>
          <Show when={groupedMessages().length === 0}>
            <div class="chat-window__empty">Нет сообщений</div>
          </Show>
        </Show>
      </div>

      <MessageInput onSend={sendMessage} isSending={isSending()} />
    </div>
  );
}
