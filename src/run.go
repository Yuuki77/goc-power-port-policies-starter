package src

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

type Args struct {
	Numbers       int
	Branch        string
	Timeout       int
	OutPutPath    string
	CsvOutputPath string
	Url           string
}

func Run() error {
	args := Args{}

	flag.IntVar(&args.Numbers, "numbers", DefaultCommitCount, "Number of last commit messages")
	flag.StringVar(&args.Url, "url", DefaultUrl, "Repository URL")
	flag.StringVar(&args.Branch, "branch", DefaultBranch, "Branch Name")
	flag.IntVar(&args.Timeout, "timeout", DefaultTimeout, "Time out")
	flag.StringVar(&args.OutPutPath, "output-dir", DefaultOutPutPath, "Commit folder path")
	flag.StringVar(&args.CsvOutputPath, "csv-dir", DefaultCsvOutPutPath, "CSV folder path")
	flag.Parse()

	maxCount := args.Numbers
	outputPath := "./" + args.OutPutPath + "/"
	branch := args.Branch
	url := args.Url
	csvOutputPath := args.CsvOutputPath

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.Timeout)*time.Second)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return err
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	cdpClient := cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := cdpClient.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = cdpClient.Page.Enable(ctx); err != nil {
		return err
	}

	documentReply := navigateToHomePage(cdpClient, ctx, domContent, url)

	if branch == "main" {
		findAndClickButton(cdpClient, ctx, documentReply, mainBranchLinkSelector)
	} else {
		navArgs := page.NewNavigateArgs(url + "/+/refs/heads/" + branch)
		_, err = cdpClient.Page.Navigate(ctx, navArgs)
		if err != nil {
			return err
		}
	}

	counter := CommitCounter{
		CommiterCounterMap: map[string]int{},
		ReviewerCounterMap: map[string]int{},
	}

	for i := 0; i < maxCount; i++ {
		documentReply = waitForPageReady(cdpClient, ctx)
		setAuthorId(cdpClient, ctx, documentReply, &counter)
		setCommitMessage(cdpClient, ctx, documentReply, &counter, outputPath)

		// go to next commit
		findAndClickButton(cdpClient, ctx, documentReply, parentLinkSelector)
		fmt.Println("current:", i+1)
	}

	GenerateCsv(csvOutputPath, &counter, csvOutputPath)
	return nil
}

func remove_tag(original_string string, tag string) string {
	replaced := strings.Replace(original_string, "<"+tag+">", "", 1)
	return strings.Replace(replaced, "</"+tag+">", "", 1)
}
