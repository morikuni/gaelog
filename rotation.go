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

// TimeBaseRotation rotates the file periodically.
type TimeBaseRotation struct {
	// MaxAge is a lifetime of the file.
	MaxAge time.Duration
}

// ShouldRotate implements RotationStrategy
func (s TimeBaseRotation) ShouldRotate(fi FileInfo) bool {
	now := time.Now()
	if fi.CreatedAt.Before(now.Add(-s.MaxAge)) {
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
