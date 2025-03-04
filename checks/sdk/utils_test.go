package sdk

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseVersionRange(t *testing.T) {
	tests := []struct {
		name        string
		giveRange   string
		giveVersion string
	}{
		{
			name:        "no upper limit",
			giveRange:   "[0.9.16,)",
			giveVersion: "1.5.16",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionRange(tt.giveRange)
			if err != nil {
				t.Errorf("ParseVersionRange() error = %v", err)
				return
			}
			assert.True(t, got.matches(tt.giveVersion))
		})
	}
}
