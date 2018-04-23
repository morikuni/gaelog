package gaelog

import (
	"context"
	"os"
	"time"
)

// CleanupStrategy is a strategy for cleaning up the old log files.
type CleanupStrategy interface {
	// Apply filters a files based on the strategy.
	Apply(fis []FileInfo) ([]FileInfo, error)
}

// RemoveCreatedBefore is a strategy that cleans up the files based on
// the created time.
type RemoveCreatedBefore struct {
	// Duration is a threshold. The files created before this duration
	// are removed.
	Duration time.Duration
}

// Apply implements CleanupStrategy.
func (s RemoveCreatedBefore) Apply(fis []FileInfo) ([]FileInfo, error) {
	now := time.Now()
	result := make([]FileInfo, 0, len(fis))
	for _, fi := range fis {
		if fi.CreatedAt.Before(now.Add(-s.Duration)) {
			result = append(result, fi)
		}
	}
	return result, nil
}

// RemoveAll is a strategy that cleans up all files.
type RemoveAll struct{}

// Apply implements CleanupStrategy.
func (RemoveAll) Apply(fis []FileInfo) ([]FileInfo, error) {
	return fis, nil
}

// ScheduleCleanup watches the old log files periodically and remove them based on the strategy.
func ScheduleCleanup(ctx context.Context, interval time.Duration, logger *CustomLogger, strategy CleanupStrategy) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		CleanUp(logger, strategy)
	}
}

// CleanUp cleans up the old log files of the logger based on the strategy.
func CleanUp(logger *CustomLogger, strategy CleanupStrategy) {
	fis, err := strategy.Apply(logger.removableFiles())
	if err != nil {
		logger.Errorf(nil, "strategy error: %s", err)
	}

	for _, fi := range fis {
		err := os.Remove(fi.Path())
		if err != nil {
			logger.Errorf(nil, "could not remove a file: %s: %s", fi.Path(), err)
		}
	}
}
