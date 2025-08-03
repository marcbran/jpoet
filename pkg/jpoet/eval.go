package jpoet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/marcbran/jpoet/internal/plugin"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Eval struct {
	vm       *jsonnet.VM
	importer CompoundImporter
	contents map[string]jsonnet.Contents
	closers  []io.Closer

	nodeInput    *ast.Node
	snippetInput *SnippetInput
	fileInput    *string

	writerOutput    io.Writer
	valueOutput     any
	directoryOutput string

	serializedFormat bool

	errs []error
}

func NewEval() *Eval {
	vm := jsonnet.MakeVM()
	importer := CompoundImporter{}
	return &Eval{
		vm:       vm,
		importer: importer,
		contents: make(map[string]jsonnet.Contents),

		// Default: output to stdout
		writerOutput: os.Stdout,

		// Default: output serialized Json
		serializedFormat: true,
	}
}

func (e *Eval) hasInput() bool {
	return e.nodeInput != nil || e.snippetInput != nil || e.fileInput != nil
}

func (e *Eval) error() error {
	if len(e.errs) == 0 {
		return nil
	}
	if len(e.errs) == 1 {
		return fmt.Errorf("failed to evaluate Jsonnet: %w", e.errs[0])
	}
	return fmt.Errorf("failed to evalute Jsonnet: %s", e.errs)
}

func (e *Eval) TLAVar(key string, val string) *Eval {
	e.vm.TLAVar(key, val)
	return e
}

func (e *Eval) TLACode(key string, val string) *Eval {
	e.vm.TLACode(key, val)
	return e
}

func (e *Eval) TLANode(key string, node ast.Node) *Eval {
	e.vm.TLANode(key, node)
	return e
}

type Input interface {
	Input() Input
}

type NodeInput struct {
	Node ast.Node
}

func (i NodeInput) Input() Input {
	return i
}

type SnippetInput struct {
	Filename string
	Snippet  string
}

func (i SnippetInput) Input() Input {
	return i
}

type FileInput struct {
	Filename string
}

func (i FileInput) Input() Input {
	return i
}

func (e *Eval) Input(input Input) *Eval {
	switch i := input.(type) {
	case NodeInput:
		return e.NodeInput(i.Node)
	case SnippetInput:
		return e.SnippetInput(i.Filename, i.Snippet)
	case FileInput:
		return e.FileInput(i.Filename)
	}
	return e
}

func (e *Eval) NodeInput(node ast.Node) *Eval {
	e.nodeInput = &node
	e.snippetInput = nil
	e.fileInput = nil
	return e
}

func (e *Eval) SnippetInput(filename string, snippet string) *Eval {
	e.nodeInput = nil
	e.snippetInput = &SnippetInput{filename, snippet}
	e.fileInput = nil
	return e
}

func (e *Eval) FileInput(filename string) *Eval {
	e.nodeInput = nil
	e.snippetInput = nil
	e.fileInput = &filename
	return e
}

func (e *Eval) Importer(i jsonnet.Importer) *Eval {
	e.importer.Importers = append(e.importer.Importers, i)
	return e
}

func (e *Eval) FileImport(jpaths []string) *Eval {
	return e.Importer(&jsonnet.FileImporter{JPaths: jpaths})
}

func (e *Eval) FSImport(f fs.FS) *Eval {
	return e.Importer(&FSImporter{Fs: f})
}

func (e *Eval) StringImport(filename string, value string) *Eval {
	e.contents[filename] = jsonnet.MakeContents(value)
	return e
}

func (e *Eval) NativeFunction(f *jsonnet.NativeFunction) *Eval {
	if f == nil {
		return e
	}
	e.vm.NativeFunction(f)
	return e
}

func (e *Eval) Plugin(p *Plugin) *Eval {
	return e.NativeFunction(p.NativeFunction())
}

type CloserFunction interface {
	io.Closer
	Function() *jsonnet.NativeFunction
}

func (e *Eval) CloserFunction(f CloserFunction) *Eval {
	if f == nil {
		return e
	}
	e.closers = append(e.closers, f)
	e.vm.NativeFunction(f.Function())
	return e
}

func (e *Eval) PluginsDir(pluginsDir string) *Eval {
	entries, err := readEntries(pluginsDir)
	if err != nil {
		e.errs = append(e.errs, err)
		return e
	}
	for _, entry := range entries {
		name := entry.Name()
		client, err := plugin.NewClient(filepath.Join(pluginsDir, name, name))
		if err != nil {
			e.errs = append(e.errs, err)
			continue
		}
		e = e.CloserFunction(client)
	}
	return e
}

func (e *Eval) WriterOutput(w io.Writer) *Eval {
	e.writerOutput = w
	e.valueOutput = nil
	e.directoryOutput = ""
	return e
}

func (e *Eval) ValueOutput(out any) *Eval {
	e.writerOutput = nil
	e.valueOutput = out
	e.directoryOutput = ""
	return e
}

func (e *Eval) DirectoryOutput(dir string) *Eval {
	e.writerOutput = nil
	e.valueOutput = nil
	e.directoryOutput = dir
	return e
}

func (e *Eval) Serialize(s bool) *Eval {
	e.serializedFormat = s
	return e
}

func (e *Eval) Eval() error {
	defer func() {
		for _, closer := range e.closers {
			err := closer.Close()
			if err != nil {
				e.errs = append(e.errs, err)
			}
		}
	}()
	if !e.hasInput() {
		e.errs = append(e.errs, errors.New("missing input"))
		return e.error()
	}
	if len(e.contents) > 0 {
		e.Importer(&MemoryImporter{
			Data: e.contents,
		})
	}
	if len(e.importer.Importers) > 0 {
		e.vm.Importer(e.importer)
	}

	var serializedJson string
	var err error
	if e.nodeInput != nil {
		serializedJson, err = e.vm.Evaluate(*e.nodeInput)
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
	} else if e.snippetInput != nil {
		serializedJson, err = e.vm.EvaluateAnonymousSnippet(e.snippetInput.Filename, e.snippetInput.Snippet)
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
	} else if e.fileInput != nil {
		serializedJson, err = e.vm.EvaluateFile(*e.fileInput)
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
	}

	if e.writerOutput != nil {
		output := serializedJson
		if !e.serializedFormat {
			err := json.Unmarshal([]byte(serializedJson), &output)
			if err != nil {
				e.errs = append(e.errs, err)
				return e.error()
			}
		}
		_, err := e.writerOutput.Write([]byte(output))
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
	} else if e.valueOutput != nil {
		if e.serializedFormat {
			e.valueOutput = serializedJson
		} else {
			err := json.Unmarshal([]byte(serializedJson), e.valueOutput)
			if err != nil {
				e.errs = append(e.errs, err)
				return e.error()
			}
		}
	} else if e.directoryOutput != "" {
		var entries map[string]any
		err = json.Unmarshal([]byte(serializedJson), &entries)
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
		err = writeEntries(e.directoryOutput, entries, e.serializedFormat)
		if err != nil {
			e.errs = append(e.errs, err)
			return e.error()
		}
	}
	return e.error()
}

func readEntries(pluginsDir string) ([]os.DirEntry, error) {
	_, err := os.Stat(pluginsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}
	return entries, nil
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
