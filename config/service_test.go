package config

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
)

func TestNewService(t *testing.T) {
	s, c := NewService()
	if s.configUpdatedChannel != c {
		t.Fatal("configUpdatedChannel was not set properly on NewService")
	}
}

func TestGetEnvironment(t *testing.T) {
	// Preserve existing env and restore on exit.
	const key = "SPEEDRUN_API_BASE"

	t.Run("defaults when env var is unset or empty", func(t *testing.T) {
		_ = os.Unsetenv(key)

		s := &Service{}
		got := s.GetEnvironment()

		if got == nil {
			t.Fatal("expected non-nil Service")
		}
		if got.SpeedRunAPIBase != "https://www.speedrun.com/api/v1" {
			t.Fatalf("expected default SpeedRunAPIBase %q, got %q", "https://www.speedrun.com/api/v1", got.SpeedRunAPIBase)
		}

		// Explicitly test empty-string behavior too.
		_ = os.Setenv(key, "")
		got2 := s.GetEnvironment()
		if got2.SpeedRunAPIBase != "https://www.speedrun.com/api/v1" {
			t.Fatalf("expected default SpeedRunAPIBase %q when env is empty, got %q", "https://www.speedrun.com/api/v1", got2.SpeedRunAPIBase)
		}
	})

	t.Run("uses env var when set", func(t *testing.T) {
		want := "http://localhost:1234/api"
		_ = os.Setenv(key, want)

		s := &Service{}
		got := s.GetEnvironment()

		if got == nil {
			t.Fatal("expected non-nil Service")
		}
		if got.SpeedRunAPIBase != want {
			t.Fatalf("expected SpeedRunAPIBase %q, got %q", want, got.SpeedRunAPIBase)
		}
	})
}

func TestUpdateKeyBinding(t *testing.T) {
	ch := make(chan *Service, 1)

	s := &Service{
		KeyConfig:            make(map[dispatcher.Command]keyinfo.KeyData),
		configUpdatedChannel: ch,
	}

	cmd := dispatcher.SPLIT
	data := keyinfo.KeyData{
		KeyCode:    32,
		LocaleName: "SPACE",
	}

	s.UpdateKeyBinding(cmd, data)

	// 1) Map updated
	got, ok := s.KeyConfig[cmd]
	if !ok {
		t.Fatalf("expected KeyConfig to contain command %v", cmd)
	}
	if !reflect.DeepEqual(got, data) {
		t.Fatalf("expected KeyConfig[%v] = %#v, got %#v", cmd, data, got)
	}

	// 2) UI update emitted (non-blocking send)
	select {
	case emitted := <-ch:
		if emitted != s {
			t.Fatalf("expected channel to receive service pointer %p, got %p", s, emitted)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for configUpdatedChannel send")
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	ch := make(chan *Service, 1)
	s := &Service{
		configUpdatedChannel: ch,
	}

	s.CreateDefaultConfig()

	if s.KeyConfig == nil {
		t.Fatal("expected KeyConfig to be initialized")
	}

	// Required commands should always be present
	required := []dispatcher.Command{
		dispatcher.SPLIT,
		dispatcher.UNDO,
		dispatcher.SKIP,
		dispatcher.PAUSE,
		dispatcher.RESET,
	}

	for _, cmd := range required {
		if _, ok := s.KeyConfig[cmd]; !ok {
			t.Fatalf("expected KeyConfig to contain %v", cmd)
		}
	}

	// UI update emitted (non-blocking send)
	select {
	case emitted := <-ch:
		if emitted != s {
			t.Fatalf("expected channel to receive service pointer %p, got %p", s, emitted)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for configUpdatedChannel send")
	}
}
