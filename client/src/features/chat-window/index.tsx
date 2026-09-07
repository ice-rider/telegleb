import { For, Show, createEffect } from "solid-js";
import { Loader } from "~/shared/components";
import type { Chat } from "~/types";
import { MessageInput } from "./components/MessageInput";
import { MessageItem } from "./components/MessageItem";
import { TopicList } from "./components/TopicList";
import { useChatWindow } from "./store";
import "./ChatWindow.css";

export { useChatWindow } from "./store";

interface ChatWindowProps {
  chat: Chat;
  onClose: () => void;
}

export function ChatWindow(props: ChatWindowProps) {
  const {
    activeTopic,
    topics,
    showingTopicList,
    groupedMessages,
    isLoading,
    isSending,
    error,
    openTopic,
    closeTopic,
    sendMessage,
  } = useChatWindow();

  let scrollRef!: HTMLDivElement;

  createEffect(() => {
    groupedMessages();
    if (scrollRef) scrollRef.scrollTop = scrollRef.scrollHeight;
  });

  // Имя отправителя нужно только там, где собеседников больше одного.
  const showSender = () => props.chat.type !== "direct";

  // Из темы возвращаемся к списку тем, из обычного чата — закрываем чат.
  const goBack = () => (activeTopic() ? closeTopic() : props.onClose());

  return (
    <div class="chat-window">
      <div class="chat-window__header">
        <button class="chat-window__back" onClick={goBack}>
          ←
        </button>
        <div class="chat-window__heading">
          <h3 class="chat-window__title">{activeTopic()?.title ?? props.chat.title}</h3>
          <Show when={activeTopic()}>
            <span class="chat-window__subtitle">{props.chat.title}</span>
          </Show>
          <Show when={showingTopicList()}>
            <span class="chat-window__subtitle">тем: {topics().length}</span>
          </Show>
        </div>
      </div>

      <Show when={!isLoading()} fallback={<div class="chat-window__loader"><Loader /></div>}>
        <Show when={error()}>
          {(message) => <div class="chat-window__error">{message()}</div>}
        </Show>

        {/* У форума нет плоской истории: сначала темы, потом сообщения темы. */}
        <Show
          when={!showingTopicList()}
          fallback={<TopicList topics={topics()} onSelect={openTopic} />}
        >
          <div class="chat-window__messages" ref={scrollRef}>
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
          </div>

          <MessageInput onSend={sendMessage} isSending={isSending()} />
        </Show>
      </Show>
    </div>
  );
}
