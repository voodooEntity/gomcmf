package cli

import (
	"flag"
	"github.com/voodooEntity/gomcmf/src/config"
	"github.com/voodooEntity/gomcmf/src/core"
	"github.com/voodooEntity/gomcmf/src/types"
	"github.com/voodooEntity/gomcmf/src/util"
	"log"
	"os"
)

var loggerOut = log.New(os.Stdout, "", 0)

func Init() {
	// first we gonne parse the args
	args := parseArgs()

	app := core.Core{
		Command:  args.Command,
		Verbose:  args.Verbose,
		Name:     args.Name,
		Sequence: args.Sequence,
		Type:     args.Type,
		Target:   args.Target,
		Input:    "",
		Pwd:      args.Pwd + "/",
	}

	// dispatch command
	switch command := args.Command; command {
	case "init":
		app.CreateDefaultProject()
	case "create":
		if "" == args.Name {
			util.Error("Missing argument '-name'")
		}
		app.CreatePage()
	case "build":
		config.Init()
		// build all template contents
		app.BuildProject()
	case "move":
		// change sequence of given page
	case "delete":
		// delete given page and resort other sequences
	default:
		// unknown command given, printing help instead
		loggerOut.Println("Unknown command given: '", command, "'")
		printHelpText()
	}
}

func parseArgs() types.Args {
	// first we check for the help flag
	if 1 < len(os.Args) {
		if ok := os.Args[1]; ok == "help" {
			printHelpText()
			os.Exit(1)
		}
	}

	// input by string
	var command string
	flag.StringVar(&command, "command", "", "-command somecommand")

	// input by string
	verbose := flag.Bool("verbose", false, "-verbose")

	// input by string
	var name string
	flag.StringVar(&name, "name", "", "-value somevalue")

	// input by string
	var ctype string
	flag.StringVar(&ctype, "type", "md", "-type contenttype")

	var sequence int
	flag.IntVar(&sequence, "sequence", -1, "-sequence intSequence")

	var target string
	flag.StringVar(&target, "target", "./", "-target /target/directory/to/build/into")

	// parse the flags
	flag.Parse()

	wdir, err := os.Getwd()
	if nil != err {
		util.Error("Could not get current working directory with error '" + err.Error() + "'")
	}

	return types.Args{
		Command:  command,
		Verbose:  *verbose,
		Name:     name,
		Sequence: sequence,
		Type:     ctype,
		Target:   target,
		Input:    "",
		Pwd:      wdir,
	}

}

func printHelpText() {
	helpText := "Gomcmf help:\n" +
		" todo :>"
	loggerOut.Println(helpText)
}
