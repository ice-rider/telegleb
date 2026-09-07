package domain

import "testing"

func TestPeerRefRoundTrip(t *testing.T) {
	cases := []Peer{
		{Kind: PeerUser, ID: 123456789, AccessHash: -8901234567890123456},
		{Kind: PeerChannel, ID: 1, AccessHash: 9223372036854775807},
		{Kind: PeerChat, ID: 42},
	}

	for _, want := range cases {
		got, err := ParsePeer(want.Ref())
		if err != nil {
			t.Fatalf("%s: неожиданная ошибка %v", want.Ref(), err)
		}
		if got != want {
			t.Errorf("%s: получено %+v, ожидалось %+v", want.Ref(), got, want)
		}
	}
}

// accessHash не помещается в double, поэтому ref остаётся строкой и точность
// не должна теряться при разборе.
func TestPeerRefKeepsFullInt64(t *testing.T) {
	p := Peer{Kind: PeerChannel, ID: 2, AccessHash: -9223372036854775808}
	got, err := ParsePeer(p.Ref())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got.AccessHash != p.AccessHash {
		t.Errorf("accessHash потерян: получено %d, ожидалось %d", got.AccessHash, p.AccessHash)
	}
}

func TestParsePeerRejectsMalformed(t *testing.T) {
	bad := []string{"", "user", "user:abc:1", "user:1", "chat:1:2", "group:1:2", "channel:1"}
	for _, ref := range bad {
		if _, err := ParsePeer(ref); err == nil {
			t.Errorf("%q должен быть отвергнут", ref)
		}
	}
}

func TestMediaRefRoundTrip(t *testing.T) {
	want := MediaRef{
		Peer:      Peer{Kind: PeerChannel, ID: 123, AccessHash: -456},
		MessageID: 78901,
		Kind:      MediaPhoto,
	}

	got, err := ParseMediaRef(want.String())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got != want {
		t.Errorf("получено %+v, ожидалось %+v", got, want)
	}
}

// Разделитель "~" выбран, чтобы ref оставался одним сегментом URL.
func TestMediaRefHasNoSlash(t *testing.T) {
	ref := MediaRef{Peer: Peer{Kind: PeerUser, ID: 1, AccessHash: 2}, MessageID: 3, Kind: MediaVideo}
	for _, r := range ref.String() {
		if r == '/' {
			t.Fatalf("ref не должен содержать слеш: %s", ref.String())
		}
	}
}
