import { Show, createSignal, onMount } from "solid-js";
import { useAuth } from "./features/auth";
import { AuthFeature } from "./features/auth";
import { ChatList, useChatList } from "./features/chat-list";
import { ChatWindow } from "./features/chat-window";
import "./App.css";

export default function App() {
  const { isLoggedIn, initFromStorage } = useAuth();
  const { ownUserId } = useChatList();
  const [activeChatId, setActiveChatId] = createSignal<number | null>(null);
  const [activeChatTitle, setActiveChatTitle] = createSignal("");

  onMount(() => {
    initFromStorage();
  });

  function handleSelectChat(id: number, title: string) {
    setActiveChatId(id);
    setActiveChatTitle(title);
  }

  function handleCloseChat() {
    setActiveChatId(null);
    setActiveChatTitle("");
  }

  return (
    <div class="app">
      <Show
        when={isLoggedIn()}
        fallback={
          <div class="app__auth">
            <AuthFeature />
          </div>
        }
      >
        <div class="app__layout">
          <div class="app__sidebar">
            <ChatList
              activeChatId={activeChatId()}
              onSelectChat={handleSelectChat}
            />
          </div>
          <div class="app__main">
            <Show
              when={activeChatId() !== null}
              fallback={
                <div class="app__placeholder">
                  <div class="app__placeholder-icon">💬</div>
                  <p>Выберите чат</p>
                </div>
              }
            >
              <ChatWindow
                chatId={activeChatId()!}
                chatTitle={activeChatTitle()}
                ownUserId={ownUserId()}
                onClose={handleCloseChat}
              />
            </Show>
          </div>
        </div>
      </Show>
    </div>
  );
}
