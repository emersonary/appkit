import type {
  Channel,
  Message,
  MessagePage,
  MessengerAdapterMode,
  MessengerEvent,
  SendMessageInput,
} from '../types';

export interface MessengerAdapter {
  mode: MessengerAdapterMode;
  listChannels(): Promise<Channel[]>;
  getMessages(channelId: string, cursor?: string): Promise<MessagePage>;
  sendMessage(input: SendMessageInput): Promise<Message>;
  markRead(channelId: string, messageId: string): Promise<void>;
  subscribeEvents(handler: (event: MessengerEvent) => void): () => void;
}
