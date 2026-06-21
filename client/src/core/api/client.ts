import axios from "axios";

export interface RequestLoginResponse {
  sessionToken: string;
}

export interface VerifyCodeResponse {
  nextStep: "AWAITING_PASSWORD" | "AUTHORIZED";
}

export interface VerifyPasswordResponse {
  status: string;
}

export interface LoadDashboardResponse {
  chats: ChatDTO[];
  folders: FolderDTO[];
  OwnUserID: number;
}

export interface OpenChatResponse {
  messages: MessageDTO[];
}

export interface SendMessageResponse {
  message: MessageDTO;
}

export interface ChatDTO {
  ID: number;
  Title: string;
  Type: number;
  UnreadCount: number;
  LastMessage: MessageDTO;
}

export interface MessageDTO {
  ID: number;
  ChatID: number;
  SenderID: number;
  Text: string;
  CreatedAt: string;
  HasMedia: boolean;
  MediaId: string;
}

export interface FolderDTO {
  ID: number;
  Title: string;
  ChatIDs: number[];
}

function mapChatType(type: number): "direct" | "group" | "channel" {
  if (type === 0) return "direct";
  if (type === 1) return "group";
  return "channel";
}

function mapChat(dto: ChatDTO) {
  return {
    id: dto.ID,
    title: dto.Title,
    type: mapChatType(dto.Type),
    unreadCount: dto.UnreadCount,
    lastMessage: {
      id: dto.LastMessage.ID,
      chatId: dto.LastMessage.ChatID,
      senderId: dto.LastMessage.SenderID,
      text: dto.LastMessage.Text,
      createdAt: dto.LastMessage.CreatedAt,
      hasMedia: dto.LastMessage.HasMedia,
      mediaId: dto.LastMessage.MediaId,
    },
  };
}

function mapMessage(dto: MessageDTO) {
  return {
    id: dto.ID,
    chatId: dto.ChatID,
    senderId: dto.SenderID,
    text: dto.Text,
    createdAt: dto.CreatedAt,
    hasMedia: dto.HasMedia,
    mediaId: dto.MediaId,
  };
}

function mapFolder(dto: FolderDTO) {
  return {
    id: dto.ID,
    title: dto.Title,
    chatIds: dto.ChatIDs,
  };
}

const http = axios.create({
  baseURL: "/api/v1",
  headers: { "Content-Type": "application/json" },
});

let sessionToken: string | null = null;

http.interceptors.request.use((config) => {
  if (sessionToken) {
    config.headers.Authorization = `Bearer ${sessionToken}`;
  }
  return config;
});

http.interceptors.response.use(
  (res) => res,
  (err) => {
    const msg =
      err.response?.data?.error || err.message || "Request failed";
    return Promise.reject(new Error(msg));
  },
);

export const api = {
  setToken(token: string) {
    sessionToken = token;
  },

  clearToken() {
    sessionToken = null;
  },

  async requestLogin(phoneNumber: string) {
    const { data } = await http.post<RequestLoginResponse>(
      "/auth/request-login",
      { PhoneNumber: phoneNumber },
    );
    sessionToken = data.sessionToken;
    return data;
  },

  async verifyCode(stoken: string, code: string) {
    const { data } = await http.post<VerifyCodeResponse>(
      "/auth/verify-code",
      { SessionToken: stoken, Code: code },
    );
    return data;
  },

  async verifyPassword(stoken: string, password: string) {
    const { data } = await http.post<VerifyPasswordResponse>(
      "/auth/verify-password",
      { SessionToken: stoken, Password: password },
    );
    return data;
  },

  async logout(stoken: string) {
    await http.post("/auth/logout", { SessionToken: stoken });
    sessionToken = null;
  },

  async loadDashboard(limit = 30, offset = 0) {
    const { data } = await http.post<LoadDashboardResponse>(
      "/messenger/dashboard",
      { SessionToken: sessionToken, Limit: limit, Offset: offset },
    );
    return {
      chats: data.chats.map(mapChat),
      folders: data.folders.map(mapFolder),
      ownUserId: data.OwnUserID,
    };
  },

  async openChat(chatId: number, limit = 50, offset = 0) {
    const { data } = await http.post<OpenChatResponse>(
      "/messenger/open-chat",
      {
        SessionToken: sessionToken,
        ChatID: String(chatId),
        Limit: limit,
        Offset: offset,
      },
    );
    return data.messages.map(mapMessage);
  },

  async sendMessage(chatId: number, content: string) {
    const { data } = await http.post<SendMessageResponse>(
      "/messenger/send-message",
      {
        SessionToken: sessionToken,
        ChatID: String(chatId),
        Content: content,
      },
    );
    return mapMessage(data.message);
  },

  getMediaUrl(mediaId: string): string {
    return `/api/v1/media/stream/${mediaId}?token=${sessionToken}`;
  },
};
