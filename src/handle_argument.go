package src

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Args struct {
	Numbers       int
	Branch        string
	Timeout       int
	OutPutPath    string
	CsvOutputPath string
	Url           string
}

func HandleArguments(args []string) Args {
	result := Args{
		Numbers:       DefaultCommitCount,
		Branch:        DefaultBranch,
		Timeout:       DefaultTimeout,
		OutPutPath:    DefaultOutPutPath,
		CsvOutputPath: DefaultCsvOutPutPath,
		Url:           DefaultUrl,
	}

	if len(args) < 1 {
		return result
	}

	args = args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if i%2 == 0 {
			arg := strings.Replace(arg, "--", "", 1)
			i++

			switch arg {
			case "numbers":
				i, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				result.Numbers = i
			case "branch":
				result.Branch = args[i]
			case "timeout":
				i, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				result.Timeout = i
			case "output":
				result.OutPutPath = args[i]
			case "url":
				result.Url = args[i]
			case "csv-output":
				result.CsvOutputPath = args[i]
			default:
				panic("unregistered flag")
			}
		}
	}

	if debug {
		out, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	}

	return result
}
