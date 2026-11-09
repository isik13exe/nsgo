package msg

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

var IsTerm bool = true

type Spliter struct {
	Err   string
	Wr    io.Writer
	Extra io.Writer
}

type Color struct {
	R uint8
	G uint8
	B uint8
}

func (c Color) String() string {
	return Rgb(c.R, c.G, c.B)
}

var (
	Red   = Color{255, 10, 10}
	Blue  = Color{10, 10, 255}
	Green = Color{10, 255, 10}
)

func (e *Spliter) Write(p []byte) (n int, err error) {
	e.Err = string(p)
	e.Wr.Write(p)
	if e.Extra != nil {
		e.Extra.Write(p)
	}
	return len(p), nil
}

func ErrorSplitter() *Spliter {
	return &Spliter{
		Wr: os.Stderr,
	}
}

func FileSplitter(filename string) *Spliter {
	if filename != "" {
		f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		return &Spliter{
			Wr:    os.Stdout,
			Extra: f,
		}
	}
	return &Spliter{
		Wr: os.Stdout,
	}
}

func Check() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		IsTerm = false
	}
}

func Rgb(r, g, b uint8) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func Reset() string {
	return "\x1b[0m"
}

func Reds(s string) string {
	return Red.String() + s + Reset()
}

func Blues(s string) string {
	return Blue.String() + s + Reset()
}

func Greens(s string) string {
	return Green.String() + s + Reset()
}

func ERROR(err error) {
	if IsTerm {
		fmt.Printf("%s %v", Reds("ERROR:"), err)
	} else {
		fmt.Printf("ERROR: %v", err)
	}
}

func ERRORf(format string, data ...any) {
	if IsTerm {
		fmt.Printf(Reds("ERROR: ")+format, data...)
	} else {
		fmt.Printf("ERROR: "+format, data...)
	}
}

func ERRORln(data ...any) {
	holder := fmt.Sprintln(data...)
	if IsTerm {
		fmt.Printf(Reds("ERROR: ")+"%s", holder)
	} else {
		fmt.Printf("ERROR: "+"%s", holder)
	}
}

func IsERROR(err error) error {
	if err != nil {
		ERRORln(err)
	}
	return err
}

func SUCCESS(err error) {
	if IsTerm {
		fmt.Printf("%s %v\n", Greens("SUCCESS:"), err)
	} else {
		fmt.Printf("SUCCESS: %v\n", err)
	}
}

func SUCCESSf(format string, data ...any) {
	if IsTerm {
		fmt.Printf(Greens("SUCCESS: ")+format, data...)
	} else {
		fmt.Printf("SUCCESS: "+format, data...)
	}
}

func SUCCESSln(data ...any) {
	holder := fmt.Sprintln(data...)
	if IsTerm {
		fmt.Printf(Greens("SUCCESS: ")+"%s", holder)
	} else {
		fmt.Printf("SUCCESS: "+"%s", holder)
	}
}

func INFO(a any) {
	if IsTerm {
		fmt.Printf("%s %v\n", Blues("INFO:"), a)
	} else {
		fmt.Printf("INFO: %v\n", a)
	}
}

func INFOf(format string, data ...any) {
	if IsTerm {
		fmt.Printf(Blues("INFO: ")+format, data...)
	} else {
		fmt.Printf("INFO: "+format, data...)
	}
}

func INFOln(data ...any) {
	holder := fmt.Sprintln(data...)
	if IsTerm {
		fmt.Printf(Blues("INFO: ")+"%s", holder)
	} else {
		fmt.Printf("INFO: "+"%s", holder)
	}
}

func Colorlitln(r uint8, g uint8, b uint8, data ...any) {
	holder := fmt.Sprintln(data...)
	if IsTerm {
		fmt.Printf(Rgb(r, g, b)+"%s", holder)
	} else {
		fmt.Printf("%s", holder)
	}
}

func Colorln(c Color, data ...any) {
	holder := fmt.Sprintln(data...)
	if IsTerm {
		fmt.Printf(c.String()+"%s", holder)
	} else {
		fmt.Printf("%s", holder)
	}
}
