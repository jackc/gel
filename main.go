package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Template struct {
	FuncName   string
	Parameters string
	Imports    map[string]bool
	Escape     string
	Body       bytes.Buffer
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
			t.writeStringSegment(segment)
			unparsed = unparsed[next:]
		case next == 0:
			unparsed = unparsed[2:]
			endGo := bytes.Index(unparsed, []byte("%>"))

			if endGo > -1 {
				segment := unparsed[:endGo]

				if segment[0] == '=' {
					t.writeInterpolationSegment(segment[1:])
				} else {
					t.writeGoSegment(segment)
				}

				unparsed = unparsed[endGo+2:]
			} else {
				return errors.New("Unable to parse")
			}
		default:
			segment := unparsed
			t.writeStringSegment(segment)
			unparsed = nil
			break
		}
	}

	return nil
}

func (t *Template) writeStringSegment(segment []byte) (err error) {
	t.Body.WriteString("io.WriteString(writer, `")
	t.Body.Write(segment)
	t.Body.WriteString("`)\n")
	return nil
}

func (t *Template) writeInterpolationSegment(segment []byte) (err error) {
	switch {
	case segment[0] == 'i':
		t.Imports["strconv"] = true
		segment = segment[1:]
		t.writeIntegerInterpolationSegment(segment)
	default:
		t.writeStringInterpolationSegment(segment)
	}

	return nil
}

func (t *Template) writeIntegerInterpolationSegment(segment []byte) (err error) {
	t.Body.WriteString("io.WriteString(writer, strconv.FormatInt(int64(")
	t.Body.Write(segment)
	t.Body.WriteString("), 10))\n")
	return nil
}

func (t *Template) writeStringInterpolationSegment(segment []byte) (err error) {
	switch t.Escape {
	case "":
		return t.writeRawStringInterpolationSegment(segment)
	case "html":
		t.Imports["html"] = true
		return t.writeHTMLEscapedStringInterpolationSegment(segment)
	default:
		return errors.New("Unknown escape type")
	}
}

func (t *Template) writeRawStringInterpolationSegment(segment []byte) (err error) {
	t.Body.WriteString("io.WriteString(writer, ")
	t.Body.Write(segment)
	t.Body.WriteString(")\n")
	return nil
}

func (t *Template) writeHTMLEscapedStringInterpolationSegment(segment []byte) (err error) {
	t.Body.WriteString("io.WriteString(writer, html.EscapeString(")
	t.Body.Write(segment)
	t.Body.WriteString("))\n")
	return nil
}

func (t *Template) writeGoSegment(segment []byte) (err error) {
	t.Body.Write(segment)
	t.Body.WriteString("\n")
	return nil
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

		fmt.Printf("func %s(%s) (err error) {\n%s\nreturn\n}\n", t.FuncName, t.Parameters, t.Body.String())
	}
}
