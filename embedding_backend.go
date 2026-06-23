package memory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/knights-analytics/hugot"
)

const (
	// EnvHugotBackend selects the hugot inference backend: go (default), ort, or auto.
	EnvHugotBackend = "MEMORY_HUGOT_BACKEND"
	// EnvORTLibraryDir is the directory containing libonnxruntime.so/.dylib (ORT builds only).
	EnvORTLibraryDir = "MEMORY_ORT_LIBRARY_DIR"
	// EnvORTCuda enables the CUDA execution provider when set to 1/true.
	EnvORTCuda = "MEMORY_ORT_CUDA"
	// EnvORTCoreML enables the CoreML execution provider when set to 1/true.
	EnvORTCoreML = "MEMORY_ORT_COREML"
	// EnvORTCudaDeviceID selects the CUDA device (default 0).
	EnvORTCudaDeviceID = "MEMORY_ORT_CUDA_DEVICE_ID"
)

// HugotBackendConfig holds resolved hugot session options from the environment.
type HugotBackendConfig struct {
	Backend    string // go, ort, auto
	LibraryDir string
	Cuda       bool
	CoreML     bool
	DeviceID   string
}

// ResolveHugotBackendFromEnv reads MEMORY_HUGOT_BACKEND and ORT accelerator env vars.
func ResolveHugotBackendFromEnv() HugotBackendConfig {
	cfg := HugotBackendConfig{
		Backend:  strings.ToLower(strings.TrimSpace(os.Getenv(EnvHugotBackend))),
		DeviceID: strings.TrimSpace(os.Getenv(EnvORTCudaDeviceID)),
	}
	if cfg.Backend == "" {
		cfg.Backend = "go"
	}
	if cfg.DeviceID == "" {
		cfg.DeviceID = "0"
	}
	cfg.LibraryDir = strings.TrimSpace(os.Getenv(EnvORTLibraryDir))
	cfg.Cuda = envBool(EnvORTCuda)
	cfg.CoreML = envBool(EnvORTCoreML)

	// auto: prefer ORT with platform-appropriate accelerator when ORT is available.
	if cfg.Backend == "auto" {
		switch runtime.GOOS {
		case "darwin":
			if !cfg.Cuda {
				cfg.CoreML = true
			}
		default:
			// Linux: CUDA only when MEMORY_ORT_CUDA is set (GPU compose image).
			cfg.Cuda = envBool(EnvORTCuda)
		}
	}

	return cfg
}

// HugotBackendLabel returns a log-friendly backend description.
func (c HugotBackendConfig) HugotBackendLabel() string {
	switch c.Backend {
	case "go":
		return "gomlx"
	case "ort", "auto":
		switch {
		case c.Cuda:
			return "ort+cuda:" + c.DeviceID
		case c.CoreML:
			return "ort+coreml"
		default:
			return "ort+cpu"
		}
	default:
		return c.Backend
	}
}

func newHugotSession(ctx context.Context, cfg HugotBackendConfig) (*hugot.Session, string, error) {
	switch cfg.Backend {
	case "go":
		s, err := hugot.NewGoSession(ctx)
		return s, "gomlx", err
	case "ort":
		s, err := newORTSession(ctx, cfg)
		return s, cfg.HugotBackendLabel(), err
	case "auto":
		if s, label, err := tryORTSession(ctx, cfg); err == nil {
			return s, label, nil
		} else {
			slog.Debug("hugot auto: ORT unavailable, falling back to gomlx", "err", err)
		}
		s, err := hugot.NewGoSession(ctx)
		return s, "gomlx", err
	default:
		return nil, "", fmt.Errorf("unsupported %s=%q (want go, ort, or auto)", EnvHugotBackend, cfg.Backend)
	}
}

func tryORTSession(ctx context.Context, cfg HugotBackendConfig) (*hugot.Session, string, error) {
	ortCfg := cfg
	ortCfg.Backend = "ort"
	s, err := newORTSession(ctx, ortCfg)
	if err != nil {
		return nil, "", err
	}
	return s, ortCfg.HugotBackendLabel(), nil
}