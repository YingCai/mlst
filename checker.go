package mlst

import (
	"flag"
	"fmt"
	"os"
)

var args []string

func printError(err string) {
	fmt.Fprintf(os.Stderr, err)
}

func printMessage(msg string) {
	fmt.Fprintf(os.Stdout, msg)
}

func init() {
	flag.Parse()
	args = flag.Args()
}

func CheckInput() []EdgeSet {
	var infile string

	if len(args) < 1 {
		infile = DefaultInputFile
	} else {
		infile = args[0]
	}

	file, err := os.Open(infile)
	if err != nil {
		printError(fmt.Sprintf("Cannot open '%s' (%s).\n", infile, err.Error()))
		return nil
	}
	defer file.Close()

	inReader := NewInFileReader(file)
	if edgeSets, err := inReader.ReadInputFile(); err != nil {
		printError("(" + infile + ") " + err.Error() + "\n")
	} else {
		printMessage(fmt.Sprintf("Input file '%s' has the correct format.\n", infile))
		return edgeSets
	}
	return nil
}

func CheckOutput(checkOutputProgramName string) {
	var outfile string

	switch numArgs := len(args); numArgs {
	case 0:
		outfile = DefaultOutputFile
	case 2:
		outfile = args[1]
	default:
		printError(fmt.Sprintf("usage: %s [file.in file.out]\n\n"+
			"  Check the format of \"file.out\" against \"file.in\".\n"+
			"Error: Must provide either two arguments, or zero arguments to use the default\n"+
			"input \"%s\" and output \"%s\". (Number of arguments is %d)\n",
			checkOutputProgramName, DefaultInputFile, DefaultOutputFile, numArgs))
		return
	}

	edgeSets := CheckInput()
	if edgeSets == nil {
		return
	}

	file, err := os.Open(outfile)
	if err != nil {
		printError(fmt.Sprintf("Cannot open '%s' (%s).\n", outfile, err.Error()))
		return
	}
	defer file.Close()

	outReader := NewOutFileReader(file)
	if NumLeaves, err := outReader.ReadOutputFile(edgeSets); err != nil {
		printError("(" + outfile + ") " + err.Error() + "\n")
	} else {
		printMessage(fmt.Sprintf("Output file '%s' has the correct format.\n", outfile))
		for i, v := range NumLeaves {
			printMessage(fmt.Sprintf("Output tree %d has %d leaves.\n", i+1, v))
		}
	}
}
