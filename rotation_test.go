package gaelog

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRotationStrategy(t *testing.T) {
	type Input struct {
		FileInfo FileInfo
	}
	type Expect struct {
		TimeBaseRotation bool
	}
	type Test struct {
		Input  Input
		Expect Expect
	}

	now := time.Now()
	maxAge := time.Hour
	tests := []Test{
		{
			Input{
				FileInfo{
					CreatedAt: now.Add(-maxAge + time.Minute),
				},
			},
			Expect{
				TimeBaseRotation: false,
			},
		},
		{
			Input{
				FileInfo{
					CreatedAt: now.Add(-maxAge - time.Minute),
				},
			},
			Expect{
				TimeBaseRotation: true,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tbr := RotateEvery{maxAge}
			nr := NeverRotate{}

			assert.Equal(t, test.Expect.TimeBaseRotation, tbr.ShouldRotate(test.Input.FileInfo))
			assert.False(t, nr.ShouldRotate(test.Input.FileInfo))
		})
	}
}
