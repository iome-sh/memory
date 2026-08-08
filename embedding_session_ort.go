//go:build cgo && (ORT || ALL)

package memory

import (
	"context"
	"fmt"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/options"
)

func newORTSession(ctx context.Context, cfg HugotBackendConfig) (*hugot.Session, error) {
	opts := make([]options.WithOption, 0, 3)
	if cfg.LibraryDir != "" {
		opts = append(opts, options.WithOnnxLibraryPath(cfg.LibraryDir))
	}
	if cfg.Cuda {
		opts = append(opts, options.WithCuda(map[string]string{
			"device_id": cfg.DeviceID,
		}))
	}
	if cfg.CoreML {
		opts = append(opts, options.WithCoreML(map[string]string{}))
	}
	session, err := hugot.NewORTSession(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create hugot ORT session: %w", err)
	}
	return session, nil
}
