.PHONY: all build test test-race cover vet fmt fmt-check tidy vuln check ci \
	test-onnx test-ort download-ort-deps build-ort-bench longmemeval-smoke longmemeval-recall-gate download-dataset \
	longmemeval-bench longmemeval-v2-bench longmemeval-bench-full longmemeval-qa-generate longmemeval-judge longmemeval-full-eval \
	longmemeval-v1-card two-process-writer-probe \
	residual-gate advanced-agent-inventory-residual-gate k2-event-time-index-residual-gate recmem-compaction-residual-gate \
	public-flip-readiness-gate \
	clean

COVER ?= coverage.out

all: check build

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=$(COVER) -covermode=atomic
	go tool cover -func=$(COVER) | tail -20

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi

tidy:
	go mod tidy
	go mod verify

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

build:
	go build ./...
	go build -o bin/longmemeval-bench ./cmd/longmemeval-bench
	go build -o bin/longmemeval-server ./cmd/longmemeval-server
	go build -o bin/longmemeval-v2-bench ./cmd/longmemeval-v2-bench

check: fmt-check vet test

# Mirrors GitHub Actions required gate (fmt + vet + test + vuln + build).
# Race/cover are optional locally (CGO/Qdrant soft paths); CI may run race on pure packages later.
ci: fmt-check vet test vuln build

# Offline residual honesty pins (s1297 inventory + s1303 K2 event-time index + s1313 RecMem compaction).
# Soft skip: SKIP_ADVANCED_AGENT_INVENTORY=1 · SKIP_K2_EVENT_TIME_INDEX=1 · SKIP_RECMEM_COMPACTION=1
residual-gate: advanced-agent-inventory-residual-gate k2-event-time-index-residual-gate recmem-compaction-residual-gate

# Offline residual honesty pin s1297 — kernel advanced agent inventory (not product Memory GA).
advanced-agent-inventory-residual-gate:
	bash scripts/advanced_agent_inventory_residual_gate.sh

# Offline residual honesty pin s1303 — K2 ListMemoryWithOptions / event-time index residual
# (filters before limit shipped · full event-time index residual · not invent index green / Memory GA).
k2-event-time-index-residual-gate:
	bash scripts/k2_event_time_index_residual_gate.sh

# Offline residual honesty pin s1313 — RecMem / compaction residual
# (AutoRecMemCompaction shipped partial · PerformCompaction · CompactionConfig · trigger advisory · HITL TUI · not invent GA token-reduction).
recmem-compaction-residual-gate:
	bash scripts/recmem_compaction_residual_gate.sh

# Offline M4 public-flip readiness residual s1467 (flip complete; gate is not the flip).
# Soft skip: SKIP_PUBLIC_FLIP_READINESS=1
# Honesty: public MIT · residual PASS ≠ public flip · gate PASS ≠ product GA · kernel first · not Memory GA · dual_write OFF (host policy, not kernel flag) · private control-plane / broker stays private.
public-flip-readiness-gate:
	bash scripts/public_flip_readiness_gate.sh

test-onnx:
	go test -count=1 -run 'ONNX|IngestRetrieve|RecallGate|Bench|HugotBackend' ./...

download-ort-deps:
	bash scripts/download_ort_deps.sh

# Requires: make download-ort-deps (libtokenizers.a + libonnxruntime in testdata/ort-deps/lib)
build-ort-bench: download-ort-deps
	@eval "$$(./scripts/ort_cgo_env.sh)" && \
	go build -tags ORT -o bin/longmemeval-bench-ort ./cmd/longmemeval-bench

# Requires CGO_ENABLED=1, -tags ORT, libonnxruntime + libtokenizers.a on the host.
test-ort: download-ort-deps
	@eval "$$(./scripts/ort_cgo_env.sh)" && \
	go test -tags ORT -count=1 -run 'ONNX|HugotBackend' ./...

longmemeval-smoke:
	go test -count=1 -run IngestRetrieve ./cmd/longmemeval-server/...

longmemeval-recall-gate:
	go test -count=1 -run RecallGate ./cmd/longmemeval-server/...

download-dataset:
	bash scripts/download_longmemeval_dataset.sh

longmemeval-bench:
	bash scripts/longmemeval_recall_bench.sh

longmemeval-v2-bench:
	go test -count=1 ./internal/longmemeval/...
	go run ./cmd/longmemeval-v2-bench -data-root testdata/longmemeval_v2_subset -tier small

longmemeval-bench-full:
	LONGMEMEVAL_DATASET=data/longmemeval_oracle.json bash scripts/longmemeval_recall_bench.sh

longmemeval-qa-generate:
	python3 scripts/longmemeval_qa_generate.py \
	--dataset $${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json} \
	--output $${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl} \
	--workers $${LONGMEMEVAL_QA_WORKERS:-4} \
	$$(if [ -n "$${LONGMEMEVAL_QA_LIMIT:-}" ] && [ "$${LONGMEMEVAL_QA_LIMIT}" != "0" ]; then echo --limit $${LONGMEMEVAL_QA_LIMIT}; fi) \
	$$(if [ -n "$${LONGMEMEVAL_QA_SAMPLE:-}" ]; then echo --sample $${LONGMEMEVAL_QA_SAMPLE}; fi)

# Official V1 judge pin is gpt-4o-2024-08-06 (docs/LONGMEMEVAL.md).
# Default gpt-4o-mini is a cheap local path — not official V1.
# Upstream evaluate_qa.py zoo key gpt-4o maps to that pin (scripts/longmemeval_judge.sh).
longmemeval-judge:
	bash scripts/longmemeval_judge.sh \
	$${LONGMEMEVAL_JUDGE_MODEL:-gpt-4o-mini} \
	$${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl} \
	$${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json}

# Official V1 mixed-run methodology card. Optional; not part of make ci.
# Missing data/longmemeval_oracle.json → SKIP exit 0 (not a CI failure).
# testdata/longmemeval_oracle_subset.json is 3 single-session-user items — not mixed official V1.
# Official judge pin is gpt-4o-2024-08-06. Default LONGMEMEVAL_JUDGE_MODEL=gpt-4o-mini is a cheap local path — not official V1.
# LONGMEMEVAL_V1_RUN=1 also generates+judges a mixed sample (needs OPENAI_API_KEY + ONNX + running server).
# Local ONNX: BGE-small-en-v1.5 if present; else in-tree testdata/models/KnightsAnalytics_all-MiniLM-L6-v2
# (384-d). MiniLM is the local path when BGE is unavailable (HF download may 401) — not the official V1 embed pin.
longmemeval-v1-card:
	bash scripts/longmemeval_v1_card.sh

# Optional last-write-wins evidence. Not make ci / make test. Always exit 0.
# Multi-process writers remain unsupported. Flock is not shipped.
two-process-writer-probe:
	bash scripts/two_process_writer_probe.sh

longmemeval-full-eval:
	bash scripts/longmemeval_full_eval.sh

clean:
	rm -rf bin/ $(COVER) coverage.html
