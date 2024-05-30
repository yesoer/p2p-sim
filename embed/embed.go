package embed

import (
	"distributed-sys-emulator/log"
	"embed"
	"os"
)

type CodeExample struct {
	Name string // the name of the embedded file
	Path string // the path of the temporary file
}

// embed code examples
//
//go:embed resources
var content embed.FS

var codeExamples []CodeExample

const codeExamplePath = "resources"
const tempDestPath = "p2p-sim"

// Get a list of all embedded files and the path to the corresponding temporary
// local file
func GetExampleList() []CodeExample {
	return codeExamples
}

// Creates temporary directories for the embedded examples, so that they can be
// handled like all other projects (edited, run etc.)
// As of now that means any changes to examples are not persistent.
// The path to the temporary root is returned for cleanup.
func InitEmbeddedExamples() string {
	entries, err := content.ReadDir(codeExamplePath)
	if err != nil {
		log.Error(err)
		return ""
	}

	wrapDir, err := os.MkdirTemp("", tempDestPath)
	if err != nil {
		log.Error(err)
		return ""
	}

	for _, entry := range entries {
		projectDir, err := os.MkdirTemp(wrapDir, "*_"+entry.Name())
		if err != nil {
			log.Error(err)
			continue
		}

		file, err := os.CreateTemp(projectDir, "*_"+entry.Name())
		if err != nil {
			log.Error(err)
			continue
		}
		newExample := CodeExample{entry.Name(), projectDir}
		codeExamples = append(codeExamples, newExample)

		code, err := content.ReadFile(codeExamplePath + "/" + entry.Name())
		if err != nil {
			log.Error(err)
			continue
		}

		_, err = file.Write(code)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Debug("Created temporary file ", file.Name())
	}

	return wrapDir
}
