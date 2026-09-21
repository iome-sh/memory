package memory

import (
	"fmt"
	"os"
	"path/filepath"
)

// Palace directory and file modes for a local tenant palace (Memory P0 / #85).
// Dirs 0700, files 0600. Not encryption at rest. Not multi-tenant isolation.
const (
	palaceDirMode  os.FileMode = 0o700
	palaceFileMode os.FileMode = 0o600
)

func palaceMkdirAll(path string) error {
	return os.MkdirAll(path, palaceDirMode)
}

func palaceWriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, palaceFileMode)
}

// palaceWriteFileAtomic writes data to dest via temp+rename under dir.
// The temp file is chmod 0600 and fsynced before rename. Caller serializes
// shared dest paths (writeMu) when dest is entity-graph.json or event-time.json.
func palaceWriteFileAtomic(dir, dest, tmpPattern string, data []byte) error {
	if err := palaceMkdirAll(dir); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return fmt.Errorf("create temp failed: %w", err)
	}
	tmpName := tmpFile.Name()
	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpName)
	}
	if err := tmpFile.Chmod(palaceFileMode); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp failed: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp failed: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("fsync temp failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp failed: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename failed: %w", err)
	}
	return nil
}

// palaceWriteFileNamed writes data to path (0600) and fsyncs before close.
// Used for WAL pending records and IngestTurn sidecars (*.tmp-<turnID>).
func palaceWriteFileNamed(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := palaceMkdirAll(dir); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, palaceFileMode)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	cleanup := func() {
		f.Close()
		os.Remove(path)
	}
	if err := f.Chmod(palaceFileMode); err != nil {
		cleanup()
		return fmt.Errorf("chmod failed: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write failed: %w", err)
	}
	if err := f.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("fsync failed: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return fmt.Errorf("close failed: %w", err)
	}
	return nil
}

func palaceFsyncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = f.Sync()
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
