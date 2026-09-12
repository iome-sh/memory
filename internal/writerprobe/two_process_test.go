package writerprobe

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestTwoProcessWriterProbe starts two OS processes against one BaseDir.
// Two PalaceStore values in one process are not this test (in-process writeMu
// is already covered by TestPalaceStore_ConcurrentSharedFileWrites).
//
// Skipped unless MEMORY_TWO_PROCESS_PROBE=1 so default `go test ./...` stays
// green. Multi-process writers remain unsupported; flock is not shipped.
func TestTwoProcessWriterProbe(t *testing.T) {
	if child := os.Getenv("MEMORY_TWO_PROCESS_PROBE_CHILD"); child != "" {
		base := os.Getenv("MEMORY_TWO_PROCESS_PROBE_BASE")
		shared := os.Getenv("MEMORY_TWO_PROCESS_PROBE_SHARED")
		n := DefaultN
		if raw := os.Getenv("MEMORY_TWO_PROCESS_PROBE_N"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil {
				n = v
			}
		}
		if err := RunWriter(base, child, shared, n); err != nil {
			t.Fatalf("child writer %s: %v", child, err)
		}
		return
	}
	if os.Getenv("MEMORY_TWO_PROCESS_PROBE") != "1" {
		t.Skip("set MEMORY_TWO_PROCESS_PROBE=1 to collect two-process last-write-wins evidence; flock is not shipped")
	}

	base := t.TempDir()
	const shared = DefaultSharedID
	n := DefaultN
	if raw := os.Getenv("MEMORY_TWO_PROCESS_PROBE_N"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			n = v
		}
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	start := func(writer string) (*exec.Cmd, *bytes.Buffer) {
		cmd := exec.Command(exe, "-test.run=^TestTwoProcessWriterProbe$", "-test.count=1")
		cmd.Env = append(probeChildEnv(),
			"MEMORY_TWO_PROCESS_PROBE=1",
			"MEMORY_TWO_PROCESS_PROBE_CHILD="+writer,
			"MEMORY_TWO_PROCESS_PROBE_BASE="+base,
			"MEMORY_TWO_PROCESS_PROBE_SHARED="+shared,
			"MEMORY_TWO_PROCESS_PROBE_N="+strconv.Itoa(n),
		)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		return cmd, &buf
	}

	cmdA, bufA := start("A")
	cmdB, bufB := start("B")
	if err := cmdA.Start(); err != nil {
		t.Fatalf("start writer A: %v", err)
	}
	if err := cmdB.Start(); err != nil {
		_ = cmdA.Process.Kill()
		t.Fatalf("start writer B: %v", err)
	}
	errA := cmdA.Wait()
	errB := cmdB.Wait()
	if errA != nil {
		t.Logf("writer A: %v\n%s", errA, bufA.String())
	}
	if errB != nil {
		t.Logf("writer B: %v\n%s", errB, bufB.String())
	}

	r := Inspect(base, shared, n, []string{"A", "B"})
	var report bytes.Buffer
	WriteReport(&report, r)
	t.Logf("\n%s", report.String())

	if !r.SharedExists {
		t.Log("evidence: shared entry file missing")
	} else if !r.SharedValidJSON {
		t.Log("evidence: shared entry is not valid JSON")
	} else {
		t.Logf("last-write-wins winner=%s", r.SharedWinner)
	}
	if r.GraphExists && !r.GraphValidJSON {
		t.Log("evidence: entity-graph.json is not valid JSON")
	}
	if r.EventTimeExists && !r.EventTimeValidJSON {
		t.Log("evidence: event-time.json is not valid JSON")
	}
	t.Log("single-writer contract unchanged; probe ≠ flock; flock is not shipped")
}

func probeChildEnv() []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, e := range env {
		switch {
		case strings.HasPrefix(e, "MEMORY_TWO_PROCESS_PROBE="),
			strings.HasPrefix(e, "MEMORY_TWO_PROCESS_PROBE_CHILD="),
			strings.HasPrefix(e, "MEMORY_TWO_PROCESS_PROBE_BASE="),
			strings.HasPrefix(e, "MEMORY_TWO_PROCESS_PROBE_SHARED="),
			strings.HasPrefix(e, "MEMORY_TWO_PROCESS_PROBE_N="):
			continue
		default:
			out = append(out, e)
		}
	}
	return out
}
