package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkch/gg"
)

// bcc is the structure of bcc format.
type bcc struct {
	Body []struct {
		From    float32 `json:"from"`
		To      float32 `json:"to"`
		Content string  `json:"content"`
	}
}

// parseBcc parses a bcc format.
func parseBcc(r io.Reader) (*bcc, error) {
	var ret bcc
	if err := json.NewDecoder(r).Decode(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// srtTime convert a time(how many seconds) to the format
// used by srt file.
func srtTime(sec float32) string {
	n := int(sec)
	h := n / 3600
	n = n % 3600
	m := n / 60
	s := n % 60
	// Do not use
	// millis := int((sec - float32(n)) * 1000)
	// Float point error!!
	millis := fmt.Sprintf("%.3f", sec)
	millis = millis[strings.IndexRune(millis, '.')+1:]
	return fmt.Sprintf("%02v:%02v:%02v,%v", h, m, s, millis)
}

var srtTmpl = template.Must(template.New("srt").
	Funcs(template.FuncMap{
		"time": srtTime,
		"add":  func(a, b int) int { return a + b }}).
	Parse(`
{{- range $i, $e :=.Body -}}
{{add $i 1}}
{{time $e.From}} --> {{time $e.To}}
{{$e.Content}}

{{end}}`))

func bcc2srt(r io.Reader, w io.Writer) error {
	bcc, err := parseBcc(r)
	if err != nil {
		return err
	}

	return srtTmpl.Execute(w, &bcc)
}

func errorExit(code int, msg string) {
	os.Stderr.WriteString(msg + "\n")
	os.Exit(1)
}

func exec0() {
	err := bcc2srt(os.Stdin, os.Stdout)
	if err != nil {
		errorExit(1, err.Error())
	}
}

// changeExt change the extension name of the last element of
// path to ext. If there is no extension name in path, a new
// extension name will be added.
func changeExt(path, ext string) string {
	if dot := strings.LastIndexByte(path, '.'); dot != -1 {
		return path[:dot] + "." + ext
	}
	return path + "." + ext
}

func exec1() {
	var inFile = flag.Arg(0)
	r, err := os.Open(inFile)
	if err != nil {
		errorExit(1, err.Error())
	}
	defer r.Close()
	var outFile = out
	if outFile == "" {
		outFile = changeExt(inFile, "srt")
	}
	w, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		errorExit(1, err.Error())
	}
	defer w.Close()
	err = bcc2srt(r, w)
	if err != nil {
		errorExit(1, err.Error())
	}
}

func execN() {
	if out != "" {
		if err := os.MkdirAll(out, 0777); err != nil && !os.IsExist(err) {
			errorExit(1, err.Error())
		}
	}
	var errored bool
	for _, inFile := range flag.Args() {
		// Use a func to let the defers execute before the next file.
		func() {
			r, err := os.Open(inFile)
			if err != nil {
				errored = true
				fmt.Fprintf(os.Stdout, "%v: %v\n", inFile, err)
				return
			}
			defer r.Close()
			var outFile = gg.If(
				out == "",
				changeExt(inFile, ".srt"),
				filepath.Join(out, changeExt(filepath.Base(inFile), "srt")))
			w, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
			if err != nil {
				errored = true
				fmt.Fprintf(os.Stdout, "%v: %v\n", inFile, err)
				return
			}
			defer w.Close()
			err = bcc2srt(r, w)
			if err != nil {
				errored = true
				fmt.Fprintf(os.Stdout, "%v: %v\n", inFile, err)
				return
			}
		}()
	}
	if errored {
		os.Exit(2)
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of %s:

bcc2srt [-out output] [input ...]

`, os.Args[0])
		flag.PrintDefaults()
	}
}

var out string

func main() {
	flag.StringVar(&out, "out", "", "The output file path if there is only one input file or the output dir if multiple input files. Default to input.srt or stdout if input is missing")
	flag.Parse()

	if flag.NArg() == 0 {
		exec0()
	} else if flag.NArg() == 1 {
		exec1()
	} else {
		execN()
	}
}
