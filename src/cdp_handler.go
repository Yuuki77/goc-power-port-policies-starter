package src

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/input"
	"github.com/mafredri/cdp/protocol/page"
)

func findAndClickButton(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, selector string) {
	query_reply, err := cdpClient.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   documentReply.Root.NodeID,
		Selector: selector,
	})

	if err != nil {
		log.Fatal(err)
	}

	boxModelReply, err := cdpClient.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
		NodeID: &query_reply.NodeID,
	})

	if err != nil {
		log.Fatal(err)
	}

	x := boxModelReply.Model.Content[0]
	y := boxModelReply.Model.Content[1]
	clickButton(cdpClient, ctx, x, y)
}

func waitForPageReady(cdpClient *cdp.Client, ctx context.Context) *dom.GetDocumentReply {
	domContent, err := cdpClient.Page.DOMContentEventFired(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = cdpClient.Page.Enable(ctx); err != nil {
		log.Fatal(err)
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		os.Exit(1)
	}

	documentReply, err := cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return documentReply
}

func getAttribute(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, selector string) *dom.GetOuterHTMLReply {
	query_reply, err := cdpClient.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   documentReply.Root.NodeID,
		Selector: selector,
	})

	if err != nil {
		log.Fatal(err)
	}

	htmlReply, err := cdpClient.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &query_reply.NodeID,
	})

	if err != nil {
		log.Fatal(err)
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

func setCommitMessage(cdpClient *cdp.Client, ctx context.Context, documentReply *dom.GetDocumentReply, counter *CommitCounter, outputPath string) {
	attributeReply := getAttribute(cdpClient, ctx, documentReply, commitMessageSelector)
	commitMesage := attributeReply.OuterHTML

	tag := "<pre class=\"u-pre u-monospace MetadataMessage\">"
	commitMesage = strings.Replace(commitMesage, tag, "", 1)
	commitMesage = strings.Replace(commitMesage, "</pre>", "", 1)

	commitId := getCommitId(cdpClient, ctx, documentReply)
	createAndWriteFile(commitId, commitMesage, outputPath)
	counter.AddReviewCount(commitMesage)
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
		log.Fatal(err)
	}

	err = c.Input.DispatchMouseEvent(ctx, &input.DispatchMouseEventArgs{
		Type:       "mouseReleased",
		X:          x,
		Y:          y,
		Button:     "left",
		ClickCount: &clickCount,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func navigateToHomePage(cdpClient *cdp.Client, ctx context.Context, domContent page.DOMContentEventFiredClient, url string) *dom.GetDocumentReply {
	// Navigate to chromium.googlesource.com/chromiumos/platform/tast-tests/
	navArgs := page.NewNavigateArgs(url).
		SetReferrer("https://duckduckgo.com")
	nav, err := cdpClient.Page.Navigate(ctx, navArgs)
	if err != nil {
		log.Fatal(err)
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	documentReply, err := cdpClient.DOM.GetDocument(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return documentReply
}
