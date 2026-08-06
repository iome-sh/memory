.PHONY: test test-onnx test-ort download-ort-deps build-ort-bench longmemeval-smoke longmemeval-recall-gate download-dataset \
	longmemeval-bench longmemeval-bench-full longmemeval-qa-generate longmemeval-judge longmemeval-full-eval \
	residual-gate advanced-agent-inventory-residual-gate k2-event-time-index-residual-gate

test:
	go test ./...

# Offline residual honesty pins (s1297 inventory + s1303 K2 event-time index).
# Soft skip: SKIP_ADVANCED_AGENT_INVENTORY=1 · SKIP_K2_EVENT_TIME_INDEX=1
residual-gate: advanced-agent-inventory-residual-gate k2-event-time-index-residual-gate

# Offline residual honesty pin s1297 — kernel advanced agent inventory (not product Memory GA).
advanced-agent-inventory-residual-gate:
	bash scripts/advanced_agent_inventory_residual_gate.sh

# Offline residual honesty pin s1303 — K2 ListMemoryWithOptions / event-time index residual
# (filters before limit shipped · full event-time index residual · not invent index green / Memory GA).
k2-event-time-index-residual-gate:
	bash scripts/k2_event_time_index_residual_gate.sh

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

longmemeval-bench-full:
	LONGMEMEVAL_DATASET=data/longmemeval_oracle.json bash scripts/longmemeval_recall_bench.sh

longmemeval-qa-generate:
	python3 scripts/longmemeval_qa_generate.py \
		--dataset $${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json} \
		--output $${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl} \
		--workers $${LONGMEMEVAL_QA_WORKERS:-4} \
		$$(if [ -n "$${LONGMEMEVAL_QA_LIMIT:-}" ] && [ "$${LONGMEMEVAL_QA_LIMIT}" != "0" ]; then echo --limit $${LONGMEMEVAL_QA_LIMIT}; fi)

longmemeval-judge:
	bash scripts/longmemeval_judge.sh \
		$${LONGMEMEVAL_JUDGE_MODEL:-gpt-4o-mini} \
		$${LONGMEMEVAL_HYPOTHESES:-hypotheses.jsonl} \
		$${LONGMEMEVAL_DATASET:-data/longmemeval_oracle.json}

longmemeval-full-eval:
	bash scripts/longmemeval_full_eval.sh