# AC-ACL-01 tip dig — membership→palace and namespace isolation

This note records what the public tree of `iome-sh/memory` shows at the commit below. It does not change library behavior.

| Field | Value |
| --- | --- |
| Date | 2026-09-24 PT (PDT) |
| Issue | [#175](https://github.com/iome-sh/memory/issues/175) |
| Tip SHA investigated | `70cba458cc6b7637a9127c9179fc8b5096c6efec` (`origin/main`, 2026-09-23) |

## Q1 membership→palace

**Q1 membership→palace: UNKNOWN**

This tip does not define what membership means for a Cloud Memory palace. No type or function binds a human, role, or workspace principal to a palace.

Search covered Go sources, tests, and in-repo docs for: `membership`, `member`, `palace`, `workspace`, `principal`, `tenant`, `entitlement`, `invite`, `grant`, `role`, `ACL`, `RBAC`, `namespace`, `folder`, `isolation`, `writer`, `single-writer`, `default-deny`, `default deny`, `omit-from-list`, `omit`.

The module is a local filesystem library. Hosted palace membership sits outside that surface, so the question stays UNKNOWN.

Evidence:

- `PalaceStore`, `PalaceConfig`, `NewPalaceStore`, and `NewPalaceStoreWithConfig` in `memory.go` take a filesystem `BaseDir`. They have no member, principal, or entitlement field.
- `DefaultPalaceBaseDir` in `memory.go` is `.palace` when `BaseDir` is empty. The comment says it is not a hosted palace.
- `SECURITY.md` describes a Go library, not a multi-tenant cloud service. Hosted Palace stays sunset. There is no cloud Memory SLA. The library does not implement mesh `X-IOMesh-Org`.
- `docs/OPEN_SOURCE_AUDIT.md` lists multitenant hosted Palace, a cloud Memory SLA, and mesh `X-IOMesh-Org` as out of scope for this package.
- The word membership appears once, on `entryEntityKeySet` in `multihop.go`: a set of entity keys for graph expansion.
- “Example workspace” in `examples/dept-rca/support/` is fixture prose. It is not a workspace principal type.

## Q2 namespace isolation without multi-writer

**Q2 namespace isolation without multi-writer: UNKNOWN**

This tip does not show isolated namespaces or folders inside one palace, and it does not show one writer per namespace. There is no namespace API.

Directories, tags, and a one-process-per-root writer contract are present. None of them is defined as namespace or folder isolation, so Q2 stays UNKNOWN.

Evidence:

- There is no `Namespace` type, folder grant, default-deny check, or list path that omits unauthorized names. `ListMemoryOptions` in `list_memory.go` applies caller-supplied filters (`SessionID`, `SessionIDs`, `Tag`, `TagPrefix`, `Query`, tier, time). An empty `Tag` does not hide entries.
- `PalaceStore.ensureDirs` and `getTierDir` in `memory.go` create tier directories under one `BaseDir`: `tier-0-subconscious`, `tier-1-working`, `tier-2-contextual`, `tier-3-archival`, `tier-4-semantic`, plus `versions`, `relations`, `indexes`, and WAL pending. Those are storage tiers. `palaceDirMode` and `palaceFileMode` in `palace_fs.go` are `0700` and `0600`. The file comment says this is not multi-tenant isolation.
- `SECURITY.md`: path layout is not a secure tenancy boundary. Use separate roots (or OS isolation) per trust domain. A shared palace root is not multi-tenant security. The supported topology is one process per palace root.
- `README.md` Topology: one process per palace root. Isolation is the directory passed as `BaseDir`. This library does not implement mesh `X-IOMesh-Org`.
- `PalaceStore.writeMu` in `memory.go` serializes in-process rewrites of `relations/entity-graph.json` and `indexes/event-time.json`. `PalaceConfig.TransactionalIngest` is an intent log for one writer (`wal/pending`), not a multi-process lock (`SECURITY.md`, `memory.go`).
- `internal/writerprobe` and `cmd/two-process-writer-probe` record last-write-wins when two OS processes share one root. The package comment says multi-process writers remain unsupported: not a lock, not flock, not tenancy.
- Tag `dept:{id}` is an optional exact match via `EntryHasTag`. `docs/TTFH.md` and `TestIngestTurn_TagFilterSearchAndFactsAsOf` in `ingest_turn_test.go` place several department tags in one palace. An empty `Tag` returns both. The kernel has no org IDs. The tag is not a kernel org filter.

One writer process per palace root is documented in the `README.md` Topology section, in `SECURITY.md` (single-writer contract), and on `PalaceStore.writeMu`. That contract is per root. It does not define namespace or folder isolation.

## Multi-human R/W

**Multi-human R/W: Gap** (not Exists).

No API grants a second human read or write on a shared palace.

## What would resolve UNKNOWN

- **Q1:** An API in this module that names a membership subject and binds that subject to one palace. `SECURITY.md` and `docs/OPEN_SOURCE_AUDIT.md` state that a hosted palace and the mesh org header are not implemented here.
- **Q2:** A contract that names a namespace or folder as a default-deny isolation boundary with one writer, and that omits unauthorized names from list results. Tier directories (`ensureDirs`, `getTierDir`) and optional `Tag` filters (`ListMemoryOptions`, `EntryHasTag`) do not do that.

## Search log

| Pattern | Result on tip `70cba458cc6b7637a9127c9179fc8b5096c6efec` |
| --- | --- |
| `membership` | `entryEntityKeySet` only (`multihop.go`) |
| `namespace` / `Namespace` | no type or API |
| `ACL` / `RBAC` / `default-deny` / `default deny` / `omit-from-list` | no access-control API (unrelated “omit” hits are temporal evidence formatting and JSON `omitempty`) |
| `principal` / `entitlement` / `invite` | no matches as an access subject |
| `workspace` | fixture prose in `examples/dept-rca/support/`; not a principal |
| `writer` / `single-writer` | one process per palace root; multi-process writers unsupported |
| `tenant` | `palace_fs.go` “local tenant palace” comment plus “Not multi-tenant isolation”; `SECURITY.md` says not multi-tenant |
