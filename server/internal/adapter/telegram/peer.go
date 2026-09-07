package telegram

import (
	"strconv"

	"telegleb/internal/core/domain"

	"github.com/gotd/td/tg"
)

// inputPeer собирает MTProto-адрес из доменного Peer. Именно наличие
// AccessHash отличает рабочий запрос от PEER_ID_INVALID.
func inputPeer(p domain.Peer) tg.InputPeerClass {
	switch p.Kind {
	case domain.PeerChat:
		return &tg.InputPeerChat{ChatID: p.ID}
	case domain.PeerChannel:
		return &tg.InputPeerChannel{ChannelID: p.ID, AccessHash: p.AccessHash}
	default:
		return &tg.InputPeerUser{UserID: p.ID, AccessHash: p.AccessHash}
	}
}

func inputChannel(p domain.Peer) *tg.InputChannel {
	return &tg.InputChannel{ChannelID: p.ID, AccessHash: p.AccessHash}
}

// peerKey — идентичность пира без accessHash. Хеш у одного и того же чата
// может отличаться между ответами (например, в описании папки он нулевой),
// поэтому сравнивать пиры можно только по виду и id.
func peerKey(kind domain.PeerKind, id int64) string {
	return string(kind) + ":" + strconv.FormatInt(id, 10)
}

func peerClassKey(p tg.PeerClass) (string, bool) {
	switch v := p.(type) {
	case *tg.PeerUser:
		return peerKey(domain.PeerUser, v.UserID), true
	case *tg.PeerChat:
		return peerKey(domain.PeerChat, v.ChatID), true
	case *tg.PeerChannel:
		return peerKey(domain.PeerChannel, v.ChannelID), true
	default:
		return "", false
	}
}

// inputPeerClassKey приводит пир из описания папки к тому же ключу.
// InputPeerSelf встречается у «Избранного» и требует собственный id.
func inputPeerClassKey(p tg.InputPeerClass, ownID int64) (string, bool) {
	switch v := p.(type) {
	case *tg.InputPeerUser:
		return peerKey(domain.PeerUser, v.UserID), true
	case *tg.InputPeerChat:
		return peerKey(domain.PeerChat, v.ChatID), true
	case *tg.InputPeerChannel:
		return peerKey(domain.PeerChannel, v.ChannelID), true
	case *tg.InputPeerSelf:
		return peerKey(domain.PeerUser, ownID), true
	default:
		return "", false
	}
}
