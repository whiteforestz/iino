package wgwatcher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func castPeerUsageToPeerAccessor(usage []PeerUsage) map[string]PeerUsage {
	accessor := make(map[string]PeerUsage, len(usage))
	for _, peer := range usage {
		accessor[peer.Name] = peer
	}

	return accessor
}

func castPeerAccessorToBinary(accessor map[string]PeerUsage) ([]byte, error) {
	b, err := json.Marshal(&accessor)
	if err != nil {
		return nil, fmt.Errorf("can't marshal: %w", err)
	}

	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}

func castPeerAccessorFromBinary(hash []byte) (map[string]PeerUsage, error) {
	b, err := base64.StdEncoding.DecodeString(string(hash))
	if err != nil {
		return nil, fmt.Errorf("can't decode base64: %w", err)
	}

	var accessor map[string]PeerUsage
	if err := json.Unmarshal(b, &accessor); err != nil {
		return nil, fmt.Errorf("cam't unmarshal: %w", err)
	}

	return accessor, nil
}
