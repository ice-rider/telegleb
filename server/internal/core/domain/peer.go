package domain

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidPeerRef = errors.New("invalid peer ref")

// PeerKind различает адресуемые сущности Telegram. От вида зависит, какой
// InputPeer нужно собрать для MTProto.
type PeerKind string

const (
	PeerUser    PeerKind = "user"
	PeerChat    PeerKind = "chat"
	PeerChannel PeerKind = "channel"
)

// Peer — полный адрес чата. MTProto требует пару (id, accessHash) для всего,
// кроме обычных групп, поэтому числового идентификатора недостаточно.
type Peer struct {
	Kind       PeerKind
	ID         int64
	AccessHash int64
}

// Ref сериализует адрес в непрозрачную для клиента строку.
//
//	user:<id>:<accessHash>
//	chat:<id>
//	channel:<id>:<accessHash>
func (p Peer) Ref() string {
	id := strconv.FormatInt(p.ID, 10)
	if p.Kind == PeerChat {
		return string(p.Kind) + ":" + id
	}
	return string(p.Kind) + ":" + id + ":" + strconv.FormatInt(p.AccessHash, 10)
}

func ParsePeer(ref string) (Peer, error) {
	parts := strings.Split(ref, ":")
	if len(parts) < 2 {
		return Peer{}, ErrInvalidPeerRef
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Peer{}, ErrInvalidPeerRef
	}

	switch PeerKind(parts[0]) {
	case PeerChat:
		if len(parts) != 2 {
			return Peer{}, ErrInvalidPeerRef
		}
		return Peer{Kind: PeerChat, ID: id}, nil
	case PeerUser, PeerChannel:
		if len(parts) != 3 {
			return Peer{}, ErrInvalidPeerRef
		}
		hash, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return Peer{}, ErrInvalidPeerRef
		}
		return Peer{Kind: PeerKind(parts[0]), ID: id, AccessHash: hash}, nil
	default:
		return Peer{}, ErrInvalidPeerRef
	}
}
