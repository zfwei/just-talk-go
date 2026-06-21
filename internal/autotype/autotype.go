package autotype

import (
	"log/slog"
)

// Paste inserts text into the currently focused input field.
func Paste(text string, logger *slog.Logger) error {
	return pastePlatform(text, logger)
}

// Backspace deletes text by simulating backspace keypresses.
func Backspace(count int, logger *slog.Logger) error {
	return backspacePlatform(count, logger)
}
