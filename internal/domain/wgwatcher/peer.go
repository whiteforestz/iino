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

func (d *Domain) updatePeerUsage(ctx context.Context) error {
	usage, err := d.getCurrentPeerUsage(ctx)
	if err != nil {
		return fmt.Errorf("can't get current usage: %w", err)
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	d.usage.Peer = usage

	return nil
}

func (d *Domain) getCurrentPeerUsage(ctx context.Context) ([]PeerUsage, error) {
	peerAccessor, err := d.getPeerAccessor()
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

	usage := make([]PeerUsage, 0, len(lines[1:]))
	for _, peerLine := range lines[1 : len(lines)-1] {
		pu, err := extractPeerData(peerLine, peerAccessor)
		if err != nil {
			return nil, fmt.Errorf("can't extract peed data: %w", err)
		}

		usage = append(usage, *pu)
	}

	sort.SliceStable(usage, func(i, j int) bool {
		return usage[i].LatestHandshakeUnix > usage[j].LatestHandshakeUnix
	})

	return usage, nil
}

func (d *Domain) getPeerAccessor() (map[string]string, error) {
	var (
		accessor = make(map[string]string)
	)

	err := fs.WalkDir(os.DirFS(d.cfg.ConfDirPath), ".", func(dePath string, de fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("fs error at %q: %w", dePath, err)
		}

		if de.IsDir() {
			return nil
		}

		match := d.cfg.ConfPatternRe.FindSubmatch([]byte(de.Name()))
		if len(match) != 2 {
			return fmt.Errorf("unexpected conf name: %q", de.Name())
		}

		var conf iniConf
		if err = ini.MapTo(&conf, path.Join(d.cfg.ConfDirPath, dePath)); err != nil {
			return fmt.Errorf("can't map conf at %q: %w", dePath, err)
		}

		accessor[conf.Peer.PresharedKey] = string(match[1])

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't walk dir at %q: %w", d.cfg.ConfDirPath, err)
	}

	return accessor, nil
}

func extractPeerData(peerLine string, peerAccessor map[string]string) (*PeerUsage, error) {
	tokens := strings.FieldsFunc(peerLine, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(tokens) != 8 {
		return nil, fmt.Errorf("unexpected line content length: %d", len(tokens))
	}

	presharedKey, latestHandshakeRaw := tokens[1], tokens[4]

	name, found := peerAccessor[presharedKey]
	if !found {
		return nil, fmt.Errorf("unknown peer with preshared key %q", presharedKey)
	}

	latestHandshakeUnix, err := strconv.ParseInt(latestHandshakeRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unexpected time format: %w", err)
	}

	return &PeerUsage{
		Name:                name,
		LatestHandshakeUnix: latestHandshakeUnix,
	}, nil
}
