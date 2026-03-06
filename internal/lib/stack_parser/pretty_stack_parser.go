package stack_parser

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/fatih/color"
)

type prettyStackParser struct {
	red     *color.Color
	cyan    *color.Color
	blue    *color.Color
	white   *color.Color
	magenta *color.Color
	yellow  *color.Color
	green   *color.Color
	black   *color.Color
}

func NewPrettyStackParser() StackParser {
	return &prettyStackParser{
		red:     color.New(color.FgRed),
		cyan:    color.New(color.FgCyan),
		blue:    color.New(color.FgBlue),
		white:   color.New(color.FgWhite),
		magenta: color.New(color.FgMagenta),
		yellow:  color.New(color.FgYellow),
		green:   color.New(color.FgGreen),
		black:   color.New(color.FgBlack),
	}
}

func (s *prettyStackParser) Parse(debugStack []byte, rvr any) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}

	buf.WriteString("\n")
	_, err = s.cyan.Fprint(buf, " panic: ")
	if err != nil {
		return nil, err
	}
	_, err = s.blue.Fprintf(buf, "%v", rvr)
	if err != nil {
		return nil, err
	}
	buf.WriteString("\n\n")

	// process debug stack info
	stack := strings.Split(string(debugStack), "\n")
	lines := []string{}

	// locate panic line, as we may have nested panics
	for i := len(stack) - 1; i > 0; i-- {
		lines = append(lines, stack[i])
		if strings.HasPrefix(stack[i], "panic(") {
			lines = lines[0 : len(lines)-2] // remove boilerplate
			break
		}
	}

	// reverse
	slices.Reverse(lines)

	// decorate
	for i, line := range lines {
		lines[i], err = s.decorateLine(line, i)
		if err != nil {
			return nil, err
		}
	}

	for _, l := range lines {
		fmt.Fprintf(buf, "%s", l)
	}
	return buf.Bytes(), nil
}

func (s *prettyStackParser) decorateLine(line string, num int) (string, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "\t") || strings.Contains(line, ".go:") {
		return s.decorateSourceLine(line, num)
	}
	if strings.HasSuffix(line, ")") {
		return s.decorateFuncCallLine(line, num)
	}
	if strings.HasPrefix(line, "\t") {
		return strings.Replace(line, "\t", "      ", 1), nil
	}
	return fmt.Sprintf("    %s\n", line), nil
}

func (s *prettyStackParser) decorateFuncCallLine(line string, num int) (string, error) {
	var err error
	idx := strings.LastIndex(line, "(")
	if idx < 0 {
		return "", errors.New("not a func call line")
	}

	buf := &bytes.Buffer{}
	pkg := line[0:idx]
	// addr := line[idx:]
	method := ""

	if idx := strings.LastIndex(pkg, string(os.PathSeparator)); idx < 0 {
		if idx := strings.Index(pkg, "."); idx > 0 {
			method = pkg[idx:]
			pkg = pkg[0:idx]
		}
	} else {
		method = pkg[idx+1:]
		pkg = pkg[0 : idx+1]
		if idx := strings.Index(method, "."); idx > 0 {
			pkg += method[0:idx]
			method = method[idx:]
		}
	}
	pkgColor := s.yellow
	methodColor := s.green

	if num == 0 {
		_, err = s.red.Fprint(buf, " -> ")
		if err != nil {
			return "", err
		}
		pkgColor = s.magenta
		methodColor = s.red
	} else {
		_, err = s.white.Fprint(buf, "    ")
		if err != nil {
			return "", err
		}
	}
	_, err = pkgColor.Fprint(buf, pkg)
	if err != nil {
		return "", err
	}
	_, err = methodColor.Fprint(buf, method)
	if err != nil {
		return "", err
	}
	buf.WriteString("\n")
	// s.black.Fprint(buf, addr)
	return buf.String(), nil
}

func (s *prettyStackParser) decorateSourceLine(line string, num int) (string, error) {
	var err error
	idx := strings.LastIndex(line, ".go:")
	if idx < 0 {
		return "", errors.New("not a source line")
	}

	buf := &bytes.Buffer{}
	path := line[0 : idx+3]
	lineno := line[idx+3:]

	idx = strings.LastIndex(path, string(os.PathSeparator))
	dir := path[0 : idx+1]
	file := path[idx+1:]

	idx = strings.Index(lineno, " ")
	if idx > 0 {
		lineno = lineno[0:idx]
	}
	fileColor := s.cyan
	lineColor := s.green

	if num == 1 {
		_, err = s.red.Fprint(buf, " ->   ")
		if err != nil {
			return "", err
		}
		fileColor = s.red
		lineColor = s.magenta
	} else {
		buf.WriteString("      ")
	}
	_, err = s.white.Fprint(buf, dir)
	if err != nil {
		return "", err
	}
	_, err = fileColor.Fprint(buf, file)
	if err != nil {
		return "", err
	}
	_, err = lineColor.Fprint(buf, lineno)
	if err != nil {
		return "", err
	}
	if num == 1 {
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	return buf.String(), nil
}
