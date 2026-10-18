package channel

import (
	"context"
	"testing"
)

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
	got   MembersUpdated
}

func (p *publisher) PublishMembers(_ context.Context, event MembersUpdated) error {
	p.calls++
	p.got = event
	return nil
}

func TestPublishMembersOnce(t *testing.T) {
	m := &members{ids: []string{"a", "b"}}
	p := &publisher{}
	if err := PublishMembersOnce(context.Background(), m, p, "ch"); err != nil {
		t.Fatal(err)
	}
	if m.calls != 1 || p.calls != 1 || len(p.got.RecipientUserIDs) != 2 {
		t.Fatalf("members=%d publish=%d event=%+v", m.calls, p.calls, p.got)
	}
}
