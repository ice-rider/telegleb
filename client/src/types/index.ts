// Формы данных сервера. Имена полей совпадают с JSON один в один — контракт
// использует lowerCamelCase на всех уровнях, поэтому слоя переименования
// между сетью и приложением больше нет.

export type ChatType = "direct" | "group" | "channel";

export type MediaKind =
  | "photo"
  | "video"
  | "audio"
  | "voice"
  | "document"
  | "sticker"
  | "gif";

export interface Me {
  id: number;
  firstName: string;
  lastName: string;
  username: string;
  phone: string;
}

export interface MessageEntity {
  offset: number;
  length: number;
  type: string;
  url?: string;
  userId?: number;
  language?: string;
}

export interface Media {
  /** Непрозрачная ссылка на вложение. Клиент её не разбирает. */
  ref: string;
  kind: MediaKind;
  mimeType?: string;
  size?: number;
  width?: number;
  height?: number;
  fileName?: string;
}

export interface ReplyPreview {
  id: number;
  senderName?: string;
  text?: string;
}

export interface Message {
  id: number;
  chatRef: string;
  senderId: number;
  senderName?: string;
  /** Единственный признак «моё сообщение». Сравнивать senderId с id профиля нельзя. */
  out: boolean;
  text: string;
  createdAt: string;
  editedAt?: string;
  entities: MessageEntity[];
  media?: Media;
  replyTo?: ReplyPreview;
  forwardedFrom?: string;
  views?: number;
  pinned?: boolean;
}

export interface Chat {
  /** Адрес чата. Все запросы идут по нему, а не по числовому id. */
  ref: string;
  id: number;
  type: ChatType;
  title: string;
  unreadCount: number;
  unreadMentionsCount: number;
  markedUnread: boolean;
  pinned: boolean;
  muted: boolean;
  /** Размещение приезжает вместе с чатом, а не отдельным справочником. */
  archived: boolean;
  folderIds: number[];
  order: number;
  /** Супергруппа с темами: плоской истории нет, сначала показывается список тем. */
  isForum: boolean;
  lastMessage?: Message;
}

/** Тема внутри форума-супергруппы — по сути отдельный чат внутри чата. */
export interface Topic {
  id: number;
  title: string;
  iconColor?: number;
  /** id кастомного эмодзи строкой: в double int64 не влезает. */
  iconEmojiId?: string;
  unreadCount: number;
  unreadMentionsCount: number;
  pinned: boolean;
  closed: boolean;
  hidden: boolean;
  order: number;
  lastMessage?: Message;
}

export interface Folder {
  id: number;
  title: string;
  emoticon?: string;
  order: number;
}

export interface DashboardStats {
  pages: number;
  rawDialogs: number;
  skippedUnknownPeer: number;
  skippedNotDialog: number;
}

export interface Dashboard {
  chats: Chat[];
  folders: Folder[];
  me: Me;
  /** Выборка диалогов упёрлась в потолок — раскладка по папкам может быть неполной. */
  truncated?: boolean;
  stats: DashboardStats;
}

export type NextStep = "code" | "password" | "done";

export type AuthStatus =
  | "checking"
  | "awaitingPhone"
  | "awaitingCode"
  | "awaitingPassword"
  | "authorized";
