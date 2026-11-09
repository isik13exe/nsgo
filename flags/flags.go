package flags

import (
	"fmt"
	"nsgo/tools"
	"os"
	"strconv"
)

type Type int

const (
	tint Type = iota
	tstring
	tbool
)

var longest int
var ToolDesc string
var Help bool

type flag struct {
	affect any
	name   string
	value  string
	typ    Type
	set    bool
	def    any
	desc   string
}

var Args []string
var Flags map[string]*flag = make(map[string]*flag)

func DescTool(desc string) {
	ToolDesc = desc
}

func Affect() tools.Error {
	for _, f := range Flags {
		if f.name == "help" || f.name == "h" {
			continue
		}
		switch f.typ {
		case tint:
			if f.value == "" {
				v := f.affect.(*int)
				*v = f.def.(int)
				continue
			}
			tmp, err := strconv.Atoi(f.value)
			if err != nil {
				return tools.NewErrorf(tools.ErrFlagGeneral, "unvalid value \"%s\" provided for int flag -%s", f.value, f.name)
			}
			v := f.affect.(*int)
			*v = tmp
		case tstring:
			if f.value == "" {
				v := f.affect.(*string)
				*v = f.def.(string)
				continue
			}
			v := f.affect.(*string)
			*v = f.value
		case tbool:
			if f.value == "" {
				v := f.affect.(*bool)
				*v = f.def.(bool)
				continue
			}
			v := f.affect.(*bool)
			if f.value == "true" {
				*v = true
			} else {
				*v = false
			}
		}
	}
	return nil
}

func PrintHelp() {
	if longest == 0 {
		for _, f := range Flags {
			if len(f.name) > longest {
				longest = len(f.name)
			}
		}
	}
	if ToolDesc != "" {
		fmt.Println(ToolDesc)
	}
	for _, f := range Flags {
		if f.name == "help" || f.name == "h" {
			continue
		}
		fmt.Printf("-%s", f.name)
		for i := 0; i < longest-len(f.name)+5; i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("%s\t", f.desc)
		switch f.typ {
		case tbool:
			var def string
			if f.def.(bool) {
				def = "true"
			} else {
				def = "false"
			}
			fmt.Printf(" (default: %s)\n", def)
		case tstring:
			fmt.Printf(" (default: \"%s\")\n", f.def)
		case tint:
			fmt.Printf(" (default: %d)\n", f.def)
		}
	}
	fmt.Println("")
}

func UseHelp() tools.Error {
	if Help {
		PrintHelp()
		return tools.NewError(tools.JustHelp, "help requested")
	}
	return nil
}

func Catch() tools.Error {
	Flags["help"] = &flag{
		affect: &Help,
		name:   "help",
		def:    "false",
		typ:    tbool,
		desc:   "print this help message",
	}
	for _, a := range os.Args[1:] {
		if a[0] == '-' {
			if a[1:] == "help" || a[1:] == "h" {
				Flags["help"].set = true
				Help = true
				continue
			}
			name, value, err := tools.Treatline(a[1:])
			if err != nil {
				_, e := Flags[name]
				if !e {
					return tools.NewErrorf(tools.ErrFlagNotFound, "flag \"%s\" not recognized", name)
				}
				return tools.NewErrorf(tools.ErrFlagGeneral, "issues with flag \"%s\": %w", name, err)
			}
			f, e := Flags[name]
			if !e {
				return tools.NewErrorf(tools.ErrFlagNotFound, "flag \"%s\" not recognized", name)
			}
			if f.set {
				return tools.NewErrorf(tools.ErrFlagAlreadySet, "attempted to reset the flag \"%s\"", f.name)
			}
			f.value = value
		} else {
			Args = append(Args, a)
		}
	}
	return Affect()
}

func StringVar(v *string, name string, def string, desc string) {
	Flags[name] = &flag{
		affect: v,
		name:   name,
		def:    def,
		typ:    tstring,
		desc:   desc,
	}
}

func IntVar(v *int, name string, def int, desc string) {
	Flags[name] = &flag{
		affect: v,
		name:   name,
		def:    def,
		typ:    tint,
		desc:   desc,
	}
}

func BoolVar(v *bool, name string, def bool, desc string) {
	Flags[name] = &flag{
		affect: v,
		name:   name,
		def:    def,
		typ:    tbool,
		desc:   desc,
	}
}
