# gaelog

[![CircleCI](https://circleci.com/gh/morikuni/gaelog/tree/master.svg?style=shield)](https://circleci.com/gh/morikuni/gaelog/tree/master)
[![GoDoc](https://godoc.org/github.com/morikuni/gaelog?status.svg)](https://godoc.org/github.com/morikuni/gaelog)
[![Go Report Card](https://goreportcard.com/badge/github.com/morikuni/gaelog)](https://goreportcard.com/report/github.com/morikuni/gaelog)
[![codecov](https://codecov.io/gh/morikuni/gaelog/branch/master/graph/badge.svg)](https://codecov.io/gh/morikuni/gaelog)

Logging library for GAE/Go.

## Example

```go
package main

import (
	"context"
	"time"

	"github.com/morikuni/gaelog"
)

func main() {
	ctx := context.Background()

	logger := &gaelog.CustomLogger{
		RotationStrategy:  gaelog.TimeBaseRotation{time.Hour}, // Rotate the log every hour.
		OnUnexpectedError: func(err error) { panic(err) },
	}

	// You can clean up the old logs, leaving latest 1 day logs.
	go gaelog.ScheduleCleanup(ctx, time.Hour, logger, gaelog.LeaveLatest{24 * time.Hour})

	logger.Infof(ctx, "hello %s", "world")
	logger.Printf(ctx, gaelog.Error, "log level error")
}
```