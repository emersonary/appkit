import type { Channel, ChannelKind, Message, MessageKind, MessengerUser, MessengerUserType } from './types';

export function mapUserType(value: number | string | undefined): MessengerUserType {
  if (value === 1 || value === 'USER_TYPE_HUMAN') return 'human';
  if (value === 2 || value === 'USER_TYPE_AGENT') return 'agent';
  if (value === 3 || value === 'USER_TYPE_SYSTEM') return 'system';
  return 'human';
}

export function mapChannelKind(value: number | string | undefined): ChannelKind {
  if (value === 2 || value === 'CHANNEL_KIND_GROUP') return 'group';
  return 'direct';
}

export function mapMessageKind(value: number | string | undefined): MessageKind {
  if (value === 2 || value === 'MESSAGE_KIND_SYSTEM') return 'system';
  return 'user';
}

export function toMessengerUser(u: {
  id?: string;
  tenantId?: string;
  userType?: number;
  displayName?: string;
  avatarUrl?: string;
  accountId?: string;
  externalRef?: string;
} | undefined): MessengerUser | undefined {
  if (!u?.id) return undefined;
  return {
    id: u.id,
    tenantId: u.tenantId ?? '',
    userType: mapUserType(u.userType),
    displayName: u.displayName ?? 'User',
    avatarUrl: u.avatarUrl || undefined,
    accountId: u.accountId || undefined,
    externalRef: u.externalRef || undefined,
  };
}

export function toChannel(c: {
  id?: string;
  tenantId?: string;
  kind?: number;
  title?: string;
  avatarUrl?: string;
  createdByUserId?: string;
  unreadCount?: number;
  lastMessagePreview?: string;
  updatedAt?: { seconds?: bigint; nanos?: number };
  members?: Array<Parameters<typeof toMessengerUser>[0]>;
}): Channel {
  return {
    id: c.id ?? '',
    tenantId: c.tenantId ?? '',
    kind: mapChannelKind(c.kind),
    title: c.title ?? '',
    avatarUrl: c.avatarUrl || undefined,
    createdByUserId: c.createdByUserId ?? '',
    unreadCount: c.unreadCount ?? 0,
    lastMessagePreview: c.lastMessagePreview || undefined,
    updatedAt: c.updatedAt?.seconds ? new Date(Number(c.updatedAt.seconds) * 1000) : undefined,
    members: (c.members ?? []).map((m) => toMessengerUser(m)).filter(Boolean) as MessengerUser[],
  };
}

export function toMessage(m: {
  id?: string;
  channelId?: string;
  kind?: number;
  sender?: Parameters<typeof toMessengerUser>[0];
  body?: string;
  eventType?: string;
  eventPayload?: { fields?: Record<string, unknown> };
  clientRef?: string;
  replyToId?: string;
  createdAt?: { seconds?: bigint; nanos?: number };
}): Message {
  let eventPayload: Record<string, unknown> | undefined;
  if (m.eventPayload && typeof m.eventPayload === 'object' && 'fields' in m.eventPayload) {
    eventPayload = m.eventPayload as Record<string, unknown>;
  } else if (m.eventPayload) {
    eventPayload = m.eventPayload as unknown as Record<string, unknown>;
  }

  return {
    id: m.id ?? '',
    channelId: m.channelId ?? '',
    kind: mapMessageKind(m.kind),
    sender: toMessengerUser(m.sender),
    body: m.body ?? '',
    eventType: m.eventType || undefined,
    eventPayload,
    clientRef: m.clientRef || undefined,
    replyToId: m.replyToId || undefined,
    createdAt: m.createdAt?.seconds ? new Date(Number(m.createdAt.seconds) * 1000) : new Date(),
  };
}
