import type { Channel, Message, MessagePage, MessengerEvent, MessengerUser, SendMessageInput } from '../types';
import type { MessengerAdapter } from './MessengerAdapter';

export type SimulationMessengerAdapterOptions = {
  allowSend?: boolean;
  initialChannels?: Channel[];
  initialMessages?: Record<string, Message[]>;
};

const defaultHuman: MessengerUser = {
  id: 'sim-human-1',
  tenantId: 'sim-tenant',
  userType: 'human',
  displayName: 'Alex Customer',
};

const defaultAgent: MessengerUser = {
  id: 'sim-agent-1',
  tenantId: 'sim-tenant',
  userType: 'agent',
  displayName: 'Support Bot',
  externalRef: 'solidia-support-bot',
};

const defaultChannel: Channel = {
  id: 'sim-channel-1',
  tenantId: 'sim-tenant',
  kind: 'direct',
  title: 'Support',
  createdByUserId: defaultHuman.id,
  unreadCount: 1,
  lastMessagePreview: 'How can I help you today?',
  members: [defaultHuman, defaultAgent],
};

const defaultMessages: Message[] = [
  {
    id: 'sim-msg-1',
    channelId: defaultChannel.id,
    kind: 'user',
    sender: defaultHuman,
    body: 'Hi, I need help planning a trip.',
    createdAt: new Date(Date.now() - 120_000),
  },
  {
    id: 'sim-msg-2',
    channelId: defaultChannel.id,
    kind: 'user',
    sender: defaultAgent,
    body: 'How can I help you today?',
    createdAt: new Date(Date.now() - 60_000),
  },
];

export class SimulationMessengerAdapter implements MessengerAdapter {
  readonly mode = 'simulation' as const;
  private channels: Channel[];
  private messagesByChannel: Map<string, Message[]>;
  private handlers = new Set<(event: MessengerEvent) => void>();
  private allowSend: boolean;
  private seq = 100;

  constructor(options: SimulationMessengerAdapterOptions = {}) {
    this.allowSend = options.allowSend ?? true;
    this.channels = options.initialChannels ?? [defaultChannel];
    this.messagesByChannel = new Map(
      Object.entries(options.initialMessages ?? { [defaultChannel.id]: defaultMessages }),
    );
  }

  async listChannels(): Promise<Channel[]> {
    return this.channels.map((ch) => ({
      ...ch,
      lastMessagePreview: this.lastPreview(ch.id),
      unreadCount: this.messagesByChannel.get(ch.id)?.length ?? 0,
    }));
  }

  async getMessages(channelId: string): Promise<MessagePage> {
    return { messages: [...(this.messagesByChannel.get(channelId) ?? [])] };
  }

  async sendMessage(input: SendMessageInput): Promise<Message> {
    if (!this.allowSend) {
      throw new Error('simulation adapter: sending disabled');
    }
    const msg: Message = {
      id: `sim-msg-${++this.seq}`,
      channelId: input.channelId,
      kind: 'user',
      sender: defaultHuman,
      body: input.body,
      clientRef: input.clientRef,
      createdAt: new Date(),
    };
    const list = this.messagesByChannel.get(input.channelId) ?? [];
    list.push(msg);
    this.messagesByChannel.set(input.channelId, list);
    dispatch(this.handlers, { type: 'message.created', payload: { channelId: input.channelId, message: msg } });
    return msg;
  }

  async markRead(channelId: string, messageId: string): Promise<void> {
    dispatch(this.handlers, { type: 'read.updated', payload: { channelId, messageId } });
  }

  subscribeEvents(handler: (event: MessengerEvent) => void): () => void {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }

  simulateHandoff(channelId: string) {
    const msg: Message = {
      id: `sim-msg-${++this.seq}`,
      channelId,
      kind: 'system',
      body: 'Conversation transferred to a human agent',
      eventType: 'handoff',
      eventPayload: { from_user_id: defaultAgent.id, to_user_id: defaultHuman.id },
      createdAt: new Date(),
    };
    const list = this.messagesByChannel.get(channelId) ?? [];
    list.push(msg);
    this.messagesByChannel.set(channelId, list);
    dispatch(this.handlers, { type: 'message.created', payload: { channelId, message: msg } });
    dispatch(this.handlers, { type: 'member.updated', payload: { channelId } });
  }

  private lastPreview(channelId: string): string | undefined {
    const msgs = this.messagesByChannel.get(channelId);
    if (!msgs?.length) return undefined;
    return msgs[msgs.length - 1]?.body;
  }
}

function dispatch(handlers: Set<(event: MessengerEvent) => void>, event: MessengerEvent) {
  for (const h of handlers) h(event);
}
