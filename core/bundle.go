package core

import (
	"bytes"
	"errors"
	"github.com/yesoer/p2p-sim/bus"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// ConcatenateGoFiles concatenates code from all .go files in a local directory
// and its subdirectories, unifying their headers
func concatenateGoFiles(projectPath string) (Code, error) {
	var concatenatedCode string
	concat := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".go" {
			code, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if concatenatedCode != "" {
				code = []byte(removeHeader(string(code)))
			}

			concatenatedCode += string(code)
		}
		return nil
	}

	// walk the directory and concatenate the code
	err := filepath.Walk(projectPath, concat)
	if err != nil {
		return "", err
	}

	concatenatedCode, err = runGoImports(concatenatedCode)
	if err != nil {
		return "", err
	}

	return Code(concatenatedCode), nil
}

// Uses the goimports command to format the code and add missing imports
// TODO : This is a bit of a hack, but it seems difficult to find or build a
// goimports package that works on strings. Here are some resources :
//   - https://github.com/golangci/gofmt/blob/master/goimports/golangci.go
//   - https://cs.opensource.google/go/x/tools/+/master:imports/forward.go
//   - https://cs.opensource.google/go/x/tools/+/master:internal/imports/imports.go;drc=4db45793ff6b9988dcf2ca4cfc42d2ef51ab971e;l=50
//
// using the imports package might be prettier than using the command so users don't have to worry about installing it. See :
//   - https://pkg.go.dev/golang.org/x/tools/imports#pkg-functions
func runGoImports(code string) (string, error) {
	goimportsCmd := exec.Command("goimports")

	var out bytes.Buffer
	goimportsCmd.Stdin = bytes.NewBufferString(code)
	goimportsCmd.Stdout = &out

	if err := goimportsCmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

// All occurrences of the pattern are removed from the input
func applyRgxFilter(input string, pattern string) string {
	regex := regexp.MustCompile(pattern)
	return regex.ReplaceAllLiteralString(input, "")
}

// Removes the package and import statements from the code
func removeHeader(input string) string {
	res := applyRgxFilter(input, `package\s+[a-zA-Z_]+`)
	res = applyRgxFilter(res, `import\s+".*"`)
	res = applyRgxFilter(res, `import\s+\((\s*"(.*)"\s)*\s*\)`)

	return res
}

// RunPackage executes the "Run" function of the designated package and returns
// its output
// TODO : For now this works quite well for simple projects but one might want
// to consider alternatives in the future. For example :
// - https://pkg.go.dev/golang.org/x/tools/cmd/bundle <- seems to be exactly what we need and from the go team
// - https://github.com/naegelejd/gocat <- built on top of bundle, don't know its benefits
func bundle(project bus.Source) (Code, error) {
	if project.Type != bus.Directory {
		err := errors.New(`Unsupported project type in bundle.go, which as of 
			now intends to only support directories`)
		return Code(""), err
	}

	return concatenateGoFiles(project.Path)
}
