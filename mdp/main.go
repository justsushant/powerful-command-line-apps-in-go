package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"html/template"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
<html>
	<head>
		<meta http=equiv="content-type" content="text/html"; charset="utf-8">
		<title>{{.Title}}</title>
	</head>
	<body>
	{{.Body}}
	</body>
	<footer>FILENAME: {{.Filename}}</footer>
</html>
`
)

// content type represents the HTML content to add into the template
type content struct {
	Title string
	Body template.HTML
	Filename string
}

func main() {
	// parsing flags
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	tFname := flag.String("t", "", "Alternate template name")
	flag.Parse()

	// if no input, show usage info
	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *tFname, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run (filename, tFname string, out io.Writer, skipPreview bool) error {
	// read all the data from the input file and check for errors
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, filename, tFname)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp("", "mdp*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()
	fmt.Fprintln(out, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}

	defer os.Remove(outName)

	return preview(outName)
}

// parses the markdown file through blackfriday and bluemonday
// for generating a valid and safe html
func parseContent(input []byte, filename, tFname string) ([]byte, error) {
	// parse markdown to generate valid & safe html
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	// parse the contents of defaultTemplate into new template
	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	if tFname != "" {
		t, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, err
		}
	}

	// instantiate the content type, adding the title and body
	c := content {
		Title: "Markdown Preview Tool",
		Body: template.HTML(body),
		Filename: filename,
	}

	// create a buffer of bytes to write to file
	var buffer bytes.Buffer

	// executing the template with the content type
	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// 0644 permisson allows write by owner and read by anyone
func saveHTML(outFname string, data []byte) error {
	return os.WriteFile(outFname, data, 0644)
}


func preview(fname string) error {
	cName := ""
	cParams := []string{}

	// define executable based on the OS
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

	// append filename to parameters slice
	cParams = append(cParams, fname)
	
	// locate executable in PATH
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	// open the file using default program
	err = exec.Command(cPath, cParams...).Run()

	// give browser some time to open the file before deleting it
	time.Sleep(2 * time.Second)
	return err
}

