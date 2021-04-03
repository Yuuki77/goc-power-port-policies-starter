package src

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Args struct {
	Numbers    int
	Branch     string
	Timeout    int
	OutPutPath string
	Url        string
}

func HandleArguments(args []string) Args {
	result := Args{
		Numbers:    DefaultCommitCount,
		Branch:     DefaultBranch,
		Timeout:    DefaultTimeout,
		OutPutPath: DefaultOutPutPath,
		Url:        DefaultUrl,
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

			if arg == "numbers" {
				i, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				result.Numbers = i

			} else if arg == "branch" {
				result.Branch = args[i]

			} else if arg == "timeout" {
				i, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				result.Timeout = i

			} else if arg == "output" {
				result.OutPutPath = args[i]
			} else if arg == "url" {
				result.Url = args[i]
			} else {
				fmt.Println(arg)
				panic("unregistered flag")
			}
		}
	}

	out, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	return result
}
