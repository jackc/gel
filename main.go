package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type Template struct {
	FuncName   string
	Parameters string
	Imports    map[string]bool
	Escape     string
	Body       bytes.Buffer
}

func parseTemplate(templateBytes []byte) (t *Template, err error) {
	t = new(Template)

	regions := bytes.SplitN(templateBytes, []byte("\n---\n"), 2)

	if len(regions) != 2 {
		return nil, errors.New("Did not find divider between header and body.")
	}

	yf := new(yaml.File)
	yf.Root, err = yaml.Parse(bytes.NewReader(regions[0]))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse YAML header: %v", err)
	}

	t.FuncName, err = yf.Get("func")
	if err != nil {
		return nil, errors.New(`Missing "func"`)
	}

	t.Escape, err = yf.Get("escape")

	t.Parameters = "writer io.Writer"
	var extraParameters string
	extraParameters, _ = yf.Get("parameters")
	if len(extraParameters) > 0 {
		t.Parameters = t.Parameters + ", " + extraParameters
	}

	t.Imports = map[string]bool{"io": true}

	var extraImports string
	extraImports, _ = yf.Get("imports")
	if len(extraImports) > 0 {
		for _, pkg := range strings.Split(extraImports, " ") {
			t.Imports[strings.Trim(pkg, " ")] = true
		}
	}

	unparsed := regions[1]

	for len(unparsed) > 0 {
		next := bytes.Index(unparsed, []byte("<%"))
		switch {
		case next > 0:
			segment := unparsed[:next]
			parseStringSegment(&t.Body, segment)
			unparsed = unparsed[next:]
		case next == 0:
			unparsed = unparsed[2:]
			endGo := bytes.Index(unparsed, []byte("%>"))

			if endGo > -1 {
				segment := unparsed[:endGo]
				switch {
				case bytes.HasPrefix(segment, []byte("=i")):
					t.Imports["strconv"] = true
					segment = segment[2:]
					parseIntegerInterpolationSegment(&t.Body, segment)

				case bytes.HasPrefix(segment, []byte("=")):
					segment = segment[1:]
					switch t.Escape {
					case "":
						parseStringInterpolationSegment(&t.Body, segment)
					case "html":
						t.Imports["html"] = true
						parseHTMLEscapedStringInterpolationSegment(&t.Body, segment)
					default:
						return nil, errors.New("Unknown escape type")
					}

				default:
					parseGoSegment(&t.Body, segment)
				}
				unparsed = unparsed[endGo+2:]
			} else {
				return nil, errors.New("Unable to parse")
			}
		default:
			segment := unparsed
			parseStringSegment(&t.Body, segment)
			unparsed = nil
			break
		}
	}

	return t, nil
}

func parseHTMLEscapedStringInterpolationSegment(buf *bytes.Buffer, segment []byte) {
	buf.WriteString("io.WriteString(writer, html.EscapeString(")
	buf.Write(segment)
	buf.WriteString("))\n")
}

func parseStringSegment(buf *bytes.Buffer, segment []byte) {
	buf.WriteString("io.WriteString(writer, `")
	buf.Write(segment)
	buf.WriteString("`)\n")
}

func parseGoSegment(buf *bytes.Buffer, segment []byte) {
	buf.Write(segment)
	buf.WriteString("\n")
}

func parseStringInterpolationSegment(buf *bytes.Buffer, segment []byte) {
	buf.WriteString("io.WriteString(writer, ")
	buf.Write(segment)
	buf.WriteString(")\n")
}

func parseIntegerInterpolationSegment(buf *bytes.Buffer, segment []byte) {
	buf.WriteString("io.WriteString(writer, strconv.FormatInt(int64(")
	buf.Write(segment)
	buf.WriteString("), 10))\n")
}

func main() {
	output := template.Must(template.New("compiledTemplate").Parse(`
package main

import (
  {{range $key, $val := .Imports}}
  "{{$key}}"
  {{end}}
)

func {{.FuncName}}({{.Parameters}}) (err error) {
  {{.Body}}
  return
}`))

	for _, path := range os.Args[1:] {
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var t *Template
		t, err = parseTemplate(fileBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid template format: %v", err)
			os.Exit(1)
		}

		err = output.Execute(os.Stdout, t)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Template execute error: %v", err)
			os.Exit(1)
		}
	}
}
