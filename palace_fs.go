package memory

import (
	"fmt"
	"os"
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
// The temp file is chmod 0600 before rename. Caller serializes shared
// dest paths (writeMu) when dest is entity-graph.json or event-time.json.
func palaceWriteFileAtomic(dir, dest, tmpPattern string, data []byte) error {
	if err := palaceMkdirAll(dir); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return fmt.Errorf("create temp failed: %w", err)
	}
	tmpName := tmpFile.Name()
	if err := tmpFile.Chmod(palaceFileMode); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("chmod temp failed: %w", err)
	}
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp failed: %w", err)
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
