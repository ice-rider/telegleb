import { createSignal, createMemo } from "solid-js";
import { api } from "../../core/api";
import type { Chat, Folder } from "../../types";

const [chats, setChats] = createSignal<Chat[]>([]);
const [folders, setFolders] = createSignal<Folder[]>([]);
const [ownUserId, setOwnUserId] = createSignal<number>(0);
const [selectedFolderId, setSelectedFolderId] = createSignal<number | null>(null);
const [isLoading, setIsLoading] = createSignal(false);
const [error, setError] = createSignal<string | null>(null);
const [searchQuery, setSearchQuery] = createSignal("");

export const useChatList = () => {
  async function loadDashboard() {
    setIsLoading(true);
    setError(null);
    try {
      const data = await api.loadDashboard(100, 0);
      setChats(data.chats);
      setFolders(data.folders);
      setOwnUserId(data.ownUserId);
    } catch (e: any) {
      setError(e.message || "Ошибка загрузки");
    } finally {
      setIsLoading(false);
    }
  }

  const filteredChats = createMemo(() => {
    const q = searchQuery().toLowerCase();
    const folderId = selectedFolderId();
    let list = chats();

    if (folderId !== null) {
      const folder = folders().find((f) => f.id === folderId);
      if (folder) {
        list = list.filter((c) => folder.chatIds.includes(c.id));
      }
    }

    if (q) {
      list = list.filter((c) => c.title.toLowerCase().includes(q));
    }

    return list.sort((a, b) => {
      const aTime = a.lastMessage.createdAt ? new Date(a.lastMessage.createdAt).getTime() : 0;
      const bTime = b.lastMessage.createdAt ? new Date(b.lastMessage.createdAt).getTime() : 0;
      return bTime - aTime;
    });
  });

  return {
    chats,
    folders,
    ownUserId,
    filteredChats,
    selectedFolderId,
    setSelectedFolderId,
    searchQuery,
    setSearchQuery,
    isLoading,
    error,
    loadDashboard,
  };
};
