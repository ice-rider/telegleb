package http

import (
	"testing"
	"time"
)

// Раскладка "2006-01-02T15:04:05Z" трактует Z как литерал и оставляет время
// локальным, из-за чего клиент получает сдвиг на таймзону сервера.
func TestFormatTimeIsAlwaysUTC(t *testing.T) {
	zone := time.FixedZone("UTC+4", 4*60*60)
	local := time.Date(2026, 9, 7, 17, 0, 0, 0, zone)

	got := formatTime(local)
	if want := "2026-09-07T13:00:00Z"; got != want {
		t.Errorf("получено %q, ожидалось %q", got, want)
	}
}

func TestFormatTimeParsesBackToSameInstant(t *testing.T) {
	original := time.Date(2026, 9, 7, 13, 0, 0, 0, time.UTC)
	parsed, err := time.Parse(time.RFC3339, formatTime(original))
	if err != nil {
		t.Fatalf("результат должен разбираться как RFC3339: %v", err)
	}
	if !parsed.Equal(original) {
		t.Errorf("момент времени изменился: %v против %v", parsed, original)
	}
}
