package escalation

import (
	"encoding/json"
	"strings"
)

// CloneEscalationPath returns a deep copy of an EscalationPath. It isolates
// the steps slice, the per-step channels slice, and the per-channel JSON
// Config so callers can mutate any of the cloned fields without affecting
// the source.
func CloneEscalationPath(input EscalationPath) EscalationPath {
	steps := make([]EscalationStep, 0, len(input.Steps))
	for _, step := range input.Steps {
		channels := make([]ChannelConfig, 0, len(step.Channels))
		for _, channel := range step.Channels {
			var cfg json.RawMessage
			if channel.Config != nil {
				cfg = append(json.RawMessage(nil), channel.Config...)
			}
			channels = append(channels, ChannelConfig{Type: channel.Type, Target: strings.TrimSpace(channel.Target), Config: cfg})
		}
		steps = append(steps, EscalationStep{ChannelID: strings.TrimSpace(step.ChannelID), DelayMinutes: step.DelayMinutes, Channels: channels})
	}
	return EscalationPath{Steps: steps}
}
