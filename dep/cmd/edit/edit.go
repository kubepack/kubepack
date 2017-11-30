package edit

import (
	"github.com/spf13/cobra"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"io"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"io/ioutil"
)

const defaultEditor = "nano"

func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (filename)",
		Short: "Edit a file.",
		Long:  "Edit a _vendor file to generate kubectl patch.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello edit Commands!!!!", args)
			RunEdit(cmd)
		},
	}

	AddEditFlag(cmd)

	return cmd
}

func AddEditFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "Edit file, provided through the --file flag.")
}

func RunEdit(cmd *cobra.Command) {
	s, err := cmd.Flags().GetString("file")
	if err != nil {
		log.Fatalln("Error occurred during get flag", err)
	}

	root, err := os.Getwd()
	if err != nil {
		log.Fatalln("Error during get root path.", err)
	}
	path := filepath.Join(root, s)
	fmt.Println("Hello path......", path)

	file, err := os.Stat(path)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(file.Name())

	fileByte, err := ioutil.ReadFile(path)
	// fileInfo, err := io.ReadFile(path)

	fmt.Println("Hello file byte array....", string(fileByte))
	// fmt.Println("Hello file byte array....", fileByte)

	// CopyFileToTempDir(path)

	edit := NewDefaultEditor()
}

func CopyFileToTempDir(src string, file os.FileInfo)  {
	// Open original file
	originalFile, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer originalFile.Close()

	for i := 0; i < 10000; i++  {
		tmpDir := os.TempDir()
		path := filepath.Join(tmpDir, file.Name())

		if  {

		}
	}
	// Create new file
	newFile, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	// Copy the bytes to destination from source
	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Copied %d bytes.", bytesWritten)

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func NewDefaultEditor() editor.Editor {
	return editor.Editor{
		Args:  []string{defaultEditor},
		Shell: false,
	}
}
