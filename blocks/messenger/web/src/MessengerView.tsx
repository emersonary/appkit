import { useCallback, useEffect, useMemo, useState } from 'react';
import type { MessengerAdapter } from './adapters/MessengerAdapter';
import type { Channel, Message, MessengerLabels, MessengerUser } from './types';

export type MessengerViewProps = {
  adapter: MessengerAdapter;
  currentUserId?: string;
  labels?: MessengerLabels;
  renderSystemMessage?: (message: Message) => React.ReactNode;
};

const defaultLabels: Required<MessengerLabels> = {
  title: 'Messenger',
  channelsEmpty: 'No conversations yet',
  selectChannel: 'Select a conversation',
  composerPlaceholder: 'Type a message',
  send: 'Send',
  humanBadge: 'Human',
  agentBadge: 'AI Agent',
  systemBadge: 'System',
};

function userTypeBadge(user: MessengerUser | undefined, labels: Required<MessengerLabels>) {
  if (!user) return null;
  if (user.userType === 'agent') return labels.agentBadge;
  if (user.userType === 'system') return labels.systemBadge;
  return labels.humanBadge;
}

function channelTitle(channel: Channel) {
  if (channel.title) return channel.title;
  const names = (channel.members ?? []).map((m) => m.displayName).filter(Boolean);
  return names.join(', ') || 'Conversation';
}

export function MessengerView({ adapter, currentUserId, labels: labelOverrides, renderSystemMessage }: MessengerViewProps) {
  const labels = useMemo(() => ({ ...defaultLabels, ...labelOverrides }), [labelOverrides]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [activeChannelId, setActiveChannelId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [draft, setDraft] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);

  const activeChannel = channels.find((c) => c.id === activeChannelId) ?? null;
  const composerDisabled = adapter.mode === 'readonly';

  const refreshChannels = useCallback(async () => {
    const list = await adapter.listChannels();
    setChannels(list);
    if (!activeChannelId && list[0]) {
      setActiveChannelId(list[0].id);
    }
  }, [adapter, activeChannelId]);

  const loadMessages = useCallback(
    async (channelId: string) => {
      const page = await adapter.getMessages(channelId);
      setMessages(page.messages);
      const last = page.messages[page.messages.length - 1];
      if (last) {
        await adapter.markRead(channelId, last.id);
      }
    },
    [adapter],
  );

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        await refreshChannels();
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [refreshChannels]);

  useEffect(() => {
    if (!activeChannelId) return;
    void loadMessages(activeChannelId);
  }, [activeChannelId, loadMessages]);

  useEffect(() => {
    return adapter.subscribeEvents((event) => {
      if (event.type === 'message.created') {
        void refreshChannels();
        if (event.payload.channelId === activeChannelId) {
          setMessages((prev) => {
            if (prev.some((m) => m.id === event.payload.message.id)) return prev;
            return [...prev, event.payload.message];
          });
        }
      }
      if (event.type === 'member.updated') {
        void refreshChannels();
      }
    });
  }, [adapter, activeChannelId, refreshChannels]);

  async function handleSend() {
    if (!activeChannelId || !draft.trim() || adapter.mode === 'readonly') return;
    setSending(true);
    try {
      const msg = await adapter.sendMessage({ channelId: activeChannelId, body: draft.trim() });
      setMessages((prev) => [...prev, msg]);
      setDraft('');
      await refreshChannels();
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="appkit-messenger">
      <aside className="appkit-messenger__sidebar">
        <header className="appkit-messenger__sidebar-header">{labels.title}</header>
        {loading ? (
          <p className="appkit-messenger__empty">{labels.channelsEmpty}</p>
        ) : channels.length === 0 ? (
          <p className="appkit-messenger__empty">{labels.channelsEmpty}</p>
        ) : (
          <ul className="appkit-messenger__channel-list">
            {channels.map((ch) => (
              <li key={ch.id}>
                <button
                  type="button"
                  className={`appkit-messenger__channel${ch.id === activeChannelId ? ' appkit-messenger__channel--active' : ''}`}
                  onClick={() => setActiveChannelId(ch.id)}
                >
                  <span className="appkit-messenger__channel-title">{channelTitle(ch)}</span>
                  {ch.lastMessagePreview ? (
                    <span className="appkit-messenger__channel-preview">{ch.lastMessagePreview}</span>
                  ) : null}
                  {ch.unreadCount > 0 ? (
                    <span className="appkit-messenger__unread">{ch.unreadCount}</span>
                  ) : null}
                </button>
              </li>
            ))}
          </ul>
        )}
      </aside>

      <section className="appkit-messenger__thread">
        {!activeChannel ? (
          <p className="appkit-messenger__empty">{labels.selectChannel}</p>
        ) : (
          <>
            <header className="appkit-messenger__thread-header">
              <h2>{channelTitle(activeChannel)}</h2>
              <div className="appkit-messenger__badges">
                {(activeChannel.members ?? []).map((m) => (
                  <span key={m.id} className={`appkit-messenger__badge appkit-messenger__badge--${m.userType}`}>
                    {m.displayName} · {userTypeBadge(m, labels)}
                  </span>
                ))}
              </div>
            </header>

            <div className="appkit-messenger__messages">
              {messages.map((msg) => {
                if (msg.kind === 'system') {
                  return (
                    <div key={msg.id} className="appkit-messenger__system-row">
                      {renderSystemMessage ? renderSystemMessage(msg) : (
                        <span>{msg.body || msg.eventType}</span>
                      )}
                    </div>
                  );
                }
                const mine = currentUserId ? msg.sender?.id === currentUserId : false;
                return (
                  <div
                    key={msg.id}
                    className={`appkit-messenger__bubble-row${mine ? ' appkit-messenger__bubble-row--mine' : ''}`}
                  >
                    <div className="appkit-messenger__bubble">
                      {msg.sender ? (
                        <span className="appkit-messenger__sender">
                          {msg.sender.displayName}
                          <span className="appkit-messenger__sender-type">{userTypeBadge(msg.sender, labels)}</span>
                        </span>
                      ) : null}
                      <p>{msg.body}</p>
                    </div>
                  </div>
                );
              })}
            </div>

            {adapter.mode !== 'readonly' ? (
              <footer className="appkit-messenger__composer">
                <input
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  placeholder={labels.composerPlaceholder}
                  disabled={composerDisabled || sending}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault();
                      void handleSend();
                    }
                  }}
                />
                <button type="button" onClick={() => void handleSend()} disabled={composerDisabled || sending || !draft.trim()}>
                  {sending ? '…' : labels.send}
                </button>
              </footer>
            ) : null}
          </>
        )}
      </section>
    </div>
  );
}
