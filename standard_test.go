package gaelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
)

func TestStandardLogger(t *testing.T) {
	i, err := aetest.NewInstance(&aetest.Options{SuppressDevAppServerLog: true})
	assert.NoError(t, err)
	defer i.Close()
	req, err := i.NewRequest("", "", nil)
	assert.NoError(t, err)
	ctx := req.Context()

	l := NewStandardLogger()

	l.Criticalf(ctx, "aaa")
	l.Errorf(ctx, "aaa")
	l.Warningf(ctx, "aaa")
	l.Infof(ctx, "aaa")
	l.Debugf(ctx, "aaa")
}
