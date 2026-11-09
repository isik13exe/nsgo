package tools

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	// "strings"
)

const (
	ErrNotRecognizedCode = iota
	ErrFlagNotFound
	ErrFlagAlreadySet
	ErrFlagGeneral
	ErrFileNotFound
	ErrToolsGeneral
	ErrBuildGeneral
	ErrChartGeneral
	JustHelp
)

type Error interface {
	Error() string
	Code() int
}

type errorImpl struct {
	code int
	msg  string
}

func (e errorImpl) Error() string {
	return e.msg
}

func (e errorImpl) Code() int {
	return e.code
}

func NewError(code int, msg string) Error {
	return errorImpl{code: code, msg: msg}
}

func NewErrorf(code int, format string, args ...any) Error {
	return errorImpl{code: code, msg: fmt.Sprintf(format, args...)}
}

func Count[S ~[]E, E comparable](s S, v E) int {
	c := 0
	for i := range s {
		if v == s[i] {
			c++
		}
	}
	return c
}

func Skipnl(raw string) []string {
	index := slices.Index([]rune(raw), rune('\n'))
	if index == -1 {
		return []string{raw}
	}
	lines := []string{}
	for {
		index := slices.Index([]rune(raw), rune('\n'))
		if index == -1 {
			break
		}
		if len(raw[:index]) > 0 && raw[:index][:2] != "//" {
			lines = append(lines, raw[:index])
		}
		raw = raw[index+1:]
	}
	return lines
}

func Skipnl0(raw string) []string {
	treated := []string{}
	part := []rune{}
	for _, v := range raw {
		if v == '\n' {
			if len(part) > 0 {
				treated = append(treated, string(part))
				part = []rune{}
			}
			continue
		}
		part = append(part, v)
	}
	if len(part) > 0 {
		treated = append(treated, string(part))
	}
	return treated
}

func SkipSpaces(raw string) string {
	treated := []rune{}
	for _, v := range raw {
		if v != ' ' {
			treated = append(treated, v)
		}
	}
	return string(treated)
}

func Treatline(raw string) (string, string, Error) {
	raw = SkipSpaces(raw)
	nc := Count([]rune(raw), rune('='))
	if nc > 1 {
		return "", "", NewErrorf(ErrToolsGeneral, "multiple instances of '='")
	}
	if nc == 0 {
		return raw, "", NewErrorf(ErrToolsGeneral, "no value assigned")
	}
	neq := slices.Index([]rune(raw), rune('='))
	return strings.ToLower(raw[:neq]), raw[neq+1:], nil
}

func parsetype(data string) int {
	if slices.Contains([]rune(data), rune(',')) {
		return 1
	}
	if slices.Contains([]rune(data), rune(':')) {
		return 2
	}
	if IsNumeric(data) {
		return 3
	}
	return 0
}

func ParseTest(data string) ([]int, Error) {
	typ := parsetype(data)
	switch typ {
	case 1:
		return parseComma(data)
	case 2:
		return parseColon(data)
	case 3:
		return parseNumeric(data)
	default:
		return nil, NewErrorf(ErrToolsGeneral, "invalid type %d", typ)
	}
}

func parseComma(data string) ([]int, Error) {

	parts := strings.Split(data, ",")
	result := []int{}
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, NewErrorf(ErrToolsGeneral, "invalid number: %w", err)
		}
		result = append(result, num)
	}
	return result, nil
}

func parseNumeric(data string) ([]int, Error) {
	num, err := strconv.Atoi(data)
	if err != nil {
		return nil, NewErrorf(ErrToolsGeneral, "invalid number: %w", err)
	}
	return []int{num}, nil
}

func parseColon(data string) ([]int, Error) {
	parts := strings.Split(data, ":")

	result := []int{}
	min, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, NewErrorf(ErrToolsGeneral, "invalid min: %w", err)
	}
	max, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, NewErrorf(ErrToolsGeneral, "invalid max: %w", err)
	}
	for i := min; i <= max; i++ {
		result = append(result, i)
	}
	return result, nil
}

func Mkdir(path string) Error {
	if path == "" {
		return nil
	}
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return NewErrorf(ErrToolsGeneral, "could not create directory: %w", err)
	}
	return nil
}

func PurgeDir(path string) Error {
	if path == "" {
		return nil
	}
	err := os.RemoveAll(path)
	if err != nil {
		return NewErrorf(ErrToolsGeneral, "could not purge directory: %w", err)
	}
	return nil
}

func IsNumeric(data string) bool {
	_, err := strconv.Atoi(data)
	return err == nil
}
