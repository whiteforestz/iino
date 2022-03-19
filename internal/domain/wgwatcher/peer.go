package wgwatcher

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/ini.v1"
)

func (d *Domain) updateUsage(ctx context.Context) error {
	usage, err := d.getCurrentUsagePeer(ctx)
	if err != nil {
		return fmt.Errorf("can't get current usage: %w", err)
	}

	enrichedUsage, err := d.enrichUsagePeer(usage)
	if err != nil {
		return fmt.Errorf("can't enrich usage: %w", err)
	}

	err = d.saveSnapshotPeerAccessor(castUsagePeerToSnapshotPeerAccessor(enrichedUsage))
	if err != nil {
		return fmt.Errorf("can't flush usage: %w", err)
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	d.usage.Peer = enrichedUsage

	return nil
}

func (d *Domain) getCurrentUsagePeer(ctx context.Context) ([]Peer, error) {
	peerNameAccessor, err := d.getPeerNameAccessor()
	if err != nil {
		return nil, fmt.Errorf("can't get peer accessor: %w", err)
	}

	raw, err := exec.CommandContext(ctx, d.cfg.Cmd, d.cfg.CmdArgs...).Output()
	if err != nil {
		return nil, fmt.Errorf("can't exec cmd: %w", err)
	}

	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("unexpected lines count: %d", len(lines))
	}

	usage := make([]Peer, 0, len(lines[1:]))
	for _, peerLine := range lines[1 : len(lines)-1] {
		p, err := extractPeerData(peerLine, peerNameAccessor)
		if err != nil {
			return nil, fmt.Errorf("can't extract peed data: %w", err)
		}

		usage = append(usage, *p)
	}

	sort.SliceStable(usage, func(i, j int) bool {
		return usage[i].LatestHandshakeUnix > usage[j].LatestHandshakeUnix
	})

	return usage, nil
}

func (d *Domain) getPeerNameAccessor() (map[string]string, error) {
	var (
		accessor = make(map[string]string)
	)

	err := fs.WalkDir(os.DirFS(d.cfg.ConfDirPath), ".", func(ep string, e fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("fs error at %q: %w", ep, err)
		}

		if e.IsDir() {
			return nil
		}

		match := d.cfg.ConfPatternRe.FindSubmatch([]byte(e.Name()))
		if len(match) != 2 {
			return fmt.Errorf("unexpected conf name: %q", e.Name())
		}

		var conf iniConf
		if err = ini.MapTo(&conf, path.Join(d.cfg.ConfDirPath, ep)); err != nil {
			return fmt.Errorf("can't map conf at %q: %w", ep, err)
		}

		accessor[conf.Peer.PresharedKey] = string(match[1])

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't walk dir at %q: %w", d.cfg.ConfDirPath, err)
	}

	return accessor, nil
}

func (d *Domain) enrichUsagePeer(usage []Peer) ([]Peer, error) {
	enrichedUsage := make([]Peer, 0, len(usage))
	for _, peer := range usage {
		snapshot, found := d.snapshotPeerAccessor[peer.Name]
		if found {
			if peer.LatestHandshakeUnix < snapshot.LatestHandshakeUnix {
				peer.LatestHandshakeUnix = snapshot.LatestHandshakeUnix
			}
		}

		enrichedUsage = append(enrichedUsage, peer)
	}

	sort.SliceStable(enrichedUsage, func(i, j int) bool {
		return enrichedUsage[i].LatestHandshakeUnix > enrichedUsage[j].LatestHandshakeUnix
	})

	return enrichedUsage, nil
}

func extractPeerData(peerLine string, peerNameAccessor map[string]string) (*Peer, error) {
	tokens := strings.FieldsFunc(peerLine, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(tokens) != 8 {
		return nil, fmt.Errorf("unexpected line content length: %d", len(tokens))
	}

	presharedKey, rawLatestHandshake := tokens[1], tokens[4]
	rawRx, rawTx := tokens[5], tokens[6]

	name, found := peerNameAccessor[presharedKey]
	if !found {
		return nil, fmt.Errorf("unknown peer with preshared key %q", presharedKey)
	}

	latestHandshakeUnix, err := strconv.ParseInt(rawLatestHandshake, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected time format: %w", err)
	}

	rx, err := strconv.ParseInt(rawRx, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected byte format: %w", err)
	}

	tx, err := strconv.ParseInt(rawTx, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected byte format: %w", err)
	}

	return &Peer{
		Name:                name,
		LatestHandshakeUnix: latestHandshakeUnix,
		TransferRx:          rx,
		TransferTx:          tx,
	}, nil
}
