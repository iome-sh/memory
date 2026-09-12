// Two-process writer probe child: one PalaceStore process against a shared root.
//
// Multi-process writers remain unsupported. This binary is evidence collection
// (last-write-wins), not a lock and not flock.
//
//	go run ./cmd/two-process-writer-probe -base DIR -writer A
//	go run ./cmd/two-process-writer-probe -inspect -base DIR
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/iome-sh/memory/internal/writerprobe"
)

func main() {
	base := flag.String("base", "", "palace BaseDir (required)")
	writer := flag.String("writer", "", "writer id (A or B)")
	shared := flag.String("shared", writerprobe.DefaultSharedID, "colliding entry id")
	n := flag.Int("n", writerprobe.DefaultN, "unique writes and shared overwrites per writer")
	inspect := flag.Bool("inspect", false, "print last-write-wins / JSON validity and exit 0")
	flag.Parse()

	baseDir := strings.TrimSpace(*base)
	if baseDir == "" {
		fmt.Fprintln(os.Stderr, "two-process-writer-probe: -base is required")
		fmt.Fprintln(os.Stderr, "note: multi-process writers remain unsupported · flock is not shipped")
		os.Exit(2)
	}

	if *inspect {
		r := writerprobe.Inspect(baseDir, *shared, *n, []string{"A", "B"})
		writerprobe.WriteReport(os.Stdout, r)
		os.Exit(0)
	}

	w := strings.TrimSpace(*writer)
	if w == "" {
		fmt.Fprintln(os.Stderr, "two-process-writer-probe: -writer is required unless -inspect")
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "two-process-writer-probe: writer=%s pid=%d base=%s n=%d shared=%s\n", w, os.Getpid(), baseDir, *n, *shared)
	if err := writerprobe.RunWriter(baseDir, w, *shared, *n); err != nil {
		fmt.Fprintf(os.Stderr, "two-process-writer-probe: writer=%s: %v\n", w, err)
		os.Exit(1)
	}
}
