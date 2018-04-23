package gaelog

import (
	"time"
)

// RotationStrategy is a strategy for rotating the log file.
type RotationStrategy interface {
	// ShouldRotate checks current log file whether it
	// should be rotated.
	ShouldRotate(fi FileInfo) bool
}

// RotateEvery rotates the file periodically.
type RotateEvery struct {
	// Duration is a lifetime of one file.
	Duration time.Duration
}

// ShouldRotate implements RotationStrategy
func (s RotateEvery) ShouldRotate(fi FileInfo) bool {
	now := time.Now()
	if fi.CreatedAt.Before(now.Add(-s.Duration)) {
		return true
	}
	return false
}

// NeverRotate never rotates the file.
type NeverRotate struct{}

// ShouldRotate implements RotationStrategy
func (NeverRotate) ShouldRotate(fi FileInfo) bool {
	return false
}
