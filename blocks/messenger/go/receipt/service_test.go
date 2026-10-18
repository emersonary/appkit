package receipt

import (
	"context"
	"errors"
	"testing"
)

type repo struct {
	calls int
	err   error
	last  Record
}

func (r *repo) SaveReadReceipt(_ context.Context, channelID, userID, messageID, deviceID string) error {
	r.calls++
	r.last = Record{ChannelID: channelID, MessengerUserID: userID, MessageID: messageID, DeviceID: deviceID}
	return r.err
}

type members struct {
	calls int
	ids   []string
}

func (m *members) ListMemberIDs(context.Context, string) ([]string, error) {
	m.calls++
	return m.ids, nil
}

type publisher struct {
	calls int
	got   Record
}

func (p *publisher) PublishRead(_ context.Context, event Record) error {
	p.calls++
	p.got = event
	return nil
}

func TestSaveStoresOnceThenPublishesRecipients(t *testing.T) {
	r := &repo{}
	m := &members{ids: []string{"reader", "peer"}}
	p := &publisher{}
	err := Save(context.Background(), r, m, p, Record{
		ChannelID: "ch", MessageID: "m1", MessengerUserID: "reader", DeviceID: "d1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.calls != 1 || m.calls != 1 || p.calls != 1 {
		t.Fatalf("repo=%d members=%d publish=%d", r.calls, m.calls, p.calls)
	}
	if p.got.RecipientUserIDs[0] != "reader" || p.got.MessageID != "m1" {
		t.Fatalf("event=%+v", p.got)
	}
}

func TestSaveDoesNotPublishWhenStoreFails(t *testing.T) {
	r := &repo{err: errors.New("db")}
	m := &members{}
	p := &publisher{}
	if err := Save(context.Background(), r, m, p, Record{ChannelID: "ch", MessageID: "m1", MessengerUserID: "reader"}); err == nil {
		t.Fatal("expected error")
	}
	if p.calls != 0 || m.calls != 0 {
		t.Fatal("published a receipt that was not stored")
	}
}
