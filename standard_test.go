package gaelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
)

func TestStandardLogger(t *testing.T) {
	ctx, close, err := aetest.NewContext()
	assert.NoError(t, err)
	defer close()

	l := StandardLogger{}

	l.Criticalf(ctx, "aaa")
	l.Errorf(ctx, "aaa")
	l.Warningf(ctx, "aaa")
	l.Infof(ctx, "aaa")
	l.Debugf(ctx, "aaa")
}
