package main

import (
	"bytes"
	"errors"
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

			if endGo > -1 {
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
			} else {
				return errors.New("Unable to parse")
			}
		default:
			t.Segments = append(t.Segments, StringSegment(unparsed))
			unparsed = nil
			break
		}
	}

	return nil
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

func main() {

	templates := make([]*Template, 0, len(os.Args[1:]))

	for _, path := range os.Args[1:] {
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var t Template
		err = t.Parse(fileBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid template format: %v", err)
			os.Exit(1)
		}
		templates = append(templates, &t)
	}

	imports := make(map[string]bool)
	for _, t := range templates {
		for pkg, _ := range t.Imports {
			imports[pkg] = true
		}
	}

	fmt.Printf("package main\n")
	fmt.Printf("import (\n")
	for pkg, _ := range imports {
		fmt.Printf("\"%s\"\n", pkg)
	}
	fmt.Printf(")\n")

	for _, t := range templates {
		fmt.Printf("func %s(%s) (err error) {\n", t.FuncName, t.Parameters)
		for _, s := range t.Segments {
			s.WriteTo(os.Stdout)
		}
		fmt.Printf("\nreturn\n}\n")
	}
}
