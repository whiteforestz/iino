package wgwatcher

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/whiteforestz/iino/internal/domain/persistor"
)

const (
	tagUsagePeer = "usage_peer"
)

type snapshotPeer struct {
	LatestHandshakeUnix int64 `json:"latestHandshakeUnix,omitempty"`
}

func (d *Domain) saveSnapshotPeerAccessor(accessor map[string]snapshotPeer) error {
	b, err := castSnapshotPeerAccessorToBinary(accessor)
	if err != nil {
		return fmt.Errorf("can't cast accessor: %w", err)
	}

	if err = d.persistor.Save(tagUsagePeer, b); err != nil {
		return fmt.Errorf("can't save: %w", err)
	}

	return nil
}

func (d *Domain) loadSnapshotPeerAccessor() (map[string]snapshotPeer, error) {
	hash, err := d.persistor.Load(tagUsagePeer)
	if err != nil {
		if errors.Is(err, persistor.ErrNotExists) {
			return make(map[string]snapshotPeer), nil
		}

		return nil, fmt.Errorf("can't load persited data: %w", err)
	}

	accessor, err := castSnapshotPeerAccessorFromBinary(hash)
	if err != nil {
		return nil, fmt.Errorf("can't cast accessor: %w", err)
	}

	return accessor, nil
}

func castUsagePeerToSnapshotPeerAccessor(usage []Peer) map[string]snapshotPeer {
	accessor := make(map[string]snapshotPeer, len(usage))
	for _, peer := range usage {
		accessor[peer.Name] = snapshotPeer{
			LatestHandshakeUnix: peer.LatestHandshakeUnix,
		}
	}

	return accessor
}

func castSnapshotPeerAccessorToBinary(accessor map[string]snapshotPeer) ([]byte, error) {
	b, err := json.Marshal(&accessor)
	if err != nil {
		return nil, fmt.Errorf("can't marshal: %w", err)
	}

	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

func castSnapshotPeerAccessorFromBinary(hash []byte) (map[string]snapshotPeer, error) {
	b, err := base64.StdEncoding.DecodeString(string(hash))
	if err != nil {
		return nil, fmt.Errorf("can't decode base64: %w", err)
	}

	var accessor map[string]snapshotPeer
	if err := json.Unmarshal(b, &accessor); err != nil {
		return nil, fmt.Errorf("can't unmarshal: %w", err)
	}

	return accessor, nil
}
