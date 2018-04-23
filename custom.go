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
	dir               string
	onUnexpectedError func(err error, level LogLevel, message string, args ...interface{})
	rotationStrategy  RotationStrategy

	mu        sync.Mutex
	file      *os.File
	createdAt time.Time
}

// NewCustomLogger create a new logger for custom runtime with given options.
func NewCustomLogger(opts ...CustomLoggerOption) *CustomLogger {
	l := &CustomLogger{
		dir:               "/var/log/app_engine/",
		onUnexpectedError: handleError,
		rotationStrategy:  TimeBaseRotation{24 * time.Hour},
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

func handleError(origErr error, level LogLevel, message string, args ...interface{}) {
	const format = "unexpected error: %s: %s: %s\n"

	fmt.Fprintf(os.Stderr, format, origErr, level, fmt.Sprintf(message, args...))

	f, err := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open error.log: %s", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, format, origErr, level, fmt.Sprintf(message, args...))
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

	if err := l.tryRotate(); err != nil {
		l.recoverError(err, level, format, args)
		return
	}

	if err := json.NewEncoder(l.file).Encode(&payload); err != nil {
		l.recoverError(err, level, format, args)
		return
	}
}

func (l *CustomLogger) recoverError(err error, level LogLevel, format string, args ...interface{}) {
	if l.onUnexpectedError != nil {
		l.onUnexpectedError(err, level, format, args...)
	}
}

func (l *CustomLogger) tryRotate() error {
	if l.file != nil {
		stat, err := l.file.Stat()
		if err != nil {
			return err
		}
		if l.rotationStrategy.ShouldRotate(FileInfo{l.dir, stat.Name(), stat.Size(), stat.ModTime(), l.createdAt}) {
			if err := l.rotate(); err != nil {
				return err
			}
		}
	} else {
		if err := l.rotate(); err != nil {
			return err
		}
	}
	return nil
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
	files, err := ioutil.ReadDir(l.dir)
	if err != nil {
		l.Errorf(context.Background(), "failed to read dir: %s: %s", l.dir, err)
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
			l.dir,
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
	path := filepath.Join(l.dir, fmt.Sprintf("app_%s.log", now.In(time.UTC).Format(timeFormat)))
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
	// dir is a directory which the file is exist.
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

// CustomLoggerOption is a option of the CustomLogger.
type CustomLoggerOption func(l *CustomLogger)

// OutputTo specifies a log file directory.
func OutputTo(dir string) CustomLoggerOption {
	return func(l *CustomLogger) {
		l.dir = dir
	}
}

// OnUnexpectedError specifies a error handler that
// is called when the logger could not output the log.
func OnUnexpectedError(f func(err error, level LogLevel, format string, args ...interface{})) CustomLoggerOption {
	return func(l *CustomLogger) {
		l.onUnexpectedError = f
	}
}

// RotatedBy specifies a rotation strategy.
func RotatedBy(s RotationStrategy) CustomLoggerOption {
	return func(l *CustomLogger) {
		l.rotationStrategy = s
	}
}
