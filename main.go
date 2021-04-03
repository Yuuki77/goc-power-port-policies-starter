package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/input"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

func main() {
	err := run(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
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

	c := cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		return err
	}

	// Navigate to chromium.googlesource.com/chromiumos/platform/tast-tests/
	navArgs := page.NewNavigateArgs("https://chromium.googlesource.com/chromiumos/platform/tast-tests/").
		SetReferrer("https://duckduckgo.com")
	nav, err := c.Page.Navigate(ctx, navArgs)
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
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	// Click main under branches on the left, it navigates you to the latest commit
	query_reply, err := c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   doc.Root.NodeID,
		Selector: "body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(1) > a",
	})

	if err != nil {
		return err
	}

	boxModelReply, err := c.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
		NodeID: &query_reply.NodeID,
	})

	if err != nil {
		return err
	}

	// click main branch link tag
	x := boxModelReply.Model.Content[0]
	y := boxModelReply.Model.Content[1]
	clickButton(c, ctx, x, y)

	// Parse page and write only commit message to the file, each file should have a unique name

	// get commit id for file title
	// body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)

	// Get the outer HTML for the page.
	doc, err = c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("HTML: %s\n", result.OuterHTML)

	query_reply, err = c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   doc.Root.NodeID,
		Selector: "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > th",
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(query_reply.NodeID)

	// NodeID: &query_reply.NodeID,
	// boxModelReply, err := c.DOM.GetBoxModel(ctx, &dom.GetBoxModelArgs{
	// 	NodeID: &query_reply.NodeID,
	// })

	// attributeReply, err := c.DOM.GetAttributes(ctx, &dom.GetAttributesArgs{
	// 	NodeID: query_reply.NodeID,
	// })

	// if err != nil {
	// 	fmt.Println(err)

	// 	panic(err)
	// }

	// fmt.Println("commit id", attributeReply)
	// getAttributes

	// Open a DOMContentEventFired client to buffer this event.
	// domContent, err = c.Page.DOMContentEventFired(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer domContent.Close()

	// if _, err = domContent.Recv(); err != nil {
	// 	return err
	// }

	// getCommitId(c, ctx, doc)

	// doc = html.Parse(strings.NewReader(result.OuterHTML))

	// Capture a screenshot of the current page.
	// screenshotName := "screenshot.jpg"
	// screenshotArgs := page.NewCaptureScreenshotArgs().
	// 	SetFormat("jpeg").
	// 	SetQuality(80)
	// screenshot, err := c.Page.CaptureScreenshot(ctx, screenshotArgs)
	// if err != nil {
	// 	return err
	// }
	// if err = ioutil.WriteFile(screenshotName, screenshot.Data, 0644); err != nil {
	// 	return err
	// }

	// fmt.Printf("Saved screenshot: %s\n", screenshotName)

	// pdfName := "page.pdf"
	// f, err := os.Create(pdfName)
	// if err != nil {
	// 	return err
	// }

	// pdfArgs := page.NewPrintToPDFArgs().
	// 	SetTransferMode("ReturnAsStream") // Request stream.
	// pdfData, err := c.Page.PrintToPDF(ctx, pdfArgs)
	// if err != nil {
	// 	return err
	// }

	// sr := c.NewIOStreamReader(ctx, *pdfData.Stream)
	// r := bufio.NewReader(sr)

	// // Write to file in ~r.Size() chunks.
	// _, err = r.WriteTo(f)
	// if err != nil {
	// 	return err
	// }

	// err = f.Close()
	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("Saved PDF: %s\n", pdfName)

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

func waitDomContent() {
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()
}

// func Body(doc *html.Node) (*html.Node, error) {
// 	var body *html.Node
// 	var crawler func(*html.Node)
// 	crawler = func(node *html.Node) {
// 		if node.Type == html.ElementNode && node.Data == "body" {
// 			body = node
// 			return
// 		}
// 		for child := node.FirstChild; child != nil; child = child.NextSibling {
// 			crawler(child)
// 		}
// 	}
// 	crawler(doc)
// 	if body != nil {
// 		return body, nil
// 	}
// 	return nil, errors.New("Missing <body> in the node tree")
// }

func getCommitId(c *cdp.Client, ctx context.Context, doc *dom.GetDocumentReply) string {
	fmt.Println(doc.Root.NodeID)

	query_reply, err := c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   doc.Root.NodeID,
		Selector: "body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)",
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(query_reply)

	attribute_rely, err := c.DOM.GetAttributes(ctx, &dom.GetAttributesArgs{
		NodeID: query_reply.NodeID,
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println("commit id", attribute_rely)
	// getAttributes

	return ""
}

func setup() {

}
