export type MessengerUserType = 'human' | 'agent' | 'system';

export type MessengerUser = {
  id: string;
  tenantId: string;
  userType: MessengerUserType;
  displayName: string;
  avatarUrl?: string;
  accountId?: string;
  externalRef?: string;
};

export type ChannelKind = 'direct' | 'group';

export type Channel = {
  id: string;
  tenantId: string;
  kind: ChannelKind;
  title: string;
  avatarUrl?: string;
  createdByUserId: string;
  unreadCount: number;
  lastMessagePreview?: string;
  updatedAt?: Date;
  members?: MessengerUser[];
};

export type MessageKind = 'user' | 'system';

export type Message = {
  id: string;
  channelId: string;
  kind: MessageKind;
  sender?: MessengerUser;
  body: string;
  eventType?: string;
  eventPayload?: Record<string, unknown>;
  clientRef?: string;
  replyToId?: string;
  createdAt: Date;
};

export type MessagePage = {
  messages: Message[];
};

export type SendMessageInput = {
  channelId: string;
  body: string;
  clientRef?: string;
  replyToId?: string;
  deviceId?: string;
};

export type MessengerEvent =
  | { type: 'message.created'; payload: { channelId: string; message: Message } }
  | { type: 'read.updated'; payload: { channelId: string; messageId: string } }
  | { type: 'member.updated'; payload: { channelId: string } };

export type MessengerAdapterMode = 'live' | 'simulation' | 'readonly';

export type MessengerSessionStorage = {
  readAccessToken: () => string | null;
};

export type MessengerLabels = {
  title?: string;
  channelsEmpty?: string;
  selectChannel?: string;
  composerPlaceholder?: string;
  send?: string;
  humanBadge?: string;
  agentBadge?: string;
  systemBadge?: string;
};
