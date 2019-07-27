package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Template struct {
	GoHeader []byte
	Segments []io.WriterTo
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
	return writeWrapped(w, "io.WriteString(w, `", s, "`)\n")
}

type GoSegment []byte

func (s GoSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeMultiple(w, s, []byte("\n"))
}

type IntegerInterpolationSegment []byte

func (s IntegerInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(w, strconv.FormatInt(int64(", s, "), 10))\n")
}

type RawStringInterpolationSegment []byte

func (s RawStringInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(w, ", s, ")\n")
}

type HTMLEscapedStringInterpolationSegment []byte

func (s HTMLEscapedStringInterpolationSegment) WriteTo(w io.Writer) (n int64, err error) {
	return writeWrapped(w, "io.WriteString(w, html.EscapeString(", s, "))\n")
}

func Parse(templateBytes []byte, escaper func([]byte) io.WriterTo) (*Template, error) {
	t := &Template{}
	regions := bytes.SplitN(templateBytes, []byte("\n---\n"), 2)

	if len(regions) != 2 {
		return nil, errors.New("Did not find divider between header and body.")
	}

	t.GoHeader = bytes.TrimSpace(regions[0])

	err := t.parseBody(regions[1], escaper)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Template) parseBody(body []byte, escaper func([]byte) io.WriterTo) (err error) {
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

			if bytes.HasPrefix(segment, []byte("=i ")) {
				t.Segments = append(t.Segments, IntegerInterpolationSegment(segment[3:]))
			} else if bytes.HasPrefix(segment, []byte("=raw ")) {
				t.Segments = append(t.Segments, RawStringInterpolationSegment(segment[5:]))
			} else if bytes.HasPrefix(segment, []byte("=")) {
				t.Segments = append(t.Segments, escaper(segment[1:]))
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

func (t *Template) WriteTo(w io.Writer) (int64, error) {
	var total int64

	n, err := w.Write(t.GoHeader)
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = fmt.Fprintf(w, " {\n")
	total += int64(n)
	if err != nil {
		return total, err
	}

	for _, s := range t.Segments {
		var n64 int64
		n64, err = s.WriteTo(w)
		total += n64
		if err != nil {
			return total, err
		}
	}

	n, err = fmt.Fprintf(w, "\nreturn nil\n}\n")
	total += int64(n)
	if err != nil {
		return total, err
	}

	return total, nil
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
	var args struct {
		escape string
	}

	flag.StringVar(&args.escape, "escape", "html", "Type of string escaping to perform (options: html, none)")
	flag.Parse()

	var escaper func(buf []byte) io.WriterTo

	switch args.escape {
	case "html":
		escaper = func(buf []byte) io.WriterTo { return HTMLEscapedStringInterpolationSegment(buf) }
	case "none":
		escaper = func(buf []byte) io.WriterTo { return RawStringInterpolationSegment(buf) }
	default:
		fmt.Fprintln(os.Stderr, "unknown escape argument:", args.escape)
		os.Exit(1)
	}

	templateBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	t, err := Parse(templateBytes, escaper)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	t.WriteTo(os.Stdout)

}
