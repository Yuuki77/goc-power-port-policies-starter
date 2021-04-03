package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/input"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"github.com/yukiOsaki/goc-power-port-policies/src"
)

var outputPath = "./output/"
var commitCount = 1
var maxCount = src.DefaultCommitCount
var branch = src.DefaultBranch
var url = src.DefaultUrl

func main() {
	args := src.HandleArguments(os.Args)
	maxCount = args.Numbers
	outputPath = "./" + args.OutPutPath + "/"
	branch = args.Branch
	url = args.Url

	// return
	err := run(time.Duration(args.Timeout) * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func run(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

	// Navigate to chromium.googlesource.com/chromiumos/platform/tast-tests/
	navArgs := page.NewNavigateArgs(url).
		SetReferrer("https://duckduckgo.com")
	nav, err := cdpClient.Page.Navigate(ctx, navArgs)
	if err != nil {
		return err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		return err
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	documentReply, err := cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	selector := ""
	if branch == "main" {
		selector = "body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(1) > a"
		findAndClickButton(cdpClient, ctx, documentReply, selector)
	} else {
		navArgs := page.NewNavigateArgs(url + "/+/refs/heads/" + branch)
		nav, err = cdpClient.Page.Navigate(ctx, navArgs)
		if err != nil {
			return err
		}
	}

	for i := 1; i < maxCount; i++ {
		documentReply = waitForPageReady(cdpClient, ctx)
		selector = "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)"
		attributeReply := getAttribute(cdpClient, ctx, documentReply, selector)

		commitId := remove_tag(attributeReply.OuterHTML, "td")
		selector = "body > div > div > pre"
		attributeReply = getAttribute(cdpClient, ctx, documentReply, selector)
		commitMesage := attributeReply.OuterHTML

		tag := "<pre class=\"u-pre u-monospace MetadataMessage\">"
		commitMesage = strings.Replace(commitMesage, tag, "", 1)
		commitMesage = strings.Replace(commitMesage, "</pre>", "", 1)

		fmt.Println("commitMessage", commitMesage)

		createAndWriteFile(commitId, commitMesage)

		// go to next commit
		commitCount++
		selector = "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(5) > td > a"
		findAndClickButton(cdpClient, ctx, documentReply, selector)
		fmt.Println("Finish", commitCount)
	}
	return nil
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

func createAndWriteFile(file_name string, content string) {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.Mkdir(outputPath, 0755)
	}

	f, err := os.Create(outputPath + file_name)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(content)

	if err2 != nil {
		log.Fatal(err2)
	}
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
