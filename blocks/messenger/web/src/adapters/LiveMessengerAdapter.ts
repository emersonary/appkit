import { MessengerService } from '../gen/v1/messenger_connect';
import { createMessengerClient } from '../connect/transport';
import { toChannel, toMessage } from '../proto-map';
import type { Channel, Message, MessagePage, MessengerEvent, MessengerSessionStorage, SendMessageInput } from '../types';
import type { MessengerAdapter } from './MessengerAdapter';

export type LiveMessengerAdapterOptions = {
  tenantId: string;
  baseUrl?: string;
  storage: MessengerSessionStorage;
  deviceKey?: string;
};

export class LiveMessengerAdapter implements MessengerAdapter {
  readonly mode = 'live' as const;
  private readonly tenantId: string;
  private readonly storage: MessengerSessionStorage;
  private readonly client: ReturnType<typeof createMessengerClient<typeof MessengerService>>;
  private readonly deviceKey: string;
  private ws: WebSocket | null = null;
  private handlers = new Set<(event: MessengerEvent) => void>();
  private deviceId: string | null = null;

  constructor(options: LiveMessengerAdapterOptions) {
    this.tenantId = options.tenantId;
    this.storage = options.storage;
    this.deviceKey = options.deviceKey ?? `web-${crypto.randomUUID()}`;
    this.client = createMessengerClient(MessengerService, {
      baseUrl: options.baseUrl ?? '',
      storage: options.storage,
    });
  }

  async listChannels(): Promise<Channel[]> {
    const response = await this.client.listChannels({ tenantId: this.tenantId });
    return (response.channels ?? []).map(toChannel);
  }

  async getMessages(channelId: string, cursor?: string): Promise<MessagePage> {
    const response = await this.client.listMessages({
      tenantId: this.tenantId,
      channelId,
      beforeMessageId: cursor ?? '',
      limit: 50,
    });
    return { messages: (response.messages ?? []).map(toMessage) };
  }

  async sendMessage(input: SendMessageInput): Promise<Message> {
    const response = await this.client.sendMessage({
      tenantId: this.tenantId,
      channelId: input.channelId,
      body: input.body,
      clientRef: input.clientRef ?? '',
      replyToId: input.replyToId ?? '',
      deviceId: input.deviceId ?? this.deviceId ?? '',
    });
    return toMessage(response);
  }

  async markRead(channelId: string, messageId: string): Promise<void> {
    await this.client.markRead({ tenantId: this.tenantId, channelId, messageId });
  }

  subscribeEvents(handler: (event: MessengerEvent) => void): () => void {
    this.handlers.add(handler);
    void this.ensureWebSocket();
    return () => {
      this.handlers.delete(handler);
    };
  }

  private async ensureWebSocket() {
    if (this.ws) return;

    const reg = await this.client.registerDevice({
      tenantId: this.tenantId,
      deviceKey: this.deviceKey,
      platform: 'web',
      pushToken: '',
    });
    this.deviceId = reg.deviceId;

    const params = new URLSearchParams({
      tenant_id: this.tenantId,
      device_id: this.deviceKey,
    });
    const token = this.storage.readAccessToken();
    if (token) {
      params.set('token', token);
    }

    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const wsUrl = `${protocol}://${window.location.host}/messenger/ws?${params}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (ev) => {
      try {
        const data = JSON.parse(String(ev.data)) as { type?: string; payload?: Record<string, unknown> };
        if (data.type === 'message.created' && data.payload) {
          const raw = data.payload as { message?: Parameters<typeof toMessage>[0]; channel_id?: string };
          const msg = toMessage(raw.message ?? {});
          dispatch(this.handlers, {
            type: 'message.created',
            payload: { channelId: raw.channel_id ?? msg.channelId, message: msg },
          });
        } else if (data.type === 'read.updated' && data.payload) {
          const raw = data.payload as { channel_id?: string; message_id?: string };
          dispatch(this.handlers, {
            type: 'read.updated',
            payload: { channelId: raw.channel_id ?? '', messageId: raw.message_id ?? '' },
          });
        } else if (data.type === 'member.updated' && data.payload) {
          const raw = data.payload as { channel_id?: string };
          dispatch(this.handlers, {
            type: 'member.updated',
            payload: { channelId: raw.channel_id ?? '' },
          });
        }
      } catch {
        // ignore malformed events
      }
    };
  }
}

function dispatch(handlers: Set<(event: MessengerEvent) => void>, event: MessengerEvent) {
  for (const h of handlers) h(event);
}
