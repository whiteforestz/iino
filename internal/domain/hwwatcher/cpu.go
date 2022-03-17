package hwwatcher

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/pkg/logger"
)

const (
	prefixCPU = "cpu"
)

func (d *Domain) updateCPUUsage(lastCPULoad []cpuCoreLoad) ([]cpuCoreLoad, error) {
	cpuLoad, err := d.getCurrentCPULoad()
	if err != nil {
		return nil, fmt.Errorf("can't get current cpu load: %w", err)
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	if len(lastCPULoad) == 0 {
		return cpuLoad, nil
	}

	if len(cpuLoad) != len(lastCPULoad) {
		return nil, errors.New("inconsistent length")
	}

	coreLoadAccessor := make(map[string]cpuCoreLoad, len(cpuLoad))
	for _, coreLoad := range cpuLoad {
		coreLoadAccessor[coreLoad.Slug] = coreLoad
	}

	cpuUsage := make([]CPUCoreUsage, 0, len(cpuLoad))
	for idx, lastCoreLoad := range lastCPULoad {
		coreLoad, found := coreLoadAccessor[lastCoreLoad.Slug]
		if !found {
			return nil, fmt.Errorf("last cpu load %q not found", lastCoreLoad.Slug)
		}

		diffTotal := coreLoad.GetTotal() - lastCoreLoad.GetTotal()
		diffIdle := coreLoad.GetTotalIdle() - lastCoreLoad.GetTotalIdle()

		var lastCoreUsage int64
		if len(d.usage.CPU) != 0 {
			lastCoreUsage = d.usage.CPU[idx].Percentage
		}

		var coreUsage int64
		if diffTotal != 0 {
			coreUsage = 100 * (diffTotal - diffIdle) / diffTotal
		}

		cpuUsage = append(cpuUsage, CPUCoreUsage{
			Slug:       lastCoreLoad.Slug,
			Percentage: (lastCoreUsage + coreUsage) / 2,
		})
	}

	d.usage.CPU = cpuUsage

	logger.Instance().Debug("cpu usage updated", zap.Any("cpuUsage", cpuUsage))

	return cpuLoad, nil
}

func (d *Domain) getCurrentCPULoad() ([]cpuCoreLoad, error) {
	raw, err := loadMagicFile(d.cfg.CPULoadSourcePath)
	if err != nil {
		return nil, fmt.Errorf("can't load magic file: %w", err)
	}
	lines := strings.Split(raw, "\n")

	var cpuLoad []cpuCoreLoad
	for _, line := range lines {
		if !strings.HasPrefix(line, prefixCPU) {
			break
		}

		coreLoad, err := extractCPUCoreLoad(line)
		if err != nil {
			return nil, fmt.Errorf("can't extract core load: %w", err)
		}

		cpuLoad = append(cpuLoad, *coreLoad)
	}
	if len(cpuLoad) == 0 {
		return nil, errors.New("core lines not found")
	}

	return cpuLoad, nil
}

func extractCPUCoreLoad(cpuLine string) (*cpuCoreLoad, error) {
	coreTokens := strings.FieldsFunc(cpuLine, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	if len(coreTokens) != 11 {
		return nil, fmt.Errorf("invalid length: %d", len(coreTokens))
	}

	var (
		coreSlug = coreTokens[0]
		recs     = make([]int64, 0, 10)
	)
	for idx, rec := range coreTokens[1:] {
		parsed, err := strconv.ParseInt(rec, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid format at %d: %w", idx, err)
		}

		recs = append(recs, parsed)
	}

	return &cpuCoreLoad{
		Slug:      coreSlug,
		User:      recs[0],
		Nice:      recs[1],
		System:    recs[2],
		Idle:      recs[3],
		IOWait:    recs[4],
		IRq:       recs[5],
		SoftIRq:   recs[6],
		Steal:     recs[7],
		Guest:     recs[8],
		GuestNice: recs[9],
	}, nil
}
