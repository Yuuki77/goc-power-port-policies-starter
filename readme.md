 /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
 package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
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
	devt := devtool.New("https://chromium.googlesource.com/chromiumos/platform/tast-tests")
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

	// Create the Navigate arguments with the optional Referrer field set.
	// navArgs := page.NewNavigateArgs("https://www.google.com").
	// 	SetReferrer("https://duckduckgo.com")
	// nav, err := c.Page.Navigate(ctx, navArgs)
	// if err != nil {
	// 	return err
	// }

	// // Wait until we have a DOMContentEventFired event.
	// if _, err = domContent.Recv(); err != nil {
	// 	return err
	// }

	// fmt.Printf("Page loaded with frame ID: %s\n", nav.FrameID)

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	doc, err := c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	// One suggestion would be to use DOM.querySelector to get a button's nodeId, DOM.getBoxModel to get the position and size of the button, and then Input.dispatchMouseEvent to issue mousedown+mouseup.
	result, err := c.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   doc.Root.NodeID,
		Selector: "body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(1) > a",
	})

	if err != nil {
		return err
	}

	fmt.Printf("HTML: %s\n", result.NodeID)

	// // Get the outer HTML for the page.
	// result, err := c.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
	// 	NodeID: &doc.Root.NodeID,
	// })
	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("HTML: %s\n", result.OuterHTML)

	// // Capture a screenshot of the current page.
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
