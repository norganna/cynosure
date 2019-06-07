package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type stringMap map[string]bool

func (i *stringMap) String() string {
	var s []string
	for k := range *i {
		s = append(s, k)
	}
	return strings.Join(s, ", ")
}

func (i *stringMap) Set(value string) error {
	(*i)[strings.ToLower(value)] = true
	return nil
}

// Reads all flagged files in the current folder and encodes them as strings literals in const.json.go
func main() {
	reBackticks := regexp.MustCompile("(`+(?:[\\w\\s.,:;_+=*#$%&!?@^`'/~[<({})>\\]-]+`)?)")

	outFile := ""
	strip := stringMap{}
	flag.Var(&strip, "strip", "Segment to remove from filenames (repeat for multiple)")
	flag.StringVar(&outFile, "out", outFile, "Filename to output (defaults to PKG.pb.cx.go)")
	flag.Parse()
	args := flag.Args()
	pkg := args[0]

	if outFile == "" {
		outFile = pkg + ".pb.cx.go"
	}

	out, _ := os.Create(outFile)

	var fs []string
	found := map[string]bool{}
	for _, pattern := range args[1:] {
		ff, err := filepath.Glob(pattern)
		if err == nil {
			for _, f := range ff {
				if !found[f] {
					fs = append(fs, f)
					found[f] = true
				}
			}
		}
	}

	_, err := out.WriteString("package " + pkg + "\n\nconst (\n")
	if err != nil {
		fmt.Println("Failure: ", err)
		os.Exit(1)
	}
	for _, f := range fs {
		var n []string
		for _, c := range strings.Split(f, ".") {
			if c != "" {
				c = strings.ToLower(c)
				if !strip[c] {
					switch c {
					case "id":
						c = "ID"
					case "ids":
						c = "IDs"
					case "url":
						c = "URL"
					case "urls":
						c = "URLs"
					default:
						c = strings.ToUpper(c[0:1]) + c[1:]
					}
					n = append(n, c)
				}
			}
		}
		name := strings.Join(n, "")
		_, err = out.WriteString(name + " = `")
		if err != nil {
			fmt.Println("Failure: ", err)
			os.Exit(1)
		}
		data, _ := ioutil.ReadFile(f)
		data = reBackticks.ReplaceAll(data, []byte("` + \"$1\" + `"))
		_, err = out.Write(data)
		if err != nil {
			fmt.Println("Failure: ", err)
			os.Exit(1)
		}
		_, err = out.WriteString("`\n")
		if err != nil {
			fmt.Println("Failure: ", err)
			os.Exit(1)
		}
	}
	_, err = out.WriteString(")\n")
	if err != nil {
		fmt.Println("Failure: ", err)
		os.Exit(1)
	}
}
