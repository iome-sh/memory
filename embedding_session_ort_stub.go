//go:build !cgo || (!ORT && !ALL)

package memory

import (
	"context"
	"fmt"

	"github.com/knights-analytics/hugot"
)

func newORTSession(ctx context.Context, cfg HugotBackendConfig) (*hugot.Session, error) {
	_ = ctx
	_ = cfg
	return nil, fmt.Errorf("ORT backend requires CGO_ENABLED=1 and go build -tags ORT (set %s=go for pure-Go GoMLX)", EnvHugotBackend)
}
