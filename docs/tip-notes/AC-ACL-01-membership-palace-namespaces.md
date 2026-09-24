# AC-ACL-01 tip dig — membership→palace and namespace isolation

Investigation only. This note stamps tip truth in `iome-sh/memory`. It does not add ACL behavior.

| Field | Value |
| --- | --- |
| Date | 2026-09-24 PT (PDT) |
| Issue | [#175](https://github.com/iome-sh/memory/issues/175) |
| Tip SHA investigated | `70cba458cc6b7637a9127c9179fc8b5096c6efec` (`origin/main`, 2026-09-23) |

## Q1 membership→palace

**Q1 membership→palace: UNKNOWN**

No symbol in this tip defines what membership means for a Cloud Memory palace. There is no membership subject, grant, or API that binds a human, role, or workspace principal to a palace.

Searched (Go + Markdown + tests + public docs in-repo): `membership`, `member`, `palace`, `workspace`, `principal`, `tenant`, `entitlement`, `invite`, `grant`, `role`, `ACL`, `RBAC`, `namespace`, `folder`, `isolation`, `writer`, `single-writer`, `default-deny`, `default deny`, `omit-from-list`, `omit`.

Why inconclusive: the module is a local filesystem library. Cloud / hosted palace membership is outside its surface, and nothing here supplies a substitute definition. Stamping a meaning would invent RBAC.

Evidence that the surface is absent (not an Exists membership model):

- `PalaceStore` / `PalaceConfig` / `NewPalaceStore` / `NewPalaceStoreWithConfig` (`memory.go`) take a filesystem `BaseDir`. No member, principal, or entitlement field.
- `DefaultPalaceBaseDir` (`memory.go`) is the local root `.palace` when `BaseDir` is empty. Comment: not a hosted palace.
- `SECURITY.md` threat model: this module is a Go library, not a multi-tenant cloud service; hosted Palace stays sunset; not a cloud Memory SLA. It does not implement mesh `X-IOMesh-Org`.
- `docs/OPEN_SOURCE_AUDIT.md`: out of scope includes multitenant hosted Palace / cloud Memory SLA, and mesh `X-IOMesh-Org` (mesh clients, not this library).
- The only in-repo use of the word membership is `entryEntityKeySet` (`multihop.go`): a set of entity keys for graph expansion. That is not palace ACL.
- Strings “Example workspace” in `examples/dept-rca/support/` are fixture prose for a walking skeleton. They are not a workspace principal type.

## Q2 namespace isolation without multi-writer

**Q2 namespace isolation without multi-writer: UNKNOWN**

This tip does not show isolated namespaces or folders inside one palace with one writer per workspace. It also does not show a namespace API that could be read as that feature.

Why inconclusive: directories and tags exist, and a one-process-per-root writer contract exists, but neither is defined as namespace or folder isolation for a workspace. Treating either as Exists namespace ACL would invent that feature. Namespace chrome stays unproven.

What was found, and why it is not that feature:

- No `Namespace` type, folder grant, default-deny check, or omit-unauthorized-from-list path. `ListMemoryOptions` (`list_memory.go`) filters are caller-supplied (`SessionID`, `SessionIDs`, `Tag`, `TagPrefix`, `Query`, tier, time). Empty `Tag` does not hide entries.
- `PalaceStore.ensureDirs` and `getTierDir` (`memory.go`) create tier directories under one `BaseDir` (`tier-0-subconscious`, `tier-1-working`, `tier-2-contextual`, `tier-3-archival`, `tier-4-semantic`, plus `versions`, `relations`, `indexes`, WAL pending). Those are storage tiers. `palaceDirMode` / `palaceFileMode` (`palace_fs.go`) are `0700` / `0600`. File comment: not multi-tenant isolation.
- `SECURITY.md`: path layout is not a secure tenancy boundary; use separate roots (or OS isolation) per trust domain. Shared palace root is not multi-tenant security. One host process per palace root.
- `README.md` Topology: one process per palace root. Isolation is the directory passed as `BaseDir`. This library does not implement mesh `X-IOMesh-Org`.
- `PalaceStore.writeMu` (`memory.go`) serializes in-process rewrites of `relations/entity-graph.json` and `indexes/event-time.json`. `PalaceConfig.TransactionalIngest` is an intent log for one writer (`wal/pending`), not a multi-process lock (`SECURITY.md`, `memory.go`).
- `internal/writerprobe` and `cmd/two-process-writer-probe`: two OS processes on one root are last-write-wins evidence. Package comment: multi-process writers remain unsupported; not a lock, not flock, not tenancy.
- Tag `dept:{id}` is an optional exact `EntryHasTag` filter. `docs/TTFH.md` and `TestIngestTurn_TagFilterSearchAndFactsAsOf` (`ingest_turn_test.go`): several department tags can sit in one palace; empty `Tag` returns both; the kernel has no org IDs; the tag is not a kernel org filter.

Adjacent process contract (not the Q2 stamp): one writer process per palace root is documented at `README.md` Topology, `SECURITY.md` (single-writer contract), and `PalaceStore.writeMu`. That is per root, not per namespace or folder, and not a workspace-member writer. It is not stamped Exists for Q2.

## Multi-human R/W

**Multi-human R/W: Gap** (not Exists).

No API grants a second human read or write on a shared palace. Do not describe multi-human read/write as Exists.

## What would unblock UNKNOWN

- **Q1:** A cited API in this module that names the membership subject and how that subject binds to one palace. This tip has no such API. Hosted palace and mesh org header are explicitly not implemented here (`SECURITY.md`, `docs/OPEN_SOURCE_AUDIT.md`).
- **Q2:** A cited contract that names a namespace or folder as a default-deny isolation boundary, with one writer, and that omits unauthorized names from list results. Tier directories (`ensureDirs` / `getTierDir`) and optional `Tag` filters (`ListMemoryOptions`, `EntryHasTag`) do not meet that bar. Until that contract exists in tip, namespace chrome stays unproven.

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
