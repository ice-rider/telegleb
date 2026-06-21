export interface MessageRecord {
  id: number;
  chatId: number;
  senderId: number;
  text: string;
  createdAt: string;
  hasMedia: boolean;
  mediaId: string;
}

export interface ChatRecord {
  id: number;
  title: string;
  type: "direct" | "group" | "channel";
  unreadCount: number;
  lastMessageText: string;
  lastMessageTime: string;
}

const DB_NAME = "telegleb";
const DB_VERSION = 1;

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains("messages")) {
        const store = db.createObjectStore("messages", { keyPath: "id" });
        store.createIndex("chatId", "chatId", { unique: false });
      }
      if (!db.objectStoreNames.contains("chats")) {
        db.createObjectStore("chats", { keyPath: "id" });
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

export async function saveMessages(messages: MessageRecord[]) {
  const db = await openDB();
  const tx = db.transaction("messages", "readwrite");
  const store = tx.objectStore("messages");
  for (const msg of messages) {
    store.put(msg);
  }
  return new Promise<void>((resolve, reject) => {
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function getMessagesByChat(
  chatId: number,
): Promise<MessageRecord[]> {
  const db = await openDB();
  const tx = db.transaction("messages", "readonly");
  const index = tx.objectStore("messages").index("chatId");
  return new Promise((resolve, reject) => {
    const req = index.getAll(chatId);
    req.onsuccess = () =>
      resolve(
        req.result.sort(
          (a, b) =>
            new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
        ),
      );
    req.onerror = () => reject(req.error);
  });
}

export async function saveChats(chats: ChatRecord[]) {
  const db = await openDB();
  const tx = db.transaction("chats", "readwrite");
  const store = tx.objectStore("chats");
  for (const chat of chats) {
    store.put(chat);
  }
  return new Promise<void>((resolve, reject) => {
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function getChats(): Promise<ChatRecord[]> {
  const db = await openDB();
  const tx = db.transaction("chats", "readonly");
  const store = tx.objectStore("chats");
  return new Promise((resolve, reject) => {
    const req = store.getAll();
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}
