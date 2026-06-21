export interface User {
  id: string;
  telegramId: number;
  firstName: string;
  lastName: string;
  username: string;
  phone: string;
  isBot: boolean;
}

export interface Chat {
  id: number;
  title: string;
  type: "direct" | "group" | "channel";
  unreadCount: number;
  lastMessage: Message;
}

export interface Message {
  id: number;
  chatId: number;
  senderId: number;
  text: string;
  createdAt: string;
  hasMedia: boolean;
  mediaId: string;
}

export interface Folder {
  id: number;
  title: string;
  chatIds: number[];
}

export type SessionStatus =
  | "AWAITING_PHONE"
  | "AWAITING_CODE"
  | "AWAITING_PASSWORD"
  | "AUTHORIZED";

export type NextStep = "AWAITING_PASSWORD" | "AUTHORIZED";

export interface AuthState {
  sessionToken: string | null;
  status: SessionStatus;
  isLoading: boolean;
  error: string | null;
}

export interface DashboardData {
  chats: Chat[];
  folders: Folder[];
  ownUserId: number;
}
