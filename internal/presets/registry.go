package presets

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/Tanq16/goff/internal/ops"
	"github.com/Tanq16/goff/internal/probe"
)

type Preset struct {
	ID          string
	Name        string
	Description string
	Category    ops.OpCategory
	Build       func(input string, p *probe.ProbeResult) (*ops.OpResult, error)
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Preset)
	order    []string
)

func Register(p Preset) {
	mu.Lock()
	defer mu.Unlock()
	key := strings.ToLower(p.ID)
	registry[key] = p
	if !slices.Contains(order, key) {
		order = append(order, key)
	}
}

func Get(id string) (*Preset, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := registry[strings.ToLower(id)]
	if !ok {
		return nil, false
	}
	return &p, true
}

func List() []Preset {
	mu.RLock()
	defer mu.RUnlock()
	var list []Preset
	for _, key := range order {
		list = append(list, registry[key])
	}
	return list
}

func ListByCategory(cat ops.OpCategory) []Preset {
	mu.RLock()
	defer mu.RUnlock()
	var list []Preset
	for _, key := range order {
		p := registry[key]
		if p.Category == cat {
			list = append(list, p)
		}
	}
	return list
}

func MatchPresetForMedia(p *probe.ProbeResult) []Preset {
	if p == nil {
		return List()
	}
	if p.IsVideo() {
		return ListByCategory(ops.CategoryVideo)
	}
	if p.IsAudioOnly() {
		return ListByCategory(ops.CategoryAudio)
	}
	return List()
}

func AvailablePresetIDs() string {
	all := List()
	ids := make([]string, len(all))
	for i, p := range all {
		ids[i] = p.ID
	}
	return fmt.Sprintf("[%s]", strings.Join(ids, ", "))
}
