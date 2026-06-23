.PHONY: test test-onnx longmemeval-smoke

test:
	go test ./...

test-onnx:
	go test -count=1 -run 'ONNX|IngestRetrieve' ./...

longmemeval-smoke:
	go test -count=1 -run IngestRetrieve ./cmd/longmemeval-server/...