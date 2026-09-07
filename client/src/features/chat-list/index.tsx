import { For, Show, onMount } from "solid-js";
import { Icon, Loader } from "~/shared/components";
import type { Chat } from "~/types";
import { ChatCard } from "./components/ChatCard";
import { FolderTabs } from "./components/FolderTabs";
import { SearchBar } from "./components/SearchBar";
import { useChatList } from "./store";
import "./ChatList.css";

export { useChatList } from "./store";
export type { Tab } from "./store";

interface ChatListProps {
  activeChatRef: string | null;
  onSelectChat: (chat: Chat) => void;
  onLogout: () => void;
}

export function ChatList(props: ChatListProps) {
  const {
    state,
    isReady,
    truncated,
    folders,
    archivedCount,
    visibleChats,
    tab,
    setTab,
    query,
    setQuery,
    load,
  } = useChatList();

  onMount(load);

  return (
    <div class="chat-list">
      <div class="chat-list__header">
        <div class="chat-list__brand">
          <h2 class="chat-list__title">telegleb</h2>
          <span class="chat-list__version mono">V1</span>
        </div>
        <div class="chat-list__actions">
          <button
            class="chat-list__action chat-list__action--danger"
            onClick={props.onLogout}
            title="Выйти"
          >
            <Icon name="logout" size={17} />
          </button>
        </div>
      </div>

      <SearchBar value={query()} onInput={setQuery} />

      {/*
        Вкладки и список появляются одновременно и только после того, как
        сервер прислал раскладку. Пока её нет, показывается загрузка — чат без
        известной папки не рисуется ни на кадр.
      */}
      <Show
        when={isReady()}
        fallback={
          <div class="chat-list__status">
            <Show
              when={state().status === "error"}
              fallback={<Loader />}
            >
              <div class="chat-list__error">
                <p>{(state() as { message: string }).message}</p>
                <button class="chat-list__retry" onClick={load}>
                  Повторить
                </button>
              </div>
            </Show>
          </div>
        }
      >
        <Show when={truncated()}>
          <div class="chat-list__notice">
            Показаны не все диалоги — папки могут быть неполными. Подробности в консоли.
          </div>
        </Show>

        <FolderTabs
          folders={folders()}
          archivedCount={archivedCount()}
          selected={tab()}
          onSelect={setTab}
        />

        <div class="chat-list__items">
          <For each={visibleChats()}>
            {(chat) => (
              <ChatCard
                chat={chat}
                isActive={props.activeChatRef === chat.ref}
                onClick={() => props.onSelectChat(chat)}
              />
            )}
          </For>
          <Show when={visibleChats().length === 0}>
            <div class="chat-list__empty">
              {tab().kind === "archive" ? "Архив пуст" : "Чатов не найдено"}
            </div>
          </Show>
        </div>
      </Show>
    </div>
  );
}
