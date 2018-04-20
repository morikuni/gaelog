package gaelog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

const timeFormat = "20060102150405"

var _ interface {
	Logger
} = &CustomLogger{}

// CustomLogger is a logger for the custom runtime on flex environment.
// The logs are written into a file dir/app*.log.
// These logs are collected by the fluentd running in another container
// on the GCE instance.
type CustomLogger struct {
	// Dir is a directory which the log files are created.
	Dir string

	// OnUnexpectedError is called when a logger cannot put the logs by some reason.
	OnUnexpectedError func(err error)

	// RotationStrategy is used to check whether logger should rotate
	// the log file.
	RotationStrategy RotationStrategy

	once      sync.Once
	mu        sync.Mutex
	file      *os.File
	createdAt time.Time
}

// Criticalf implements Logger.
func (l *CustomLogger) Criticalf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Critical, format, args...)
}

// Errorf implements Logger.
func (l *CustomLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Error, format, args...)
}

// Warningf implements Logger.
func (l *CustomLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Warning, format, args...)
}

// Infof implements Logger.
func (l *CustomLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Info, format, args...)
}

// Debugf implements Logger.
func (l *CustomLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Debug, format, args...)
}

// Printf implements Logger.
func (l *CustomLogger) Printf(_ context.Context, level LogLevel, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	now := time.Now()
	payload := logPayload{
		Timestamp: struct {
			Seconds int64 `json:"seconds"`
			Nanos   int   `json:"nanos"`
		}{
			Seconds: now.Unix(),
			Nanos:   now.Nanosecond(),
		},
		Severity: level,
		Message:  msg,
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.init()
	l.tryRotate()

	if err := json.NewEncoder(l.file).Encode(&payload); err != nil {
		l.recoverError(err)
	}
}

func (l *CustomLogger) recoverError(err error) {
	if l.OnUnexpectedError != nil {
		l.OnUnexpectedError(err)
	}
}

func (l *CustomLogger) tryRotate() {
	stat, err := l.file.Stat()
	if err != nil {
		l.recoverError(err)
	}
	if l.RotationStrategy.ShouldRotate(FileInfo{l.Dir, stat.Name(), stat.Size(), stat.ModTime(), l.createdAt}) {
		if err := l.rotate(); err != nil {
			l.recoverError(err)
		}
	}
}

func (l *CustomLogger) init() {
	l.once.Do(func() {
		if l.Dir == "" {
			l.Dir = "/var/log/app_engine/"
		}
		if l.file == nil {
			if err := l.rotate(); err != nil {
				l.recoverError(err)
			}
		}
		if l.RotationStrategy == nil {
			l.RotationStrategy = TimeBaseRotation{24 * time.Hour}
		}
	})
}

// Close closes a current log file.
func (l *CustomLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.close()
}

func (l *CustomLogger) close() {
	if l.file != nil {
		l.file.Close()
	}
	l.file = nil
	l.createdAt = time.Time{}
}

// removableFiles returns information of the files that
// are no longer used by the logger.
func (l *CustomLogger) removableFiles() []FileInfo {
	files, err := ioutil.ReadDir(l.Dir)
	if err != nil {
		l.Errorf(context.Background(), "failed to read dir: %s: %s", l.Dir, err)
		return nil
	}

	re := regexp.MustCompile("app_([0-9]+).log")

	fis := make([]FileInfo, 0, len(files))
	for _, file := range files {
		if l.file != nil && file.Name() == filepath.Base(l.file.Name()) {
			continue
		}
		ms := re.FindStringSubmatch(file.Name())
		if len(ms) == 0 {
			continue
		}
		t, err := time.ParseInLocation(timeFormat, ms[1], time.UTC)
		if err != nil {
			l.Errorf(context.Background(), "unexpected file name: %s", file.Name())
			return nil
		}
		fis = append(fis, FileInfo{
			l.Dir,
			file.Name(),
			file.Size(),
			file.ModTime(),
			t,
		})
	}
	return fis
}

func (l *CustomLogger) rotate() error {
	now := time.Now()
	path := filepath.Join(l.Dir, fmt.Sprintf("app_%s.log", now.In(time.UTC).Format(timeFormat)))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	l.close()
	l.file = f
	l.createdAt = now
	return nil
}

type logPayload struct {
	Timestamp struct {
		Seconds int64 `json:"seconds"`
		Nanos   int   `json:"nanos"`
	} `json:"timestamp"`
	Severity LogLevel `json:"severity"`
	Message  string   `json:"message"`
}

// FileInfo is a information of a log file.
type FileInfo struct {
	// Dir is a directory which the file is exist.
	Dir string
	// Name is a base name.
	Name string
	// Size is a length of bytes.
	Size int64
	// UpdatedAt is a time the file was last updated.
	UpdatedAt time.Time
	// CreatedAt is a time the file was created.
	CreatedAt time.Time
}

// Path returns a path to a log file.
func (fi FileInfo) Path() string {
	return filepath.Join(fi.Dir, fi.Name)
}
