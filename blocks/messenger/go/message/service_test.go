package message

import (
	"context"
	"errors"
	"testing"
)

type members struct {
	calls int
	ids   []string
	err   error
}

func (m *members) ListMemberIDs(context.Context, string) ([]string, error) {
	m.calls++
	return m.ids, m.err
}

type publisher struct {
	calls int
	got   Created
	err   error
}

func (p *publisher) PublishCreated(_ context.Context, event Created) error {
	p.calls++
	p.got = event
	return p.err
}

func TestPublishOnceResolvesMembersASingleTime(t *testing.T) {
	m := &members{ids: []string{"a", "b", "c"}}
	p := &publisher{}
	err := PublishOnce(context.Background(), m, p, Created{ChannelID: "ch", TenantID: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if m.calls != 1 || p.calls != 1 {
		t.Fatalf("members=%d publish=%d", m.calls, p.calls)
	}
	if len(p.got.RecipientUserIDs) != 3 || p.got.RecipientUserIDs[1] != "b" {
		t.Fatalf("recipients=%v", p.got.RecipientUserIDs)
	}
}

func TestPublishOnceSkipsWhenPublisherMissing(t *testing.T) {
	m := &members{ids: []string{"a"}}
	if err := PublishOnce(context.Background(), m, nil, Created{ChannelID: "ch"}); err != nil {
		t.Fatal(err)
	}
	if m.calls != 0 {
		t.Fatalf("listed members without a publisher: %d", m.calls)
	}
}

func TestPublishOnceStopsWhenMembersFail(t *testing.T) {
	m := &members{err: errors.New("db")}
	p := &publisher{}
	if err := PublishOnce(context.Background(), m, p, Created{ChannelID: "ch"}); err == nil {
		t.Fatal("expected error")
	}
	if p.calls != 0 {
		t.Fatal("published after member lookup failed")
	}
}
