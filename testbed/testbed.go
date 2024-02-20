package testbed

import (
	"context"
	"github.com/antonydenyer/block-builder-mempool/app"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewRequest(method, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(method, target, body)
}

func StartApp(t *testing.T) (context.Context, *app.App) {
	ctx, app, err := app.Start(context.TODO(), "test", "test")
	require.NoError(t, err)
	return ctx, app
}
