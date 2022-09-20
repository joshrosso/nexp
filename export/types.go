package export

import (
	na "github.com/jomei/notionapi"
)

// Renderer defines how to translate Notion API block objects
// (https://developers.notion.com/reference/block) into a text representation.
// By implementing this interface, you'll be able to export Notion pages to
// whatever custom format you prefer.
//
// From the Notion GUI, most of these RenderX methods map to Notion's Basic
// Blocks (https://www.notion.so/help/writing-and-editing-basics#basic-blocks)
// Renderer methods return a string representation of that block, which will be
// appended to a []byte representing the Notion page
// (https://developers.notion.com/reference/page) being exported.
type Renderer interface {
	// RenderPageHeader receives reference to the original Page object. This
	// has metadata around the page, including the title. It returns the string
	// representation of anything that should be added to the top of the page
	// before the block content from the other renders are added.
	RenderPageHeader(page *na.Page, o ...headerFooterOverride) string
	// RenderPageFooter receives reference to the original Page object. This
	// has metadata around the page, including the title. It returns the string
	// representation of anything that should be added to the bottom of the
	// page after the block content from the other renders are added.
	RenderPageFooter(page *na.Page, o ...headerFooterOverride) string

	// RenderText is used in all other render calls. It converts blocks
	// instances of Notion Rich Text
	// (https://developers.notion.com/reference/rich-text) to a string
	// representation. In composing the string, RenderText should look at each
	// instance of na.RichText.Annotations to determine how to create the
	// string representation. These annotations will include formatting such
	// as:
	//
	// * bold
	// * italic
	// * strikethrough
	// * underline
	// * inline-code blocks
	RenderText([]na.RichText, ...richTextOverride) string

	// RenderPageHeader1 receives a Block that contains text with all
	// stylization (eg. bold, italic, etc) accounted for via RenderText. The
	// Block also has a reference to the original Notion Block object. Using
	// Block.GetType(), this can be casted to the appropriate type to
	// access its fields.
	//
	// Optionally, a blockOverride can be provided, this defines alternative
	// Render functionality, enabling you to change one aspect of the block
	// rendering without reimplementing the entire Renderer interface.
	//
	// Returned a string representation of the Notion Block.
	RenderPageHeader1(*Block, ...blockOverride) string
	// RenderPageHeader2 receives a Block that contains text with all
	// stylization (eg. bold, italic, etc) accounted for via RenderText. The
	// Block also has a reference to the original Notion Block object. Using
	// Block.GetType(), this can be casted to the appropriate type to
	// access its fields.
	//
	// Optionally, a blockOverride can be provided, this defines alternative
	// Render functionality, enabling you to change one aspect of the block
	// rendering without reimplementing the entire Renderer interface.
	//
	// Returned a string representation of the Notion Block.
	RenderPageHeader2(*Block, ...blockOverride) string
	// RenderPageHeader3 receives a Block that contains text with all
	// stylization (eg. bold, italic, etc) accounted for via RenderText. The
	// Block also has a reference to the original Notion Block object. Using
	// Block.GetType(), this can be casted to the appropriate type to
	// access its fields.
	//
	// Optionally, a blockOverride can be provided, this defines alternative
	// Render functionality, enabling you to change one aspect of the block
	// rendering without reimplementing the entire Renderer interface.
	//
	// Returned a string representation of the Notion Block.
	RenderPageHeader3(*Block, ...blockOverride) string

	// RenderParagraph receives text, which has been run through RenderText,
	// and a reference to the original ParagraphBlock object. It returns the
	// string representation of the paragraph.
	RenderParagraph(*Block, ...blockOverride) string

	// RenderBulletedList receives text, which has been run through RenderText,
	// and a reference to the original BullletedListItemBlock object. It
	// returns the string representation of the bulleted list item.
	RenderBulletedList(*Block, ...blockOverride) string
	// RenderNumberedList receives text, which has been run through RenderText,
	// and a reference to the original NumberedListItemBlock object. It returns
	// the string representation of the numbered list item.
	RenderNumberedList(*Block, ...blockOverride) string
	// RenderTodoList receives text, which has been run through RenderText,
	// and a reference to the original ToDoBlock object. It returns the
	// string representation of the todo list item.
	RenderTodoList(*Block, ...blockOverride) string

	// RenderCallout receives text, which has been run through RenderText,
	// and a reference to the original CalloutBlock object. It returns the
	// string representation of the callout.
	RenderCallout(*Block, ...blockOverride) string
	// RenderQuote receives text, which has been run through RenderText,
	// and a reference to the original QuoteBlock object. It returns the
	// string representation of the quote.
	RenderQuote(*Block, ...blockOverride) string

	// RenderCode receives text, which has been run through RenderText,
	// and a reference to the original CodeBlock object. It returns the
	// string representation of the quote.
	RenderCode(*Block, ...blockOverride) string

	// RenderDivder receives a reference to the original DividerBlock object.
	// It returns the string representation of the divider.
	RenderDivider(*Block, ...blockOverride) string
	// RenderImage receives a reference to the original ImageBlock object and
	// ImageSaveOptions which define instructions for how to handle images. It
	// will return a string representation of how the image should be
	// referenced. For example, in a markdown Renderer, the return may looks
	// like ![some-name](https://joshrosso.com/files/images/bmo.jpg).
	// RenderImage needs to handle both external images (hosted outside of
	// Notion) and internal images (hosted within Notion). For internal images,
	// a Renderer implementation should be able to download and save the image
	// to the local filesystem.
	RenderImage(*Block, ...imageOverride) (string, error)

	// RenderTableRow receives a list of cells that contain text that has been
	// run through ParseText and metadata around the table the row belongs to.
	// The cells passed in represent 1 row. By introspecting the tableCell
	// metadata, you'll find details like whether the row is a header, and
	// which row you're on (e.g. 0 == row 1, 2 == row 3).
	RenderTableRow([]tableCell, ...rowOverride) string

	// AddPadding receives the rendered text of string and the depth of the
	// Notion block. It returns the text with padding prefixed to each line
	// based on the depth. Depth is increased when an Block in Notion is
	// indented under another. For example, see the example below for depths:
	//
	// * list item one <-- depth: 0
	//     * list item two <-- depth: 1
	//       > quote <-- depth: 2
	//
	// In the above, blocks like '> quote' enter without padding. Thus,
	// implementations of AddPadding must calculate how many spaces (or tabs)
	// should be prefixed and return that representation to the caller.
	AddPadding(*Block, ...blockOverride) string
	// AddSectionSeperation is responsible for adding additional seperation
	// (often linebreaks) based on what the previous type was. For example. If
	// the previousType was a list element and current type is a list element,
	// you may wish to add 1 break. However, if the previous element was a
	// list element but the current is a paragraph, you may wish to add two
	// breaks to desperate them properly. This example is what's required to
	// separate sections in markdown.
	AddSectionSeperation(previousType string, currentType string,
		o ...seperationOverride) string
}

type exporter struct {
	c        *na.Client
	page     []byte
	Renderer Renderer
}

type Block struct {
	Text     string
	BlockRef na.Block
	Opts     []RenderOptions
	Depth    int
	// Reference to the page in case retrieving metadata (properties) are
	// useful for rending behavior.
	PageRef *na.Page
}

type ExporterOptions struct {
	NotionToken string
	ClientOpts  na.ClientOption
	// The desired format used to create the appropraite renderer for the exporter.
	Format string
	// The optional renderer instance to be used in the exporter. This acts as
	// a full override for injecting a custom renderer into an exporter. When
	// this is set, the Format option is ignored.
	Renderer Renderer
}
