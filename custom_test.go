package gaelog

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runCustom(t *testing.T, name string, f func(t *testing.T, l *CustomLogger)) {
	t.Run(name, func(t *testing.T) {
		l := NewCustomLogger(
			OutputTo("test"),
			OnUnexpectedError(func(err error, l LogLevel, f string, args ...interface{}) {
				t.Fatal(err, l, f, args)
			}),
			RotatedBy(NeverRotate{}),
		)
		defer CleanUp(l, CleanUpAll{})
		defer l.Close()

		f(t, l)
	})
}

func TestCustomLogger(t *testing.T) {
	runCustom(t, "Criticalf", func(t *testing.T, l *CustomLogger) {
		l.Criticalf(nil, "critical: %d", 1)
		r, err := ioutil.ReadFile(l.file.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(r), "CRITICAL")
		assert.Contains(t, string(r), "critical: 1")
	})

	runCustom(t, "Errorf", func(t *testing.T, l *CustomLogger) {
		l.Errorf(nil, "error: %d", 2)
		r, err := ioutil.ReadFile(l.file.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(r), "ERROR")
		assert.Contains(t, string(r), "error: 2")
	})

	runCustom(t, "Warningf", func(t *testing.T, l *CustomLogger) {
		l.Warningf(nil, "warning: %d", 3)
		r, err := ioutil.ReadFile(l.file.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(r), "WARNING")
		assert.Contains(t, string(r), "warning: 3")
	})

	runCustom(t, "Infof", func(t *testing.T, l *CustomLogger) {
		l.Infof(nil, "info: %d", 4)
		r, err := ioutil.ReadFile(l.file.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(r), "INFO")
		assert.Contains(t, string(r), "info: 4")
	})

	runCustom(t, "Debugf", func(t *testing.T, l *CustomLogger) {
		l.Debugf(nil, "debug: %d", 5)
		r, err := ioutil.ReadFile(l.file.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(r), "DEBUG")
		assert.Contains(t, string(r), "debug: 5")
	})

	runCustom(t, "RemovableFile", func(t *testing.T, l *CustomLogger) {
		l.rotationStrategy = TimeBaseRotation{time.Second}
		l.Debugf(nil, "hello world")
		assert.Len(t, l.removableFiles(), 0)
		name := filepath.Base(l.file.Name())

		time.Sleep(time.Second)
		l.Debugf(nil, "hello world") // rotate by this log
		rfs := l.removableFiles()
		if assert.Len(t, rfs, 1) {
			assert.Equal(t, l.dir, rfs[0].Dir)
			assert.Equal(t, name, rfs[0].Name)
		}
	})
}

func BenchmarkCustomLogger(b *testing.B) {
	l := NewCustomLogger(
		OutputTo("test"),
		OnUnexpectedError(func(err error, l LogLevel, f string, args ...interface{}) {
			b.Fatal(err, l, f, args)
		}),
		RotatedBy(NeverRotate{}),
	)
	defer CleanUp(l, CleanUpAll{})
	defer l.Close()

	for i := 0; i < b.N; i++ {
		l.Errorf(nil, "aaa")
	}
}
