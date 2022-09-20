package export

import (
	"context"
	"fmt"
	"os"

	na "github.com/jomei/notionapi"
	"github.com/joshrosso/nexp/config"
)

const (
	notionApiEnvVar = "NOTION_TOKEN"
	defaultFormat   = "markdown"
)

// Render retrieves a Notion Page, renders its Blocks, and returns a []byte
// representation of the contents.
//
// pageID is the UUID that represents the Notion page. If you get a Notion page
// URL, such as
// https://www.notion.so/joshrosso/Climbing-de4d2477f3214ec98614fd46a4e1487f
// then the pageID is de4d2477f3214ec98614fd46a4e1487f. r is the.Renderer,
// which dictates the export format that will be used. opts accepts optional
// additional configuration to pass to the.Renderer for how things should work.
// While opts are variadic, only the first option arguemnt passed will be
// respected.
//
// If there are client issue retrieving the Page, Blocks, or other elements,
// and error is returned.
func (e *exporter) Render(pageID string, opts ...RenderOptions) ([]byte, error) {

	config := resolveRenderConfig(opts...)

	e.page = []byte{}

	p, err := e.c.Page.Get(context.Background(), na.PageID(pageID))
	if err != nil {
		return e.page, fmt.Errorf("Failed getting Notion page (%s), "+
			"error from client: %s", pageID, err)
	}
	e.page = append(e.page, e.Renderer.RenderPageHeader(p, config.Overrides.PageHeader)...)

	e.page, err = e.renderBlocks(pageID, opts...)
	if err != nil {
		return e.page, fmt.Errorf("Failed rendering Notion page, error: %s",
			err)
	}

	// add footer
	e.page = append(e.page, e.Renderer.RenderPageFooter(p, config.Overrides.PageFooter)...)

	return e.page, nil
}

// RenderAppend is the same as Render, except it appends to any existing page
// the exporter has already rendered. See the Render API docs for details on
// arguments and behavior.
func (e *exporter) RenderAppend(pageID string, opts ...RenderOptions) ([]byte, error) {

	// before appending, add separation
	e.page = append(e.page, "\n\n"...)
	return e.renderBlocks(pageID, opts...)
}

// NewRenderer returns a renderer based on the kind (export format) provided.
// An error is returned when no renderer for the kind is known.
func NewRenderer(kind string) (Renderer, error) {
	switch kind {
	case "markdown":
		return &MDRenderer{}, nil
	case "md":
		return &MDRenderer{}, nil
	}

	return nil, fmt.Errorf("No renderer support for type %s", kind)
}

// NewExporter returns an exporter instance with an underlying Notion API
// client attached. The exporter instance is used to call Render functionality.
func NewExporter(opts ...ExporterOptions) (*exporter, error) {
	// set up default render. Will be overwritten if provided in options.
	r, err := NewRenderer(defaultFormat)
	var token string
	var notionClientOpts na.ClientOption

	// TODO(joshrosso): Clean this up into a dedicated options resolver func
	if len(opts) > 0 {
		if opts[0].NotionToken != "" {
			token = opts[0].NotionToken
		}
		if opts[0].ClientOpts != nil {
			notionClientOpts = opts[0].ClientOpts
		}
		if opts[0].Renderer != nil {
			r = opts[0].Renderer
		} else {
			format := opts[0].Format
			if format == "" {
				format = defaultFormat
			}
			r, err = NewRenderer(opts[0].Format)
			if err != nil {
				return nil, err
			}
		}
	}

	// when no token is passed, attempt to resolve via env var or ${HOME}/.config/nexp.yaml
	if token == "" {
		token, err = resolveNotionToken()
		if err != nil {
			return nil, err
		}
	}

	if notionClientOpts == nil {
		return &exporter{c: na.NewClient(na.Token(token)), Renderer: r}, nil
	}

	return &exporter{c: na.NewClient(na.Token(token), notionClientOpts), Renderer: r}, nil
}

// ResolveTitleInPage takes a Notion page object and loops through its
// properties to find the property which is a title Type. It then returns the
// plain text representation of that property.
func ResolveTitleInPage(p *na.Page) string {
	// loops through properties attached to the page to find the property of
	// type title (there can only be one). This is then used as the title for
	// the document.
	var title *na.TitleProperty
	for _, v := range p.Properties {
		if v.GetType() == "title" {
			title = v.(*na.TitleProperty)
		}
	}
	if len(title.Title) < 1 {
		return ""
	}
	return title.Title[0].PlainText

}

// renderBlocks retrieves the blocks that compose a page. It iterates over
// every block retrieved calling appropriate render functionality. As blocks
// are rendered into their string representation, they are appended to the
// []byte stored in the exporter instance. After all blocks are rendered, the
// resulting []byte is returned. If the caller provided any override functiosn
// in OverrideOptions, those are passed and will be respected for the
// appropriate block render(s). An error is returned if there are issues with
// client access to page, blocks, or other objects.
func (e *exporter) renderBlocks(pageID string, opts ...RenderOptions) ([]byte, error) {
	// Retrieve page object to pass to renderer in case render behavior depends
	// on looking up metadata about the page.
	page, err := e.c.Page.Get(context.Background(), na.PageID(pageID))
	if err != nil {
		return e.page, fmt.Errorf("failed to retrieve page from Notion. "+
			"Error: %s.", err)
	}

	config := resolveRenderConfig(opts...)

	blocks, err := e.c.Block.GetChildren(context.Background(),
		na.BlockID(pageID), &na.Pagination{})
	if err != nil {
		return e.page, fmt.Errorf("failed to retrieve data from Notion. "+
			"Error: %s.", err)
	}

	for _, b := range blocks.Results {
		var rend string
		switch b.GetType() {

		case "heading_1":
			in := b.(*na.Heading1Block)
			txt := e.Renderer.RenderText(in.Heading1.RichText)

			rend = e.Renderer.RenderPageHeader1(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Header1)

		case "heading_2":
			in := b.(*na.Heading2Block)
			txt := e.Renderer.RenderText(in.Heading2.RichText)
			rend = e.Renderer.RenderPageHeader2(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Header2)

		case "heading_3":
			in := b.(*na.Heading3Block)
			txt := e.Renderer.RenderText(in.Heading3.RichText)
			rend = e.Renderer.RenderPageHeader3(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Header3)

		case "paragraph":
			in := b.(*na.ParagraphBlock)
			// A blank paragraph block in Notion provides an empty RichText
			// slice. When the SkipEmptyParagraphs option is true, skip this
			// block entirely.
			if config.SkipEmptyParagraphs && len(in.Paragraph.RichText) < 1 {
				continue
			}
			txt := e.Renderer.RenderText(in.Paragraph.RichText)
			rend = e.Renderer.RenderParagraph(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Paragraph)

		case "bulleted_list_item":
			in := b.(*na.BulletedListItemBlock)
			txt := e.Renderer.RenderText(in.BulletedListItem.RichText)
			rend = e.Renderer.RenderBulletedList(&Block{txt, in, opts, config.depth, page},
				config.Overrides.BulletedList)

		case "numbered_list_item":
			in := b.(*na.NumberedListItemBlock)
			txt := e.Renderer.RenderText(in.NumberedListItem.RichText)
			rend = e.Renderer.RenderNumberedList(&Block{txt, in, opts, config.depth, page},
				config.Overrides.NumberedList)

		case "to_do":
			in := b.(*na.ToDoBlock)
			txt := e.Renderer.RenderText(in.ToDo.RichText)
			rend = e.Renderer.RenderTodoList(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Todo)

		case "divider":
			in := b.(*na.DividerBlock)
			rend = e.Renderer.RenderDivider(&Block{BlockRef: in},
				config.Overrides.Divider)

		case "code":
			in := b.(*na.CodeBlock)
			txt := e.Renderer.RenderText(in.Code.RichText)
			rend = e.Renderer.RenderCode(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Code)

		// new table detected. setup table state to support rendering
		// future rows
		case "table":
			config.tableState.tableBlock = b.(*na.TableBlock)
			config.tableState.currentRow = 0

		case "table_row":
			in := b.(*na.TableRowBlock)

			var cells []tableCell
			for i, c := range in.TableRow.Cells {
				var rHeader bool
				var cHeader bool

				if config.tableState.tableBlock.Table.HasRowHeader &&
					config.tableState.currentRow == 0 {
					rHeader = true
				}

				if config.tableState.tableBlock.Table.HasColumnHeader &&
					i == 0 {
					cHeader = true
				}

				tc := tableCell{
					rowTxt:         e.Renderer.RenderText(c),
					isRowHeader:    rHeader,
					isColumnHeader: cHeader,
					tableRef:       config.tableState,
				}
				cells = append(cells, tc)
			}

			rend = e.Renderer.RenderTableRow(cells, config.Overrides.Row)
			// this row has completed rendering, increment current row for
			// future calls.
			config.tableState.currentRow++

		case "quote":
			in := b.(*na.QuoteBlock)
			txt := e.Renderer.RenderText(in.Quote.RichText)
			rend = e.Renderer.RenderQuote(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Quote)

		case "callout":
			in := b.(*na.CalloutBlock)
			txt := e.Renderer.RenderText(in.Callout.RichText)
			rend = e.Renderer.RenderCallout(&Block{txt, in, opts, config.depth, page},
				config.Overrides.Callout)

		case "image":
			// when ignore images is specified, do not send this image block to
			// the renderer and continue to the next block.
			if config.ImageOpts.IgnoreImages {
				continue
			}
			in := b.(*na.ImageBlock)
			rend, err = e.Renderer.RenderImage(&Block{BlockRef: in, Opts: opts, PageRef: page},
				config.Overrides.Image)
			if err != nil {
				return e.page, err
			}
		}

		rend = e.Renderer.AddPadding(&Block{Text: rend, BlockRef: b,
			Depth: config.depth})

		e.page = append(e.page,
			e.Renderer.AddSectionSeperation(config.previousElementType,
				string(b.GetType()))...)

		e.page = append(e.page, rend...)
		config.previousElementType = string(b.GetType())

		// When a child exists, recursively call r.ParseBlocks with the padding
		// value incremented.
		if b.GetHasChildren() {
			configCopy := config
			// when the type is table, it has children (rows) but not with
			// increased depth
			if b.GetType() != "table" {
				configCopy.depth += 1
			}
			e.renderBlocks(string(b.GetID()), configCopy)
		}

	}

	return e.page, nil
}

// resolveNotionToken attempts to find a Notion integration token
// (https://developers.notion.com/docs/authorization). It will prefer a token
// set in the NOTION_TOKEN environment variable. If not present, it looks for
// this token in ${HOME}/.config/nexp.yaml. An error is returned when
// no token is found.
func resolveNotionToken() (string, error) {
	var t string
	t = os.Getenv(notionApiEnvVar)
	if t != "" {
		fmt.Println(t)
		return t, nil
	}

	conf, err := config.LoadNexpConfig()
	if err != nil {
		return t, err
	}
	if conf.Token == "" {
		return t, fmt.Errorf("Token retrieved from configuration was empty")
	}

	return conf.Token, nil
}

// resolveRenderConfig takes a set of RenderOptions and returns the first
// instance. This omits all subsequent instances that are passed.
func resolveRenderConfig(opts ...RenderOptions) RenderOptions {
	var config RenderOptions

	if len(opts) < 1 {
		return config
	}

	return opts[0]
}
