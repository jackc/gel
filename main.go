package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Template struct {
	FuncName   string
	Parameters string
	Imports    map[string]bool
	Escape     string
	Segments   []io.WriterTo
}

type Imports []string

func (i Imports) WriteTo(w io.Writer) (total int64, err error) {
	var n int

	n, err = fmt.Fprintf(w, "import (\n")
	total += int64(n)
	if err != nil {
		return
	}

	for _, pkg := range i {
		n, err = fmt.Fprintf(w, "\"%s\"\n", pkg)
		total += int64(n)
		if err != nil {
			return
		}
	}

	n, err = fmt.Fprintf(w, ")\n")
	total += int64(n)
	if err != nil {
		return
	}

	return
}

type StringSegment []byte

func (s StringSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(writer, `", s, "`)\n")
}

type GoSegment []byte

func (s GoSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeMultiple(w, s, []byte("\n"))
}

type IntegerInterpolationSegment []byte

func (s IntegerInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(writer, strconv.FormatInt(int64(", s, "), 10))\n")
}

type RawStringInterpolationSegment []byte

func (s RawStringInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(writer, ", s, ")\n")
}

type HTMLEscapedStringInterpolationSegment []byte

func (s HTMLEscapedStringInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(writer, html.EscapeString(", s, "))\n")
}

func (t *Template) ParseFile(path string) error {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = t.Parse(fileBytes)
	if err != nil {
		return err
	}

	return nil
}

func (t *Template) Parse(templateBytes []byte) (err error) {
	regions := bytes.SplitN(templateBytes, []byte("\n---\n"), 2)

	if len(regions) != 2 {
		return errors.New("Did not find divider between header and body.")
	}

	err = t.parseHeader(regions[0])
	if err != nil {
		return err
	}

	err = t.parseBody(regions[1])
	if err != nil {
		return err
	}

	return nil
}

func (t *Template) parseHeader(header []byte) (err error) {
	lines := bytes.Split(header, []byte("\n"))
	options := make(map[string]string, len(lines))
	for i, l := range lines {
		pair := bytes.SplitN(l, []byte(":"), 2)
		if len(pair) != 2 {
			return fmt.Errorf("Bad header line: %d", i)
		}
		options[string(pair[0])] = string(bytes.Trim(pair[1], " "))
	}

	t.FuncName = options["func"]
	if t.FuncName == "" {
		return errors.New(`Missing "func"`)
	}

	t.Escape = options["escape"]

	t.Parameters = "writer io.Writer"
	var extraParameters string
	extraParameters = options["parameters"]
	if len(extraParameters) > 0 {
		t.Parameters = t.Parameters + ", " + extraParameters
	}

	t.Imports = map[string]bool{"io": true}

	var extraImports string
	extraImports = options["imports"]
	if len(extraImports) > 0 {
		for _, pkg := range strings.Split(extraImports, " ") {
			t.Imports[strings.Trim(pkg, " ")] = true
		}
	}

	return nil
}

func (t *Template) parseBody(body []byte) (err error) {
	unparsed := body

	for len(unparsed) > 0 {
		next := bytes.Index(unparsed, []byte("<%"))
		switch {
		case next > 0:
			segment := unparsed[:next]
			t.Segments = append(t.Segments, StringSegment(segment))
			unparsed = unparsed[next:]
		case next == 0:
			unparsed = unparsed[2:]
			endGo := bytes.Index(unparsed, []byte("%>"))
			if endGo == -1 {
				return errors.New("Unable to parse")
			}

			segment := unparsed[:endGo]

			if segment[0] == '=' {
				if segment[1] == 'i' {
					t.Imports["strconv"] = true
					t.Segments = append(t.Segments, IntegerInterpolationSegment(segment[2:]))
				} else {
					switch t.Escape {
					case "":
						t.Segments = append(t.Segments, RawStringInterpolationSegment(segment[1:]))
					case "html":
						t.Imports["html"] = true
						t.Segments = append(t.Segments, HTMLEscapedStringInterpolationSegment(segment[1:]))
					default:
						return errors.New("Unknown escape type")
					}
				}
			} else {
				t.Segments = append(t.Segments, GoSegment(segment))
			}

			unparsed = unparsed[endGo+2:]

		default:
			t.Segments = append(t.Segments, StringSegment(unparsed))
			unparsed = nil
			break
		}
	}

	return nil
}

func (t *Template) WriteTo(w io.Writer) (total int64, err error) {
	var n int
	n, err = fmt.Printf("func %s(%s) (err error) {\n", t.FuncName, t.Parameters)
	total += int64(n)
	if err != nil {
		return
	}

	for _, s := range t.Segments {
		var n64 int64
		n64, err = s.WriteTo(os.Stdout)
		total += n64
		if err != nil {
			return
		}
	}

	n, err = fmt.Printf("\nreturn\n}\n")
	total += int64(n)
	if err != nil {
		return
	}

	return
}

func writeWrapped(w io.Writer, prefix string, data []byte, suffix string) (count int64, err error) {
	return writeMultiple(w, []byte(prefix), data, []byte(suffix))
}

func writeMultiple(w io.Writer, segments ...[]byte) (count int64, err error) {
	for _, s := range segments {
		var n int
		n, err = w.Write(s)
		count += int64(n)
		if err != nil {
			return
		}
	}

	return count, err
}

func parseTemplateFiles(paths []string) (templates []*Template, err error) {
	for _, path := range paths {
		var t Template
		err := t.ParseFile(path)
		if err != nil {
			return templates, fmt.Errorf("Unable to parse file %v: %v", path, err)
		}
		templates = append(templates, &t)
	}

	return templates, nil
}

func extractImports(templates []*Template) (imports Imports) {
	importSet := make(map[string]bool)

	for _, t := range templates {
		for pkg, _ := range t.Imports {
			importSet[pkg] = true
		}
	}

	for pkg, _ := range importSet {
		imports = append(imports, pkg)
	}

	return imports
}

func main() {
	var args struct {
		pkg string
	}

	flag.StringVar(&args.pkg, "package", "main", "package to which compiled templates belong")
	flag.Parse()

	templates, err := parseTemplateFiles(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = fmt.Fprintf(os.Stdout, "package %s\n", args.pkg)
	if err != nil {
		return
	}

	imports := extractImports(templates)
	_, err = imports.WriteTo(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, t := range templates {
		_, err := t.WriteTo(os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
