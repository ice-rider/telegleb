package telegram

import (
	"testing"

	"github.com/gotd/td/tg"
)

// У каналов нумерация сообщений своя и начинается с единицы, поэтому плоская
// карта по msg.ID склеивает сообщения разных чатов и подставляет в список
// диалогов чужое превью.
func TestMessagesAreKeyedPerChat(t *testing.T) {
	d := newDicts(nil, nil, 0)
	d.addMessages([]tg.MessageClass{
		&tg.Message{ID: 1, Message: "из первого канала", PeerID: &tg.PeerChannel{ChannelID: 100}},
		&tg.Message{ID: 1, Message: "из второго канала", PeerID: &tg.PeerChannel{ChannelID: 200}},
		&tg.Message{ID: 1, Message: "из личного чата", PeerID: &tg.PeerUser{UserID: 300}},
	})

	if len(d.msgs) != 3 {
		t.Fatalf("ожидалось 3 сообщения в карте, получено %d", len(d.msgs))
	}

	cases := []struct {
		peer string
		want string
	}{
		{peerKey("channel", 100), "из первого канала"},
		{peerKey("channel", 200), "из второго канала"},
		{peerKey("user", 300), "из личного чата"},
	}
	for _, c := range cases {
		got, ok := d.msgs[msgKey{peer: c.peer, msgID: 1}]
		if !ok {
			t.Fatalf("сообщение для %s не найдено", c.peer)
		}
		if got.Message != c.want {
			t.Errorf("для %s ожидалось %q, получено %q", c.peer, c.want, got.Message)
		}
	}
}

func TestUserTitleFallsBackToUsername(t *testing.T) {
	if got := userTitle(&tg.User{Username: "gleb"}, false); got != "gleb" {
		t.Errorf("ожидался username, получено %q", got)
	}
	if got := userTitle(&tg.User{FirstName: "Глеб", LastName: "А"}, false); got != "Глеб А" {
		t.Errorf("ожидалось полное имя, получено %q", got)
	}
	if got := userTitle(&tg.User{FirstName: "Глеб"}, true); got != "Избранное" {
		t.Errorf("собственный чат должен называться Избранное, получено %q", got)
	}
}
