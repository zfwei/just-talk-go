//go:build linux

package voice

import (
	"fmt"
	"os/exec"
	"strings"
)

func pickCommand() (*exec.Cmd, string, error) {
	return pickCommandWithDevice("")
}

func pickCommandWithDevice(device string) (*exec.Cmd, string, error) {
	// ALSA arecord: reliable s16le output, proper format conversion
	al := []string{"-r", "16000", "-f", "S16_LE", "-c", "1", "-t", "raw"}
	if device != "" {
		al = append(al, "-D", device)
	}
	al = append(al, "-")

	return firstFound([]candidate{
		{"arecord", al},
	})
}

func errNoBackend(candidates []candidate) error {
	return fmt.Errorf("no recording backend found; install alsa-utils for arecord")
}

// ListDevices returns available audio input devices.
func ListDevices() ([]string, error) {
	if al, err := exec.LookPath("arecord"); err == nil {
		cmd := exec.Command(al, "-L")
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("arecord list input devices: %w", err)
		}
		devices := parseALSADevices(string(out))
		if len(devices) > 0 {
			return devices, nil
		}
	}
	return nil, fmt.Errorf("no input devices found; install alsa-utils for arecord")
}

func parseALSADevices(output string) []string {
	var devices []string
	// ALSA -L lists device descriptions; extract hw: or plughw: entries
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "hw:") || strings.HasPrefix(line, "plughw:") ||
			strings.HasPrefix(line, "default") {
			devices = append(devices, line)
		}
	}
	return devices
}
