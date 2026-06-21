import { Show, For, onMount } from "solid-js";
import { useChatList } from "./store";
import { SearchBar } from "./components/SearchBar";
import { ChatCard } from "./components/ChatCard";
import { FolderTabs } from "./components/FolderTabs";
import { Loader } from "../../shared/components";
import "./ChatList.css";

interface ChatListProps {
  activeChatId: number | null;
  onSelectChat: (id: number) => void;
}

export function ChatList(props: ChatListProps) {
  const {
    filteredChats,
    folders,
    selectedFolderId,
    setSelectedFolderId,
    searchQuery,
    setSearchQuery,
    isLoading,
    loadDashboard,
  } = useChatList();

  onMount(() => {
    loadDashboard();
  });

  return (
    <div class="chat-list">
      <div class="chat-list__header">
        <h2 class="chat-list__title">Чаты</h2>
      </div>

      <SearchBar
        value={searchQuery()}
        onInput={setSearchQuery}
      />

      <FolderTabs
        folders={folders()}
        selectedId={selectedFolderId()}
        onSelect={setSelectedFolderId}
      />

      <Show when={!isLoading()} fallback={<div class="chat-list__loader"><Loader /></div>}>
        <div class="chat-list__items">
          <For each={filteredChats()}>
            {(chat) => (
              <ChatCard
                chat={chat}
                isActive={props.activeChatId === chat.id}
                onClick={() => props.onSelectChat(chat.id)}
              />
            )}
          </For>
          <Show when={filteredChats().length === 0}>
            <div class="chat-list__empty">Чатов не найдено</div>
          </Show>
        </div>
      </Show>
    </div>
  );
}
