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

export interface MessageEntity {
  offset: number;
  length: number;
  type: string;
  url?: string;
  userId?: number;
}

export interface Message {
  id: number;
  chatId: number;
  senderId: number;
  text: string;
  createdAt: string;
  hasMedia: boolean;
  mediaId: string;

  out: boolean;
  mentioned: boolean;
  silent: boolean;
  post: boolean;
  pinned: boolean;
  noforwards: boolean;
  editDate: string;
  views: number;
  forwards: number;
  groupedId: number;
  viaBotId: number;
  postAuthor: string;
  ttlPeriod: number;

  replyToMsgId: number;
  replyToPeer: number;

  fwdFromName: string;
  fwdFromDate: string;
  fwdFromChannelId: number;
  fwdFromUserId: number;

  repliesCount: number;
  repliesMaxId: number;

  entities: MessageEntity[];
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
