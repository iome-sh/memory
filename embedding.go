package memory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

const (
	// MiniLMEmbeddingDim is the output size for all-MiniLM-L6-v2 (and KnightsAnalytics ONNX export).
	MiniLMEmbeddingDim = 384
	// DefaultHashEmbeddingDim is the dimension used by GenerateSimpleEmbedding when dim <= 0.
	DefaultHashEmbeddingDim = 768
	// EnvONNXModelPath points at a hugot model directory (tokenizer + model.onnx) or a single .onnx file.
	EnvONNXModelPath = "MEMORY_ONNX_MODEL_PATH"
	// EnvEmbeddingStrict disables hash fallback when ONNX inference fails (recommended in production).
	EnvEmbeddingStrict = "MEMORY_EMBEDDING_STRICT"
)

// GONNXOptions configures a pure-Go ONNX embedder (hugot backend, no ORT dylibs).
type GONNXOptions struct {
	ModelPath string
	// Strict disables silent fallback to GenerateSimpleEmbedding on inference errors.
	Strict bool
}

// GONNXEmbedder runs sentence embeddings via hugot FeatureExtractionPipeline.
// Call Close when done to release the hugot session.
type GONNXEmbedder struct {
	session  *hugot.Session
	pipeline *pipelines.FeatureExtractionPipeline
	ctx      context.Context
	dim      int
	strict   bool

	fallbackOnce sync.Once
}

// NewGONNXEmbedder loads a hugot-compatible ONNX model directory and starts a feature-extraction pipeline.
func NewGONNXEmbedder(opts GONNXOptions) (*GONNXEmbedder, error) {
	dir, onnxFile, err := resolveONNXModelDir(opts.ModelPath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	session, err := hugot.NewGoSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("create hugot session: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		ModelPath:    dir,
		Name:         "memory-embedding",
		OnnxFilename: onnxFile,
		Options: []hugot.FeatureExtractionOption{
			pipelines.WithNormalization(),
		},
	}

	pipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		_ = session.Destroy()
		return nil, fmt.Errorf("create hugot feature extraction pipeline: %w", err)
	}

	dim := MiniLMEmbeddingDim
	if probe, probeErr := pipeline.RunPipeline(ctx, []string{"dimension probe"}); probeErr == nil && len(probe.Embeddings) > 0 && len(probe.Embeddings[0]) > 0 {
		dim = len(probe.Embeddings[0])
	}

	return &GONNXEmbedder{
		session:  session,
		pipeline: pipeline,
		ctx:      ctx,
		dim:      dim,
		strict:   opts.Strict,
	}, nil
}

// Dimension returns the native embedding width produced by the loaded ONNX model.
func (e *GONNXEmbedder) Dimension() int {
	if e == nil || e.dim <= 0 {
		return MiniLMEmbeddingDim
	}
	return e.dim
}

// Close releases hugot session resources.
func (e *GONNXEmbedder) Close() error {
	if e == nil || e.session == nil {
		return nil
	}
	return e.session.Destroy()
}

// Embed returns an L2-normalized embedding for text.
func (e *GONNXEmbedder) Embed(text string) ([]float32, error) {
	if e == nil || e.pipeline == nil {
		return nil, fmt.Errorf("onnx embedder not initialized")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return make([]float32, e.Dimension()), nil
	}

	out, err := e.pipeline.RunPipeline(e.ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("onnx embed: %w", err)
	}
	if out == nil || len(out.Embeddings) == 0 || len(out.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("onnx embed: empty embedding output")
	}
	vec := make([]float32, len(out.Embeddings[0]))
	copy(vec, out.Embeddings[0])
	return vec, nil
}

// Func returns an EmbeddingFunc that preserves the PalaceConfig injectable signature.
// The returned closure keeps the embedder alive; callers should not discard the embedder
// until all embedding work is finished (or use Func() only via NewGONNXEmbeddingFunc).
func (e *GONNXEmbedder) Func() EmbeddingFunc {
	return func(text string, dim int) []float32 {
		vec, err := e.Embed(text)
		if err != nil {
			if e.strict {
				slog.Error("onnx embedding failed in strict mode", "err", err)
				return fitEmbeddingDim(make([]float32, e.Dimension()), dim)
			}
			e.fallbackOnce.Do(func() {
				slog.Warn("onnx embedding failed; falling back to hash embeddings for this process", "err", err)
			})
			return GenerateSimpleEmbedding(text, dim)
		}
		return fitEmbeddingDim(vec, dim)
	}
}

// NewGONNXEmbeddingFunc returns a Palace-compatible EmbeddingFunc backed by pure-Go ONNX.
// modelPath may be empty to use GenerateSimpleEmbedding (dev/tests).
// modelPath may be a hugot model directory or a direct path to a single .onnx file.
func NewGONNXEmbeddingFunc(modelPath string) (EmbeddingFunc, error) {
	if strings.TrimSpace(modelPath) == "" {
		return GenerateSimpleEmbedding, nil
	}
	emb, err := NewGONNXEmbedder(GONNXOptions{
		ModelPath: modelPath,
		Strict:    envBool(EnvEmbeddingStrict),
	})
	if err != nil {
		return nil, err
	}
	return emb.Func(), nil
}

// NewGONNXEmbeddingFuncFromEnv loads MEMORY_ONNX_MODEL_PATH when set; otherwise hash embeddings.
func NewGONNXEmbeddingFuncFromEnv() (EmbeddingFunc, error) {
	return NewGONNXEmbeddingFunc(os.Getenv(EnvONNXModelPath))
}

// resolveONNXModelDir accepts a hugot model directory or a lone .onnx file path.
func resolveONNXModelDir(modelPath string) (dir string, onnxFile string, err error) {
	modelPath = strings.TrimSpace(modelPath)
	if modelPath == "" {
		return "", "", fmt.Errorf("onnx model path is required")
	}
	info, err := os.Stat(modelPath)
	if err != nil {
		return "", "", fmt.Errorf("onnx model not found at %s: %w", modelPath, err)
	}
	if info.IsDir() {
		return modelPath, "", nil
	}
	if strings.EqualFold(filepath.Ext(modelPath), ".onnx") {
		return filepath.Dir(modelPath), filepath.Base(modelPath), nil
	}
	return "", "", fmt.Errorf("onnx model path must be a directory or .onnx file: %s", modelPath)
}

// fitEmbeddingDim aligns an embedding vector with the requested Palace/Qdrant dimension.
// When dim <= 0 the native model dimension is returned. Mismatched dims truncate or zero-pad;
// production deployments should set Qdrant collection size == model dimension (384 for MiniLM).
func fitEmbeddingDim(vec []float32, dim int) []float32 {
	if dim <= 0 {
		return vec
	}
	if len(vec) == dim {
		out := make([]float32, dim)
		copy(out, vec)
		return out
	}
	if len(vec) > dim {
		return append([]float32(nil), vec[:dim]...)
	}
	out := make([]float32, dim)
	copy(out, vec)
	return out
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}