package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	header = `<!DOCTYPE html>
	<html>
	<head>
	<meta http-equiv="content-type" content="text/html; charset=utf-8">
	<title>Markdown Preview Tool</title>
	</head>
	<body>
	`
	footer = `
	</body>
	</html>
	`
	)

	func main(){
		files := flag.String("file","","Markdown to preview")
		skipPreview := flag.Bool("s",false,"Skip auto-preview")

		flag.Parse()

		if *files == "" {
			flag.Usage()
			os.Exit(1)
		}
		if err := run(*files,os.Stdout,*skipPreview); err != nil {
				fmt.Fprintf(os.Stderr,"An error occurred: %v\n",err)
				os.Exit(1)
		}
	}

	func run(filename string,out io.Writer,skipPreview bool) error{
		// Specify the base directory for the temporary directory
		baseDir,err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			return nil
		}
		// Create a temporary directory inside the specified base directory
		tempDir, err := os.MkdirTemp(baseDir, "tmp")
		if err != nil {
			fmt.Println("Error creating temporary directory:", err)
			return nil
		}
		input, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		htmlData := parseContent(input)
		// Create temporary file and check for errors
		temp,err :=  os.CreateTemp(tempDir,"mdp*.html")
		if err != nil {
			return err
		}
		if err := temp.Close(); err != nil {
			return err
		}
		outName := temp.Name()
		
		fmt.Fprintln(out,outName)
		if err := saveHTML(outName, htmlData); err != nil {
			return err
		}
		if skipPreview {
			return nil
		}
		defer os.Remove(outName)
		return preview(outName)
	}

	func parseContent(input []byte) []byte{
		// Parse the markdown file through blackfriday and bluemonday
		// to generate a valid and safe HTML
		output := blackfriday.Run(input)
		body := bluemonday.UGCPolicy().SanitizeBytes(output)
		var buffer bytes.Buffer
		buffer.WriteString(header)
		buffer.Write(body)
		buffer.WriteString(footer)
		return buffer.Bytes()
	}

	func saveHTML(fileName string,htmlData []byte) error{
		return os.WriteFile(fileName,htmlData,0644)
	}

	func preview(fname string) error {
		cName := ""
		cParams := []string{}
		switch runtime.GOOS {
			case "linux":
				cName = "xdg-open"
			case "windows":
				cName = "cmd.exe"
				cParams = []string{"/C", "start"}
			case "darwin":
				cName = "open"
			default:
				return fmt.Errorf("OS not supported")
		}
		cParams = append(cParams, fname)
		// Locate executable in PATH
		cPath, err := exec.LookPath(cName)
		if err != nil {
			return err
		}
		// Open the file using default program
		err = exec.Command(cPath,cParams...).Run()
		// Give the browser some time to open the file before deleting it
		time.Sleep(2 * time.Second)
		return err
	}