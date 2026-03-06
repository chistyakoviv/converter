package stack_parser

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

type simpleStackParser struct {
}

func NewSimpleStackParser() StackParser {
	return &simpleStackParser{}
}

func (s *simpleStackParser) Parse(debugStack []byte, rvr any) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}

	buf.WriteString("\n")
	buf.WriteString(" panic: ")
	fmt.Fprintf(buf, "%v", rvr)
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

func (s *simpleStackParser) decorateLine(line string, num int) (string, error) {
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

func (s *simpleStackParser) decorateFuncCallLine(line string, num int) (string, error) {
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

	if num == 0 {
		buf.WriteString(" -> ")
	} else {
		buf.WriteString("    ")
	}
	buf.WriteString(pkg)
	buf.WriteString(method)
	buf.WriteString("\n")
	// buf.WriteString(addr)
	return buf.String(), nil
}

func (s *simpleStackParser) decorateSourceLine(line string, num int) (string, error) {
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

	if num == 1 {
		buf.WriteString(" ->   ")
	} else {
		buf.WriteString("      ")
	}
	buf.WriteString(dir)
	buf.WriteString(file)
	buf.WriteString(lineno)
	if num == 1 {
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	return buf.String(), nil
}
