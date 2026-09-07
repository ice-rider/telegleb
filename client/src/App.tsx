import { Match, Show, Switch, onMount } from "solid-js";
import { AuthFeature, useAuth } from "./features/auth";
import { ChatList } from "./features/chat-list";
import { ChatWindow, useChatWindow } from "./features/chat-window";
import { Loader } from "./shared/components";
import "./App.css";

export default function App() {
  const { status, bootstrap, logout } = useAuth();
  const { activeChat, openChat, closeChat } = useChatWindow();

  onMount(bootstrap);

  async function handleLogout() {
    closeChat();
    await logout();
  }

  return (
    <div class="app">
      <Switch>
        <Match when={status() === "checking"}>
          <div class="app__boot">
            <Loader />
          </div>
        </Match>

        <Match when={status() === "authorized"}>
          <div class="app__layout">
            <div class="app__sidebar">
              <ChatList
                activeChatRef={activeChat()?.ref ?? null}
                onSelectChat={openChat}
                onLogout={handleLogout}
              />
            </div>
            <div class="app__main">
              {/*
                keyed: при переходе между чатами компонент пересоздаётся.
                Без этого <Show> сохраняет прежнего потомка, и в новом чате
                остаются сообщения предыдущего.
              */}
              <Show
                when={activeChat()}
                keyed
                fallback={
                  <div class="app__placeholder">
                    <div class="app__placeholder-icon">💬</div>
                    <p>Выберите чат</p>
                  </div>
                }
              >
                {(chat) => <ChatWindow chat={chat} onClose={closeChat} />}
              </Show>
            </div>
          </div>
        </Match>

        <Match when={true}>
          <div class="app__auth">
            <AuthFeature />
          </div>
        </Match>
      </Switch>
    </div>
  );
}
