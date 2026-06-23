.PHONY: test test-onnx longmemeval-smoke longmemeval-recall-gate download-dataset longmemeval-bench longmemeval-bench-full

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