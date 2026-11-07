package jpoet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func (e *Environment) Eval(ctx context.Context, opts ...EvalOption) error {
	if err := e.lifecycle.Enter(); err != nil {
		return err
	}
	defer e.lifecycle.Leave()

	c := newEvalConfig()
	for _, opt := range opts {
		opt(&c)
	}
	err := e.flushBarrier.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer e.flushBarrier.Release(1)
	v, err := e.pool.acquire(ctx)
	if err != nil {
		return err
	}
	defer e.pool.release(v)
	v.refresh(e.generation)
	return c.run(ctx, v)
}

type EvalOption func(*evalConfig)

type evalConfig struct {
	nodeInput    *ast.Node
	snippetInput *snippetInput
	fileInput    *string

	writerOutput    io.Writer
	valueOutput     any
	directoryOutput string

	serializedFormat bool

	recordedInvocations *[]Invocation
	recordedPaths       *[]string
}

type snippetInput struct {
	filename string
	snippet  string
}

func newEvalConfig() evalConfig {
	return evalConfig{
		writerOutput:     os.Stdout,
		serializedFormat: true,
	}
}

func (c *evalConfig) run(ctx context.Context, v *pooledVM) error {
	if !c.hasInput() {
		return errors.New("missing input")
	}
	err := ctx.Err()
	if err != nil {
		return err
	}
	state := &evalState{ctx: ctx}
	v.slot.state = state
	serializedJson, err := evaluateInput(v.vm, c.nodeInput, c.snippetInput, c.fileInput)
	v.slot.state = nil
	if c.recordedInvocations != nil {
		*c.recordedInvocations = state.invocations
	}
	if c.recordedPaths != nil {
		*c.recordedPaths = state.paths
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return err
	}
	return writeOutput(serializedJson, c.writerOutput, c.valueOutput, c.directoryOutput, c.serializedFormat)
}

func (c *evalConfig) hasInput() bool {
	return c.nodeInput != nil || c.snippetInput != nil || c.fileInput != nil
}

func evaluateInput(vm *jsonnet.VM, nodeInput *ast.Node, snippetInput *snippetInput, fileInput *string) (string, error) {
	if nodeInput != nil {
		return vm.Evaluate(*nodeInput)
	}
	if snippetInput != nil {
		return vm.EvaluateAnonymousSnippet(snippetInput.filename, snippetInput.snippet)
	}
	return vm.EvaluateFile(*fileInput)
}

func writeOutput(serializedJson string, writerOutput io.Writer, valueOutput any, directoryOutput string, serializedFormat bool) error {
	if writerOutput != nil {
		output := serializedJson
		if !serializedFormat {
			err := json.Unmarshal([]byte(serializedJson), &output)
			if err != nil {
				return err
			}
		}
		_, err := writerOutput.Write([]byte(output))
		return err
	}
	if valueOutput != nil {
		if serializedFormat {
			return nil
		}
		return json.Unmarshal([]byte(serializedJson), valueOutput)
	}
	if directoryOutput != "" {
		var entries map[string]any
		err := json.Unmarshal([]byte(serializedJson), &entries)
		if err != nil {
			return err
		}
		return writeEntries(directoryOutput, entries, serializedFormat)
	}
	return nil
}

func writeEntries(directory string, entries map[string]any, serialized bool) error {
	for filename, c := range entries {
		switch content := c.(type) {
		case map[string]any:
			err := writeEntries(filepath.Join(directory, filename), content, serialized)
			if err != nil {
				return err
			}
		default:
			err := writeFile(filepath.Join(directory, filename), content, serialized)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(filename string, content any, serialized bool) error {
	var fileContent []byte
	if serialized {
		var err error
		fileContent, err = json.Marshal(content)
		if err != nil {
			return err
		}
	} else {
		stringContent, ok := content.(string)
		if !ok {
			return fmt.Errorf("expect string when writing output to file: %s, but got %T", filename, content)
		}
		fileContent = []byte(stringContent)
	}

	_, err := os.Stat(filename)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err == nil {
		existingContent, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if bytes.Equal(existingContent, fileContent) {
			return nil
		}
	}

	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, fileContent, 0666)
	if err != nil {
		return err
	}
	return nil
}

func EvalNodeInput(node ast.Node) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = &node
		c.snippetInput = nil
		c.fileInput = nil
	}
}

func EvalSnippetInput(filename, snippet string) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = &snippetInput{filename, snippet}
		c.fileInput = nil
	}
}

func EvalFileInput(filename string) EvalOption {
	return func(c *evalConfig) {
		c.nodeInput = nil
		c.snippetInput = nil
		c.fileInput = &filename
	}
}

func EvalWriterOutput(w io.Writer) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = w
		c.valueOutput = nil
		c.directoryOutput = ""
	}
}

func EvalValueOutput(out any) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = out
		c.directoryOutput = ""
	}
}

func EvalDirectoryOutput(dir string) EvalOption {
	return func(c *evalConfig) {
		c.writerOutput = nil
		c.valueOutput = nil
		c.directoryOutput = dir
	}
}

func EvalSerialize(s bool) EvalOption {
	return func(c *evalConfig) {
		c.serializedFormat = s
	}
}

func EvalInvocations(out *[]Invocation) EvalOption {
	return func(c *evalConfig) {
		c.recordedInvocations = out
	}
}

func EvalImportedPaths(out *[]string) EvalOption {
	return func(c *evalConfig) {
		c.recordedPaths = out
	}
}
