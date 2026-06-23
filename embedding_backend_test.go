package memory

import (
	"runtime"
	"testing"
)

func TestResolveHugotBackendFromEnv_DefaultGo(t *testing.T) {
	t.Setenv(EnvHugotBackend, "")
	t.Setenv(EnvORTCuda, "")
	t.Setenv(EnvORTCoreML, "")
	cfg := ResolveHugotBackendFromEnv()
	if cfg.Backend != "go" {
		t.Fatalf("backend = %q, want go", cfg.Backend)
	}
	if got := cfg.HugotBackendLabel(); got != "gomlx" {
		t.Fatalf("label = %q, want gomlx", got)
	}
}

func TestResolveHugotBackendFromEnv_ORTCuda(t *testing.T) {
	t.Setenv(EnvHugotBackend, "ort")
	t.Setenv(EnvORTCuda, "1")
	t.Setenv(EnvORTCudaDeviceID, "2")
	cfg := ResolveHugotBackendFromEnv()
	if !cfg.Cuda {
		t.Fatal("expected cuda enabled")
	}
	if got := cfg.HugotBackendLabel(); got != "ort+cuda:2" {
		t.Fatalf("label = %q, want ort+cuda:2", got)
	}
}

func TestResolveHugotBackendFromEnv_AutoDarwinCoreML(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only auto CoreML default")
	}
	t.Setenv(EnvHugotBackend, "auto")
	t.Setenv(EnvORTCuda, "")
	t.Setenv(EnvORTCoreML, "")
	cfg := ResolveHugotBackendFromEnv()
	if !cfg.CoreML {
		t.Fatal("expected coreml enabled on darwin auto")
	}
	if got := cfg.HugotBackendLabel(); got != "ort+coreml" {
		t.Fatalf("label = %q, want ort+coreml", got)
	}
}

func TestNewORTSession_StubWithoutBuildTag(t *testing.T) {
	t.Setenv(EnvHugotBackend, "ort")
	_, err := newORTSession(t.Context(), ResolveHugotBackendFromEnv())
	if err == nil {
		t.Fatal("expected error when ORT not linked")
	}
}