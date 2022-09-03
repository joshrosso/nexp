package export

// This file contains the configuration structs and function overrides used to
// support rendering of Notion Blocks.

import (
	na "github.com/jomei/notionapi"
)

// RenderOptions contains settings for how rendering should occur. These render
// options are looked up by Renderer implementations to inform how to operate
// on Notion Blocks.
type RenderOptions struct {
	ImageOpts ImageSaveOptions
	Overrides OverrideOptions
	// SkipEmptyParagraphs will not send empty paragraphs to the renderer when
	// true.
	SkipEmptyParagraphs bool
	tableState          tableState
	previousElementType string
	depth               int
}

// OverrideOptions contains optional function definitions that can override the
// default behaviour of a block renderer.
//
// For example, when exporting a Notion Page, the RenderParagraph function may
// be called. A implementaion, such are a Markdown renderer will have default
// behavior defined. However, if a blockOverride function was defined and added
// to OverrideOptions.Paragraph, the instructions defined in that override will
// execute instead for every Paragraph Block.
type OverrideOptions struct {
	Header1      blockOverride
	Header2      blockOverride
	Header3      blockOverride
	Paragraph    blockOverride
	BulletedList blockOverride
	NumberedList blockOverride
	Divider      blockOverride
	Code         blockOverride
	Todo         blockOverride
	Quote        blockOverride
	Callout      blockOverride
	Image        blockOverride
	Padding      blockOverride
	Row          rowOverride
}

// ImageSaveOptions define how Image blocks may be handled.
type ImageSaveOptions struct {
	// SavePath is the location to persist images downloaded in this library.
	// When not set, the default is ./images.
	SavePath string
	// IgnoreImages instructs the renderer to not add images to the exported
	// output.
	IgnoreImages bool
	// OverwriteExisting forces the redownload of images even if the image
	// already exists on the local filesystem at the SavePath.
	OverwriteExisting bool
}

type tableState struct {
	tableBlock  *na.TableBlock
	rowQuantity int
	currentRow  int
}

type tableCell struct {
	rowTxt         string
	isRowHeader    bool
	isColumnHeader bool
	tableRef       tableState
	rowBlockRef    *na.TableRowBlock
}

// headerFooterOverride enables custom headers and footers for a renderer. It's
// provided an instance of the Notion Page, which holds properties (metadata)
// on the page in its Properties field. You can use this information to inform
// how a header or footer should be generated. When not provided, the
// Renderer's default header/footer creation behaviour will occur.
type headerFooterOverride func(*na.Page) string

// blockOverride enables custom rendering for a Block renderer. The primary use
// case is when you'd like to change how a block type, (e.g. paragraph, list,
// etc) is processed without implementing an entirely new Renderer.
// blockOverride is passed a Block object which contains text that has been run
// through RenderText, meaning text stylization (e.g. bold, italic, underline,
// etc) should be accounted for. While often you'll only need the rendered txt
// field, the Block argument also contains a reference to the original Notion
// Block object. This is an interface that, using GetType(), you can cast into
// the approriate type and access its fields.
type blockOverride func(*Block) string

// rowOverride enables custom rendering for all table rows in Notion Blocks.
//
// It receives a slice of tableCell elements where each slice represents a full
// row. Each element in tableCell represents a cell in the row. Inside each
// tableCell is the stylized text along with metadata about the table so you
// can make decisions such as whether the row is a header and, if so, stylize
// the row output as such.
type rowOverride func([]tableCell) string

// richTextOverride enables custom rendering for all text in the Notion Blocks.
//
// It receives a slice of RichText elements. You can expect each of these to contain a contiguous block of consistently stylized text.
//
// For example, consider the RichText that would come in as an argument for the following paragraph:
// "The unexamined life is **not** worth living."
//
// This would enter richTextOverride as a []na.RichText of size 3. The first
// element would contain "The unexamined life is ", the second "not", and third
// " worth living". In this case, the second element would contain the
// Annotation "Bold". Telling you this set of text was stylized as bold. This
// should be taken into account when building your string representation of the
// text.
//
// richTextOverride implementations should account for all possible sylizations
// one can do to text in the Notion (GUI) client.
type richTextOverride func([]na.RichText) string

// seperationOverride enables custom rendering for the seperation between
// sections. For example, this may be the seperation between a paragraph and an
// image. In the case of markdown, the default might be to add 2 linebreaks
// between these blocks. Instead of this default, setting a seperation override
// enables using an arbitrary number of breaks.
//
// p contains the previous element type while c contains the current. An
// implementation of seperation override should account for every Block type
// and return the correct seperation representation as a string.
type seperationOverride func(p string, c string) string
