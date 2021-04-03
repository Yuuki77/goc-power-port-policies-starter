package src

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/input"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

var outputPath = DefaultOutPutPath
var commitCount = 0
var maxCount = DefaultCommitCount
var branch = DefaultBranch
var url = DefaultUrl
var csvOutputPath = DefaultCsvOutPutPath

func Run(flags []string) error {
	args := HandleArguments(flags)
	maxCount = args.Numbers
	outputPath = "./" + args.OutPutPath + "/"
	branch = args.Branch
	url = args.Url
	csvOutputPath = args.CsvOutputPath

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

	documentReply := navigateToHomePage(cdpClient, ctx, domContent)

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
		setCommitMessage(cdpClient, ctx, documentReply, &counter)

		// go to next commit
		commitCount++
		findAndClickButton(cdpClient, ctx, documentReply, parentLinkSelector)
		fmt.Println("current:", commitCount)
	}

	GenerateCsv(csvOutputPath, &counter)
	return nil
}

func navigateToHomePage(cdpClient *cdp.Client, ctx context.Context, domContent page.DOMContentEventFiredClient) *dom.GetDocumentReply {
	// Navigate to chromium.googlesource.com/chromiumos/platform/tast-tests/
	navArgs := page.NewNavigateArgs(url).
		SetReferrer("https://duckduckgo.com")
	nav, err := cdpClient.Page.Navigate(ctx, navArgs)
	if err != nil {
		panic(err)
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		panic(err)
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	documentReply, err := cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		panic(err)
	}

	return documentReply
}

func clickButton(c *cdp.Client, ctx context.Context, x float64, y float64) {
	clickCount := 1
	err := c.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mousePressed",
		X:          x,
		Y:          y,
		Button:     "left",
		ClickCount: &clickCount,
	})

	if err != nil {
		panic(err)
	}

	err = c.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mouseReleased",
		X:          x,
		Y:          y,
		Button:     "left",
		ClickCount: &clickCount,
	})

	if err != nil {
		panic(err)
	}
}

func remove_tag(original_string string, tag string) string {
	replaced := strings.Replace(original_string, "<"+tag+">", "", 1)
	return strings.Replace(replaced, "</"+tag+">", "", 1)
}

func findAndClickButton(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, selector string) {
	query_reply, err := cdpClient.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   documentReply.Root.NodeID,
		Selector: selector,
	})

	if err != nil {
		panic(err)
	}

	boxModelReply, err := cdpClient.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
		NodeID: &query_reply.NodeID,
	})

	if err != nil {
		panic(err)
	}

	x := boxModelReply.Model.Content[0]
	y := boxModelReply.Model.Content[1]
	clickButton(cdpClient, ctx, x, y)
}

func waitForPageReady(cdpClient *cdp.Client, ctx context.Context) *dom.GetDocumentReply {
	domContent, err := cdpClient.Page.DOMContentEventFired(ctx)
	if err != nil {
		panic(err)
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = cdpClient.Page.Enable(ctx); err != nil {
		panic(err)
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		os.Exit(1)
	}

	documentReply, err := cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		panic(err)
	}

	return documentReply
}

func getAttribute(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, selector string) *dom.GetOuterHTMLReply {
	query_reply, err := cdpClient.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   documentReply.Root.NodeID,
		Selector: selector,
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	htmlReply, err := cdpClient.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &query_reply.NodeID,
	})

	if err != nil {
		fmt.Println(err)

		panic(err)
	}

	return htmlReply
}

func getCommitId(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply) string {
	attributeReply := getAttribute(cdpClient, ctx, documentReply, commitIdSelector)
	return remove_tag(attributeReply.OuterHTML, "td")
}

func setAuthorId(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, counter *CommitCounter) {
	attributeReply := getAttribute(cdpClient, ctx, documentReply, authorSelector)
	author := remove_tag(attributeReply.OuterHTML, "td")

	counter.AddCommitCount(author)
}

func setCommitMessage(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, counter *CommitCounter) {
	attributeReply := getAttribute(cdpClient, ctx, documentReply, commitMessageSelector)
	commitMesage := attributeReply.OuterHTML

	tag := "<pre class=\"u-pre u-monospace MetadataMessage\">"
	commitMesage = strings.Replace(commitMesage, tag, "", 1)
	commitMesage = strings.Replace(commitMesage, "</pre>", "", 1)

	commitId := getCommitId(cdpClient, ctx, documentReply)
	createAndWriteFile(commitId, commitMesage)
	counter.AddReviewCount(commitMesage)
}
