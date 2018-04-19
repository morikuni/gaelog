package gaelog

import (
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
		LeaveLatest int
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
			ll := LeaveLatest{latest}
			ca := CleanUpAll{}

			var fis []FileInfo
			for _, t := range test.Input.Times {
				fis = append(fis, FileInfo{CreatedAt: t})
			}

			result, err := ll.Apply(fis)
			assert.NoError(t, err)
			assert.Len(t, result, test.Expect.LeaveLatest)

			result, err = ca.Apply(fis)
			assert.NoError(t, err)
			assert.Len(t, result, len(fis))
		})
	}
}
