package gaelog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanupStrategy(t *testing.T) {
	type Input struct {
		Times []time.Time
	}
	type Expect struct {
		KeepLatest int
	}
	type Test struct {
		Input  Input
		Expect Expect
	}

	now := time.Now()
	latest := 3 * time.Hour
	tests := []Test{
		{
			Input{
				[]time.Time{
					now.Add(-1 * time.Hour),
					now.Add(-2 * time.Hour),
					now.Add(-4 * time.Hour),
					now.Add(-5 * time.Hour),
				},
			},
			Expect{
				2,
			},
		},
		{
			Input{
				[]time.Time{
					now.Add(-4 * time.Hour),
					now.Add(-5 * time.Hour),
					now.Add(-6 * time.Hour),
				},
			},
			Expect{
				3,
			},
		},
		{
			Input{
				[]time.Time{
					now,
					now.Add(-1 * time.Hour),
					now.Add(-2 * time.Hour),
				},
			},
			Expect{
				0,
			},
		},
		{
			Input{
				[]time.Time{},
			},
			Expect{
				0,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ll := KeepLatest{latest}
			ca := CleanUpAll{}

			var fis []FileInfo
			for _, t := range test.Input.Times {
				fis = append(fis, FileInfo{CreatedAt: t})
			}

			result, err := ll.Apply(fis)
			assert.NoError(t, err)
			assert.Len(t, result, test.Expect.KeepLatest)

			result, err = ca.Apply(fis)
			assert.NoError(t, err)
			assert.Len(t, result, len(fis))
		})
	}
}

func TestScheduleCleanup(t *testing.T) {
	runCustom(t, "success", func(t *testing.T, l *CustomLogger) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go ScheduleCleanup(ctx, 2*time.Second, l, CleanUpAll{})
		l.rotate()
		time.Sleep(time.Second)
		l.rotate()
		assert.Len(t, l.removableFiles(), 1)
		time.Sleep(2 * time.Second)
		assert.Len(t, l.removableFiles(), 0)
		l.rotate()
		assert.Len(t, l.removableFiles(), 1)
		time.Sleep(2 * time.Second)
		assert.Len(t, l.removableFiles(), 0)
	})
}
