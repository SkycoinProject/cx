package playground

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	cxcore "github.com/skycoin/cx/cx"
	"github.com/skycoin/cx/cxgo/actions"
	cxgo0 "github.com/skycoin/cx/cxgo/cxgrammer"
	cxgo "github.com/skycoin/cx/cxgo/cxparser"
	"github.com/skycoin/cx/cxgo/parser"
)

var (
	exampleCollection map[string]string
	exampleNames      []string
	examplesDir       string
)

type ExampleContent struct {
	ExampleName string
}

var InitPlayground = func(workingDir string) error {
	examplesDir = filepath.Join(workingDir, "../../examples")
	exampleCollection = make(map[string]string)

	exampleInfoList, err := ioutil.ReadDir(examplesDir)
	if err != nil {
		fmt.Printf("Fail to get file list under examples dir: %s\n", err.Error())
		return err
	}
	for _, exp := range exampleInfoList {
		if exp.IsDir() {
			continue
		}
		path := filepath.Join(examplesDir, exp.Name())
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Fail to read example file %s\n", path)
			// coninue if fail to read the current example file
			continue
		}

		exampleName := exp.Name()
		exampleNames = append(exampleNames, exampleName)
		exampleCollection[exampleName] = string(bytes)
	}

	return nil
}

func GetExampleFileList(w http.ResponseWriter, r *http.Request) {

	exampleNamesBytes, err := json.Marshal(exampleNames)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(exampleNamesBytes))
}

func GetExampleFileContent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		b   []byte
		err error
	)
	if b, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var example ExampleContent
	if err := json.Unmarshal(b, &example); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	content, err := getExampleFileContent(example.ExampleName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, content)
}

var getExampleFileContent = func(exampleName string) (string, error) {
	exampleContent, ok := exampleCollection[exampleName]
	if !ok {
		err := fmt.Errorf("example name %s not found", exampleName)

		return "", err
	}
	return exampleContent, nil
}

type SourceCode struct {
	Code string `json:"code,omitempty"`
}

func RunProgram(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var b []byte
	var err error
	if b, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var source SourceCode
	if err := json.Unmarshal(b, &source); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := r.ParseForm(); err == nil {
		fmt.Fprintf(w, "%s", eval(source.Code+"\n"))
	}
}

func unsafeeval(code string) (out string) {
	var lexer *parser.Lexer
	defer func(lexer *parser.Lexer) {
		if r := recover(); r != nil {
			out = fmt.Sprintf("%v", r)
			// lexer.Stop()
			return
		}
	}(lexer)

	// storing strings sent to standard output
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	os.Stdout = w

	actions.LineNo = 0

	actions.PRGRM = cxcore.MakeProgram()
	cxgo0.PRGRM0 = actions.PRGRM

	cxgo0.Parse(code)

	actions.PRGRM = cxgo0.PRGRM0

	lexer = parser.NewLexer(bytes.NewBufferString(code))
	parser.Parse(lexer)
	//yyParse(lexer)

	err = cxgo.AddInitFunction(actions.PRGRM)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	if err := actions.PRGRM.RunCompiled(0, nil); err != nil {
		actions.PRGRM = cxcore.MakeProgram()
		return fmt.Sprintf("%s", err)
	}

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old // restoring the real stdout
	out = <-outC

	actions.PRGRM = cxcore.MakeProgram()
	return out
}

func eval(code string) string {
	runtime.GOMAXPROCS(2)
	ch := make(chan string, 1)

	var result string

	go func() {
		result = unsafeeval(code)
		ch <- result
	}()

	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()

	select {
	case <-ch:
		return result
	case <-timer.C:
		actions.PRGRM = cxcore.MakeProgram()
		return "Timed out."
	}
}
