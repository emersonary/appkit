package messenger

import (
	"context"
	"database/sql"
	"fmt"

	platformuuid "github.com/emersonary/appkit/uuid"
)

func (s *Store) SaveReadReceipt(ctx context.Context, channelID, userID, messageID, deviceID string) (ReadReceipt, error) {
	if err := s.MarkRead(ctx, channelID, userID, messageID); err != nil {
		return ReadReceipt{}, err
	}
	id := platformuuid.NewString()
	var device any
	if deviceID != "" {
		device = deviceID
	}
	q := fmt.Sprintf(`
INSERT INTO %s (id, id_message, id_channel, id_messenger_user, id_device, dt_read)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (id_message, id_messenger_user) DO UPDATE
SET dt_read = NOW(), id_device = EXCLUDED.id_device
RETURNING id, id_message, id_channel, id_messenger_user, id_device, dt_read`, s.receiptsTable())
	var rec ReadReceipt
	var dev sql.NullString
	err := s.db.QueryRowContext(ctx, q, id, messageID, channelID, userID, device).Scan(
		&rec.ID, &rec.MessageID, &rec.ChannelID, &rec.MessengerUserID, &dev, &rec.ReadAt,
	)
	if err != nil {
		return ReadReceipt{}, err
	}
	if dev.Valid {
		rec.DeviceID = &dev.String
	}
	return rec, nil
}

func (s *Store) ListReadReceipts(ctx context.Context, channelID, messageID string) ([]ReadReceipt, error) {
	q := fmt.Sprintf(`
SELECT id, id_message, id_channel, id_messenger_user, id_device, dt_read
FROM %s
WHERE id_channel = $1 AND id_message = $2
ORDER BY dt_read ASC`, s.receiptsTable())
	rows, err := s.db.QueryContext(ctx, q, channelID, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReadReceipt
	for rows.Next() {
		var rec ReadReceipt
		var dev sql.NullString
		if err := rows.Scan(&rec.ID, &rec.MessageID, &rec.ChannelID, &rec.MessengerUserID, &dev, &rec.ReadAt); err != nil {
			return nil, err
		}
		if dev.Valid {
			rec.DeviceID = &dev.String
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *Store) ListMemberIDs(ctx context.Context, channelID string) ([]string, error) {
	members, err := s.ListChannelMembers(ctx, channelID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.MessengerUserID)
	}
	return ids, nil
}
