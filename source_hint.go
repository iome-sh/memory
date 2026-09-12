package memory

import (
	"strings"
)

// SourceHintPrivate is the default local-palace ingest class. It maps to the
// TUI ClassifyDigestSourceHint private bucket (private, palace, local, …).
const SourceHintPrivate = "private"

// SourceHintTagPrefix is the Content.Tags / TemporalTags prefix for an
// observable source class (source_hint:private, source_hint:mesh, …).
const SourceHintTagPrefix = "source_hint:"

const (
	ingestSourceMesh    = "mesh"
	ingestSourcePrivate = "private"
)

// FormatSourceHintTag returns source_hint:<hint> for a non-empty hint.
func FormatSourceHintTag(hint string) string {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return ""
	}
	return SourceHintTagPrefix + hint
}

// ClassifyIngestSourceHint maps a source hint or tag to mesh, private, or "".
// Aligned with TUI ClassifyDigestSourceHint cite-both buckets. Host process
// labels (mcp_memory_ingest_turn, source:iomesh-memory-mcp) are not a class.
// Catalog/grant/external stay host concerns and return "".
func ClassifyIngestSourceHint(hint string) string {
	h := normalizeIngestSourceHint(hint)
	if h == "" {
		return ""
	}
	switch h {
	case "mesh", "mesh_stream", "mesh_consume", "mesh_pulse", "mesh_incident", "mesh_incidents",
		"broker", "stream", "consume", "ops_pulse":
		return ingestSourceMesh
	case "private", "private_overlay", "private_rca", "palace", "palace_timeline",
		"local", "local_palace", "rca", "overlay", "agent_brief", "voc_brief", "market_telling":
		return ingestSourcePrivate
	}
	switch {
	case strings.HasPrefix(h, "mesh_"):
		return ingestSourceMesh
	case strings.HasPrefix(h, "private_"), strings.HasPrefix(h, "palace_"):
		return ingestSourcePrivate
	default:
		return ""
	}
}

func normalizeIngestSourceHint(hint string) string {
	h := strings.ToLower(strings.TrimSpace(hint))
	h = strings.ReplaceAll(h, "-", "_")
	if h == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(h, "source_hint:"):
		h = strings.TrimSpace(h[len("source_hint:"):])
	case strings.HasPrefix(h, "source:"):
		h = strings.TrimSpace(h[len("source:"):])
	}
	return strings.ReplaceAll(h, "-", "_")
}

// ensurePrivateIngestSource stamps an observable private source class on a
// local-palace IngestTurn when the caller did not already supply mesh or
// private provenance. Mesh-class hints stay distinct.
func ensurePrivateIngestSource(e *MemoryEntry) {
	if e == nil {
		return
	}
	class, hint := entryIngestSourceClass(*e)
	switch class {
	case ingestSourceMesh:
		if strings.TrimSpace(e.Provenance.SourceHint) == "" {
			e.Provenance.SourceHint = hint
		}
		e.Content.Tags = appendUniqueTag(e.Content.Tags, FormatSourceHintTag(e.Provenance.SourceHint))
	case ingestSourcePrivate:
		if strings.TrimSpace(e.Provenance.SourceHint) == "" {
			e.Provenance.SourceHint = hint
		}
		e.Content.Tags = appendUniqueTag(e.Content.Tags, FormatSourceHintTag(e.Provenance.SourceHint))
	default:
		e.Provenance.SourceHint = SourceHintPrivate
		e.Content.Tags = appendUniqueTag(e.Content.Tags, FormatSourceHintTag(SourceHintPrivate))
	}
}

func entryIngestSourceClass(e MemoryEntry) (class, hint string) {
	candidates := make([]string, 0, 1+len(e.Content.Tags)+len(e.TemporalTags))
	if s := strings.TrimSpace(e.Provenance.SourceHint); s != "" {
		candidates = append(candidates, s)
	}
	candidates = append(candidates, e.Content.Tags...)
	candidates = append(candidates, e.TemporalTags...)
	for _, c := range candidates {
		if cl := ClassifyIngestSourceHint(c); cl == ingestSourceMesh {
			return cl, firstClassifiableHint(c)
		}
	}
	for _, c := range candidates {
		if cl := ClassifyIngestSourceHint(c); cl == ingestSourcePrivate {
			return cl, firstClassifiableHint(c)
		}
	}
	return "", ""
}

func firstClassifiableHint(raw string) string {
	h := normalizeIngestSourceHint(raw)
	if h == "" {
		return SourceHintPrivate
	}
	return h
}

func appendUniqueTag(tags []string, tag string) []string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tags
	}
	for _, t := range tags {
		if strings.TrimSpace(t) == tag {
			return tags
		}
	}
	return append(tags, tag)
}
