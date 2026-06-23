.PHONY: test test-onnx longmemeval-smoke longmemeval-recall-gate download-dataset \
	longmemeval-bench longmemeval-bench-full longmemeval-qa-generate longmemeval-judge longmemeval-full-eval

test:
	go test ./...

test-onnx:
	go test -count=1 -run 'ONNX|IngestRetrieve|RecallGate|Bench' ./...

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