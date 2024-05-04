package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type config struct {
	numTimes int
	htmlPath string
}

var errPosArgsSpecified = errors.New("positional arguments specified")

func parseArgs(w io.Writer, args []string) (config, error) {
	c := config{}
	fs := flag.NewFlagSet("greeter", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.IntVar(&c.numTimes, "n", 0, "Number of times to greet")
	fs.StringVar(&c.htmlPath, "o", "", "Path to HTML file")
	err := fs.Parse(args)
	if err != nil {
		return c, err
	}
	if fs.NArg() != 0 {
		return c, errPosArgsSpecified
	}
	return c, nil
}

func getName(r io.Reader, w io.Writer) (string, error) {
	msg := "Your name please? Press the Enter key when done.\n"
	fmt.Fprint(w, msg)
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	name := scanner.Text()
	if len(name) == 0 {
		return "", errors.New("you didn't enter your name")
	}
	return name, nil
}

func validateArgs(c config) error {
	if !(c.numTimes > 0) {
		return errors.New("must specify a number greater than 0")
	}
	if c.htmlPath != "" {
		if _, err := os.Stat(c.htmlPath); os.IsNotExist(err) {
			if err := os.MkdirAll(c.htmlPath, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func runCmd(r io.Reader, w io.Writer, c config) error {

	name, err := getName(r, w)
	if err != nil {
		return err
	}
	if c.htmlPath == "" {
		greetUserAtStdout(c, name, w)
	} else {
		err := greetUserWithHTML(c, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func greetUserAtStdout(c config, name string, w io.Writer) {
	msg := fmt.Sprintf("Nice to meet you %s\n", name)
	for i := 0; i < c.numTimes; i++ {
		fmt.Fprint(w, msg)
	}
}

func greetUserWithHTML(c config, name string) error {
	element := fmt.Sprintf("<h1>Nice to meet you %s<h1>\n", name)
	f, err := os.Create(c.htmlPath + "/greeting.html")
	if err != nil {
		return err
	}
	defer f.Close()
	for i := 0; i < c.numTimes; i++ {
		_, err := f.WriteString(element)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	c, err := parseArgs(os.Stderr, os.Args[1:])
	if err != nil {
		if errors.Is(err, errPosArgsSpecified) {
			fmt.Fprint(os.Stdout, err)
		}
		os.Exit(1)
	}
	err = validateArgs(c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
	err = runCmd(os.Stdin, os.Stdout, c)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
