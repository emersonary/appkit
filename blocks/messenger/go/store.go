package messenger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	platformuuid "github.com/emersonary/appkit/uuid"
)

const visibleRowStatus = "('A', 'I')"

type Store struct {
	db  *sql.DB
	cfg Config
}

func NewStore(db *sql.DB, cfg Config) *Store {
	cfg.normalize()
	return &Store{db: db, cfg: cfg}
}

func (s *Store) usersTable() string    { return qualifiedName(s.cfg.Schema, "tbl_users") }
func (s *Store) devicesTable() string  { return qualifiedName(s.cfg.Schema, "tbl_devices") }
func (s *Store) channelsTable() string { return qualifiedName(s.cfg.Schema, "tbl_channels") }
func (s *Store) directTable() string   { return qualifiedName(s.cfg.Schema, "tbl_direct_channels") }
func (s *Store) membersTable() string  { return qualifiedName(s.cfg.Schema, "tbl_channel_members") }
func (s *Store) messagesTable() string { return qualifiedName(s.cfg.Schema, "tbl_messages") }
func (s *Store) receiptsTable() string { return qualifiedName(s.cfg.Schema, "tbl_message_receipts") }
func (s *Store) handoffsTable() string { return qualifiedName(s.cfg.Schema, "tbl_channel_handoffs") }

func sortUserPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

func scanUser(row scannable) (User, error) {
	var u User
	var accountID, avatarURL, externalRef sql.NullString
	err := row.Scan(
		&u.ID, &u.TenantID, &u.UserType, &accountID, &u.DisplayName, &avatarURL, &externalRef, &u.CreatedAt, &u.UpdatedAt,
	)
	if accountID.Valid {
		u.AccountID = &accountID.String
	}
	if avatarURL.Valid {
		u.AvatarURL = &avatarURL.String
	}
	if externalRef.Valid {
		u.ExternalRef = &externalRef.String
	}
	return u, err
}

type scannable interface {
	Scan(dest ...any) error
}

func (s *Store) GetUserByID(ctx context.Context, tenantID, userID string) (User, error) {
	q := fmt.Sprintf(`
SELECT id, id_tenant, str_user_type, id_account, str_display_name, str_avatar_url, str_external_ref, dt_created, dt_updated
FROM %s
WHERE id = $1 AND id_tenant = $2 AND str_row_status IN %s`, s.usersTable(), visibleRowStatus)
	row := s.db.QueryRowContext(ctx, q, userID, tenantID)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return User{}, ErrNotFound.With("resource", "user")
	}
	return u, err
}

func (s *Store) GetUserByAccount(ctx context.Context, tenantID, accountID string) (User, error) {
	q := fmt.Sprintf(`
SELECT id, id_tenant, str_user_type, id_account, str_display_name, str_avatar_url, str_external_ref, dt_created, dt_updated
FROM %s
WHERE id_tenant = $1 AND id_account = $2 AND str_row_status IN %s`, s.usersTable(), visibleRowStatus)
	row := s.db.QueryRowContext(ctx, q, tenantID, accountID)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return User{}, ErrNotFound.With("resource", "user")
	}
	return u, err
}

func (s *Store) GetUserByExternalRef(ctx context.Context, tenantID, externalRef string) (User, error) {
	q := fmt.Sprintf(`
SELECT id, id_tenant, str_user_type, id_account, str_display_name, str_avatar_url, str_external_ref, dt_created, dt_updated
FROM %s
WHERE id_tenant = $1 AND str_external_ref = $2 AND str_row_status IN %s`, s.usersTable(), visibleRowStatus)
	row := s.db.QueryRowContext(ctx, q, tenantID, externalRef)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return User{}, ErrNotFound.With("resource", "user")
	}
	return u, err
}

func (s *Store) UpsertHumanUser(ctx context.Context, tenantID, accountID, displayName string) (User, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = "User"
	}
	q := fmt.Sprintf(`
INSERT INTO %s (id_tenant, str_user_type, id_account, str_display_name)
VALUES ($1, 'H', $2, $3)
ON CONFLICT (id_tenant, id_account) WHERE id_account IS NOT NULL
DO UPDATE SET str_display_name = COALESCE(NULLIF(EXCLUDED.str_display_name, ''), %s.str_display_name), dt_updated = NOW()
RETURNING id, id_tenant, str_user_type, id_account, str_display_name, str_avatar_url, str_external_ref, dt_created, dt_updated`,
		s.usersTable(), s.usersTable())
	return scanUser(s.db.QueryRowContext(ctx, q, tenantID, accountID, displayName))
}

func (s *Store) CreateAgentUser(ctx context.Context, tenantID, displayName, externalRef, avatarURL string) (User, error) {
	q := fmt.Sprintf(`
INSERT INTO %s (id_tenant, str_user_type, str_display_name, str_external_ref, str_avatar_url)
VALUES ($1, 'A', $2, NULLIF($3, ''), NULLIF($4, ''))
RETURNING id, id_tenant, str_user_type, id_account, str_display_name, str_avatar_url, str_external_ref, dt_created, dt_updated`,
		s.usersTable())
	return scanUser(s.db.QueryRowContext(ctx, q, tenantID, displayName, externalRef, avatarURL))
}

func (s *Store) RegisterDevice(ctx context.Context, userID, deviceKey, platform, pushToken string) (Device, error) {
	q := fmt.Sprintf(`
INSERT INTO %s (id_messenger_user, str_device_key, str_platform, str_push_token, dt_last_seen)
VALUES ($1, $2, $3, NULLIF($4, ''), NOW())
ON CONFLICT (id_messenger_user, str_device_key)
DO UPDATE SET str_platform = EXCLUDED.str_platform,
              str_push_token = COALESCE(EXCLUDED.str_push_token, %s.str_push_token),
              dt_last_seen = NOW(),
              dt_updated = NOW()
RETURNING id, id_messenger_user, str_device_key, str_platform, str_push_token, dt_last_seen`,
		s.devicesTable(), s.devicesTable())
	var d Device
	var push sql.NullString
	err := s.db.QueryRowContext(ctx, q, userID, deviceKey, platform, pushToken).Scan(
		&d.ID, &d.MessengerUserID, &d.DeviceKey, &d.Platform, &push, &d.LastSeen,
	)
	if push.Valid {
		d.PushToken = &push.String
	}
	return d, err
}

func (s *Store) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s WHERE id_channel = $1 AND id_messenger_user = $2 AND str_row_status IN %s`,
		s.membersTable(), visibleRowStatus)
	var one int
	err := s.db.QueryRowContext(ctx, q, channelID, userID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) addMember(ctx context.Context, tx *sql.Tx, channelID, userID, role string) error {
	q := fmt.Sprintf(`
INSERT INTO %s (id_channel, id_messenger_user, str_role)
VALUES ($1, $2, $3)
ON CONFLICT (id_channel, id_messenger_user) DO NOTHING`, s.membersTable())
	_, err := tx.ExecContext(ctx, q, channelID, userID, role)
	return err
}

func (s *Store) GetOrCreateDirect(ctx context.Context, tenantID, userA, userB, createdBy string) (Channel, error) {
	a, b := sortUserPair(userA, userB)
	var channelID string
	qLookup := fmt.Sprintf(`SELECT id_channel FROM %s WHERE id_tenant = $1 AND id_messenger_user_a = $2 AND id_messenger_user_b = $3`, s.directTable())
	err := s.db.QueryRowContext(ctx, qLookup, tenantID, a, b).Scan(&channelID)
	if err == nil {
		return s.GetChannelForUser(ctx, tenantID, channelID, createdBy)
	}
	if err != sql.ErrNoRows {
		return Channel{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Channel{}, err
	}
	defer tx.Rollback()

	channelID = platformuuid.NewString()
	qChannel := fmt.Sprintf(`
INSERT INTO %s (id, id_tenant, str_kind, id_created_by)
VALUES ($1, $2, 'D', $3)`, s.channelsTable())
	if _, err := tx.ExecContext(ctx, qChannel, channelID, tenantID, createdBy); err != nil {
		return Channel{}, err
	}

	qDirect := fmt.Sprintf(`
INSERT INTO %s (id_tenant, id_messenger_user_a, id_messenger_user_b, id_channel)
VALUES ($1, $2, $3, $4)`, s.directTable())
	if _, err := tx.ExecContext(ctx, qDirect, tenantID, a, b, channelID); err != nil {
		return Channel{}, err
	}

	for _, uid := range []string{userA, userB} {
		if err := s.addMember(ctx, tx, channelID, uid, MemberRoleMember); err != nil {
			return Channel{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Channel{}, err
	}
	return s.GetChannelForUser(ctx, tenantID, channelID, createdBy)
}

func (s *Store) CreateGroup(ctx context.Context, tenantID, title, createdBy string, memberIDs []string) (Channel, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Channel{}, err
	}
	defer tx.Rollback()

	channelID := platformuuid.NewString()
	qChannel := fmt.Sprintf(`
INSERT INTO %s (id, id_tenant, str_kind, str_title, id_created_by)
VALUES ($1, $2, 'G', $3, $4)`, s.channelsTable())
	if _, err := tx.ExecContext(ctx, qChannel, channelID, tenantID, title, createdBy); err != nil {
		return Channel{}, err
	}

	seen := map[string]struct{}{createdBy: {}}
	if err := s.addMember(ctx, tx, channelID, createdBy, MemberRoleAdmin); err != nil {
		return Channel{}, err
	}
	for _, uid := range memberIDs {
		if uid == "" || uid == createdBy {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		if err := s.addMember(ctx, tx, channelID, uid, MemberRoleMember); err != nil {
			return Channel{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Channel{}, err
	}
	return s.GetChannelForUser(ctx, tenantID, channelID, createdBy)
}

func (s *Store) ListChannelsForUser(ctx context.Context, tenantID, userID string) ([]Channel, error) {
	q := fmt.Sprintf(`
SELECT c.id, c.id_tenant, c.str_kind, c.str_title, c.str_avatar_url, c.id_created_by, c.dt_updated,
       COALESCE(m.i_unread_count, 0),
       (
         SELECT str_body FROM %s msg
         WHERE msg.id_channel = c.id AND msg.str_row_status IN %s
         ORDER BY msg.dt_created DESC LIMIT 1
       )
FROM %s c
JOIN %s m ON m.id_channel = c.id AND m.id_messenger_user = $2 AND m.str_row_status IN %s
WHERE c.id_tenant = $1 AND c.str_row_status IN %s
ORDER BY c.dt_updated DESC`,
		s.messagesTable(), visibleRowStatus,
		s.channelsTable(), s.membersTable(), visibleRowStatus, visibleRowStatus)

	rows, err := s.db.QueryContext(ctx, q, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Channel
	for rows.Next() {
		var ch Channel
		var title, avatar, preview sql.NullString
		if err := rows.Scan(&ch.ID, &ch.TenantID, &ch.Kind, &title, &avatar, &ch.CreatedByUserID, &ch.UpdatedAt, &ch.UnreadCount, &preview); err != nil {
			return nil, err
		}
		if title.Valid {
			ch.Title = &title.String
		}
		if avatar.Valid {
			ch.AvatarURL = &avatar.String
		}
		if preview.Valid {
			ch.LastPreview = &preview.String
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func (s *Store) GetChannelForUser(ctx context.Context, tenantID, channelID, userID string) (Channel, error) {
	ok, err := s.IsMember(ctx, channelID, userID)
	if err != nil {
		return Channel{}, err
	}
	if !ok {
		return Channel{}, ErrForbidden
	}

	q := fmt.Sprintf(`
SELECT c.id, c.id_tenant, c.str_kind, c.str_title, c.str_avatar_url, c.id_created_by, c.dt_updated,
       COALESCE(m.i_unread_count, 0)
FROM %s c
JOIN %s m ON m.id_channel = c.id AND m.id_messenger_user = $3
WHERE c.id = $1 AND c.id_tenant = $2 AND c.str_row_status IN %s AND m.str_row_status IN %s`,
		s.channelsTable(), s.membersTable(), visibleRowStatus, visibleRowStatus)

	var ch Channel
	var title, avatar sql.NullString
	err = s.db.QueryRowContext(ctx, q, channelID, tenantID, userID).Scan(
		&ch.ID, &ch.TenantID, &ch.Kind, &title, &avatar, &ch.CreatedByUserID, &ch.UpdatedAt, &ch.UnreadCount,
	)
	if err == sql.ErrNoRows {
		return Channel{}, ErrNotFound.With("resource", "channel")
	}
	if title.Valid {
		ch.Title = &title.String
	}
	if avatar.Valid {
		ch.AvatarURL = &avatar.String
	}

	members, err := s.ListChannelMembers(ctx, channelID)
	if err != nil {
		return Channel{}, err
	}
	ch.Members = members
	return ch, nil
}

func (s *Store) ListChannelMembers(ctx context.Context, channelID string) ([]Member, error) {
	q := fmt.Sprintf(`
SELECT m.id, m.id_channel, m.id_messenger_user, m.str_role, m.id_last_read_message, m.i_unread_count,
       u.id, u.id_tenant, u.str_user_type, u.id_account, u.str_display_name, u.str_avatar_url, u.str_external_ref, u.dt_created, u.dt_updated
FROM %s m
JOIN %s u ON u.id = m.id_messenger_user
WHERE m.id_channel = $1 AND m.str_row_status IN %s AND u.str_row_status IN %s`,
		s.membersTable(), s.usersTable(), visibleRowStatus, visibleRowStatus)

	rows, err := s.db.QueryContext(ctx, q, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Member
	for rows.Next() {
		var m Member
		var lastRead sql.NullString
		var accountID, avatar, ext sql.NullString
		if err := rows.Scan(
			&m.ID, &m.ChannelID, &m.MessengerUserID, &m.Role, &lastRead, &m.UnreadCount,
			&m.User.ID, &m.User.TenantID, &m.User.UserType, &accountID, &m.User.DisplayName, &avatar, &ext, &m.User.CreatedAt, &m.User.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if lastRead.Valid {
			m.LastReadMessage = &lastRead.String
		}
		if accountID.Valid {
			m.User.AccountID = &accountID.String
		}
		if avatar.Valid {
			m.User.AvatarURL = &avatar.String
		}
		if ext.Valid {
			m.User.ExternalRef = &ext.String
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ListMessages(ctx context.Context, channelID string, beforeID string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows *sql.Rows
	var err error
	base := fmt.Sprintf(`
SELECT msg.id, msg.id_channel, msg.str_kind, msg.id_sender_messenger_user, msg.str_body, msg.str_event_type, msg.jsn_event_payload,
       msg.str_client_ref, msg.id_reply_to, msg.dt_created,
       u.id, u.id_tenant, u.str_user_type, u.id_account, u.str_display_name, u.str_avatar_url, u.str_external_ref, u.dt_created, u.dt_updated
FROM %s msg
LEFT JOIN %s u ON u.id = msg.id_sender_messenger_user
WHERE msg.id_channel = $1 AND msg.str_row_status IN %s`,
		s.messagesTable(), s.usersTable(), visibleRowStatus)

	if beforeID != "" {
		q := base + ` AND msg.dt_created < (SELECT dt_created FROM ` + s.messagesTable() + ` WHERE id = $2)
ORDER BY msg.dt_created DESC LIMIT $3`
		rows, err = s.db.QueryContext(ctx, q, channelID, beforeID, limit)
	} else {
		q := base + ` ORDER BY msg.dt_created DESC LIMIT $2`
		rows, err = s.db.QueryContext(ctx, q, channelID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	var out []Message
	for rows.Next() {
		msg, err := scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}

func scanMessageRow(row scannable) (Message, error) {
	var msg Message
	var senderID, body, eventType, clientRef, replyTo sql.NullString
	var payload []byte
	var uID, uTenant, uType, uName sql.NullString
	var uAccount, uAvatar, uExt sql.NullString
	var uCreated, uUpdated sql.NullTime

	err := row.Scan(
		&msg.ID, &msg.ChannelID, &msg.Kind, &senderID, &body, &eventType, &payload, &clientRef, &replyTo, &msg.CreatedAt,
		&uID, &uTenant, &uType, &uAccount, &uName, &uAvatar, &uExt, &uCreated, &uUpdated,
	)
	if err != nil {
		return Message{}, err
	}
	if senderID.Valid {
		msg.SenderUserID = &senderID.String
	}
	if body.Valid {
		msg.Body = &body.String
	}
	if eventType.Valid {
		msg.EventType = &eventType.String
	}
	if len(payload) > 0 {
		msg.EventPayload = json.RawMessage(payload)
	}
	if clientRef.Valid {
		msg.ClientRef = &clientRef.String
	}
	if replyTo.Valid {
		msg.ReplyToID = &replyTo.String
	}
	if uID.Valid {
		u := User{
			ID: uID.String, TenantID: uTenant.String, UserType: uType.String, DisplayName: uName.String,
		}
		if uAccount.Valid {
			u.AccountID = &uAccount.String
		}
		if uAvatar.Valid {
			u.AvatarURL = &uAvatar.String
		}
		if uExt.Valid {
			u.ExternalRef = &uExt.String
		}
		if uCreated.Valid {
			u.CreatedAt = uCreated.Time
		}
		if uUpdated.Valid {
			u.UpdatedAt = uUpdated.Time
		}
		msg.Sender = &u
	}
	return msg, nil
}

type InsertMessageInput struct {
	ChannelID    string
	Kind         string
	SenderUserID *string
	Body         *string
	EventType    *string
	EventPayload json.RawMessage
	ClientRef    *string
	ReplyToID    *string
}

func (s *Store) InsertMessage(ctx context.Context, in InsertMessageInput) (Message, error) {
	if in.ClientRef != nil && *in.ClientRef != "" {
		existing, err := s.getMessageByClientRef(ctx, in.ChannelID, *in.ClientRef)
		if err == nil {
			return existing, nil
		}
		if !isNotFound(err) {
			return Message{}, err
		}
	}

	id := platformuuid.NewString()
	q := fmt.Sprintf(`
INSERT INTO %s (id, id_channel, id_sender_messenger_user, str_kind, str_body, str_event_type, jsn_event_payload, str_client_ref, id_reply_to)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`, s.messagesTable())

	var payload any
	if len(in.EventPayload) > 0 {
		payload = in.EventPayload
	}
	_, err := s.db.ExecContext(ctx, q, id, in.ChannelID, in.SenderUserID, in.Kind, in.Body, in.EventType, payload, in.ClientRef, in.ReplyToID)
	if err != nil {
		return Message{}, err
	}

	qBump := fmt.Sprintf(`UPDATE %s SET dt_updated = NOW() WHERE id = $1`, s.channelsTable())
	_, _ = s.db.ExecContext(ctx, qBump, in.ChannelID)

	if in.Kind == MessageKindUser && in.SenderUserID != nil {
		qUnread := fmt.Sprintf(`
UPDATE %s SET i_unread_count = i_unread_count + 1
WHERE id_channel = $1 AND id_messenger_user <> $2 AND str_row_status IN %s`,
			s.membersTable(), visibleRowStatus)
		_, _ = s.db.ExecContext(ctx, qUnread, in.ChannelID, *in.SenderUserID)
	}

	msgs, err := s.ListMessages(ctx, in.ChannelID, "", 1)
	if err != nil || len(msgs) == 0 {
		return Message{ID: id, ChannelID: in.ChannelID, Kind: in.Kind}, err
	}
	return msgs[0], nil
}

func (s *Store) getMessageByClientRef(ctx context.Context, channelID, clientRef string) (Message, error) {
	q := fmt.Sprintf(`
SELECT msg.id, msg.id_channel, msg.str_kind, msg.id_sender_messenger_user, msg.str_body, msg.str_event_type, msg.jsn_event_payload,
       msg.str_client_ref, msg.id_reply_to, msg.dt_created,
       u.id, u.id_tenant, u.str_user_type, u.id_account, u.str_display_name, u.str_avatar_url, u.str_external_ref, u.dt_created, u.dt_updated
FROM %s msg
LEFT JOIN %s u ON u.id = msg.id_sender_messenger_user
WHERE msg.id_channel = $1 AND msg.str_client_ref = $2 AND msg.str_row_status IN %s`,
		s.messagesTable(), s.usersTable(), visibleRowStatus)
	return scanMessageRow(s.db.QueryRowContext(ctx, q, channelID, clientRef))
}

func (s *Store) MarkRead(ctx context.Context, channelID, userID, messageID string) error {
	q := fmt.Sprintf(`
UPDATE %s
SET id_last_read_message = $3, dt_last_read = NOW(), i_unread_count = 0, dt_updated = NOW()
WHERE id_channel = $1 AND id_messenger_user = $2`, s.membersTable())
	res, err := s.db.ExecContext(ctx, q, channelID, userID, messageID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound.With("resource", "membership")
	}
	return nil
}

func (s *Store) MessagesAfter(ctx context.Context, channelID, afterMessageID string) ([]Message, error) {
	q := fmt.Sprintf(`
SELECT msg.id, msg.id_channel, msg.str_kind, msg.id_sender_messenger_user, msg.str_body, msg.str_event_type, msg.jsn_event_payload,
       msg.str_client_ref, msg.id_reply_to, msg.dt_created,
       u.id, u.id_tenant, u.str_user_type, u.id_account, u.str_display_name, u.str_avatar_url, u.str_external_ref, u.dt_created, u.dt_updated
FROM %s msg
LEFT JOIN %s u ON u.id = msg.id_sender_messenger_user
WHERE msg.id_channel = $1 AND msg.str_row_status IN %s
  AND msg.dt_created > COALESCE((SELECT dt_created FROM %s WHERE id = $2), '-infinity'::timestamptz)
ORDER BY msg.dt_created ASC LIMIT 200`,
		s.messagesTable(), s.usersTable(), visibleRowStatus, s.messagesTable())
	rows, err := s.db.QueryContext(ctx, q, channelID, nullIfEmpty(afterMessageID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func (s *Store) AddChannelMember(ctx context.Context, channelID, userID, role string) error {
	q := fmt.Sprintf(`
INSERT INTO %s (id_channel, id_messenger_user, str_role)
VALUES ($1, $2, $3)
ON CONFLICT (id_channel, id_messenger_user) DO NOTHING`, s.membersTable())
	_, err := s.db.ExecContext(ctx, q, channelID, userID, role)
	return err
}

func (s *Store) InsertHandoff(ctx context.Context, channelID, fromUserID, toUserID string, reason, summary *string) (Handoff, error) {
	id := platformuuid.NewString()
	q := fmt.Sprintf(`
INSERT INTO %s (id, id_channel, id_from_messenger_user, id_to_messenger_user, str_reason, str_context_summary)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, id_channel, id_from_messenger_user, id_to_messenger_user, str_reason, str_context_summary, dt_created`, s.handoffsTable())
	var h Handoff
	var r, ssum sql.NullString
	err := s.db.QueryRowContext(ctx, q, id, channelID, fromUserID, toUserID, reason, summary).Scan(
		&h.ID, &h.ChannelID, &h.FromMessengerUserID, &h.ToMessengerUserID, &r, &ssum, &h.CreatedAt,
	)
	if r.Valid {
		h.Reason = &r.String
	}
	if ssum.Valid {
		h.ContextSummary = &ssum.String
	}
	return h, err
}

func (s *Store) TouchDevice(ctx context.Context, deviceID string) error {
	q := fmt.Sprintf(`UPDATE %s SET dt_last_seen = NOW(), dt_updated = NOW() WHERE id = $1`, s.devicesTable())
	_, err := s.db.ExecContext(ctx, q, deviceID)
	return err
}

func (s *Store) GetDevice(ctx context.Context, deviceID, userID string) (Device, error) {
	q := fmt.Sprintf(`
SELECT id, id_messenger_user, str_device_key, str_platform, str_push_token, dt_last_seen
FROM %s WHERE id = $1 AND id_messenger_user = $2 AND str_row_status IN %s`, s.devicesTable(), visibleRowStatus)
	var d Device
	var push sql.NullString
	err := s.db.QueryRowContext(ctx, q, deviceID, userID).Scan(&d.ID, &d.MessengerUserID, &d.DeviceKey, &d.Platform, &push, &d.LastSeen)
	if err == sql.ErrNoRows {
		return Device{}, ErrNotFound.With("resource", "device")
	}
	if push.Valid {
		d.PushToken = &push.String
	}
	return d, err
}
