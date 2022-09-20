package export

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	na "github.com/jomei/notionapi"
)

const (
	tokenEnvVarName = "NOTION_TOKEN"

	mdCodeBlockDelimiter   = "```"
	mdHeadingOnePattern    = "# %s"
	mdHeadingTwoPattern    = "## %s"
	mdHeadingThreePattern  = "### %s"
	mdLinkPattern          = "[%s](%s)"
	mdBoldPattern          = "**%s**"
	mdItalicPattern        = "_%s_"
	mdStrikeThroughPattern = "~%s~"
	mdInlineCodePattern    = "`%s`"
	mdListItemPattern      = "* %s"
	mdNumItemPattern       = "1. %s"
	mdTodoUncheckedPattern = "* [] %s"
	mdTodoCheckedPattern   = "* [x] %s"
	MdImagePattern         = "![%s](%s)"
	mdTableElementPattern  = "| %s "
	mdDividerPattern       = "---"
	mdQuotePattern         = "> %s"

	defaultImageSaveLocation = "images"
	notionImageExtension     = ".png"
)

var languages map[string]string

func init() {
	// list of language names which need to be swapped from the Notion
	// represention to a represntation friendlier for markdown parsers.
	languages = map[string]string{
		"c++":        "cpp",
		"c#":         "csharp",
		"f#":         "fsharp",
		"plain text": "txt",
	}
}

type MDRenderer struct {
}

// RenderPageHeader for MDRenderer takes a client's custom pageOverrider
// definition and returns its results. If a pageOverrider is not provided, it
// defaults to returning the title of the page, which should be added to the
// top of the page.
func (m *MDRenderer) RenderPageHeader(page *na.Page,
	o ...headerFooterOverride) string {

	// when an overrider is provided, use its render functionality.
	if len(o) > 0 {
		return o[0](page)
	}

	output := fmt.Sprintf(mdHeadingOnePattern, ResolveTitleInPage(page))

	return output
}

// RenderPageFooter for MDRenderer takes a client's custom pageOverrider
// definition and returns its results. If a pageOverrider is not provided, a
// blank footer is returned.
func (m *MDRenderer) RenderPageFooter(page *na.Page, o ...headerFooterOverride) string {
	// when an overrider is provided, use its render functionality.
	if len(o) > 0 {
		return o[0](page)
	}

	return ""
}

// RenderPageHeader1 for MDRenderer takes a client's the text object present in
// the Block and prepends "# " to it, resulting in a markdown header of level
// 1. This is returned to the caller. If an override is provided, that function
// is run and returned value is used instead.
func (m *MDRenderer) RenderPageHeader1(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdHeadingOnePattern, b.Text)
}

// RenderPageHeader2 for MDRenderer takes a client's the text object present in
// the Block and prepends "## " to it, resulting in a markdown header of level
// 2. This is returned to the caller. If an override is provided, that function
// is run and returned value is used instead.
func (m *MDRenderer) RenderPageHeader2(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdHeadingTwoPattern, b.Text)
}

// RenderPageHeader3 for MDRenderer takes a client's the text object present in
// the Block and prepends "### " to it, resulting in a markdown header of level
// 3. This is returned to the caller. If an override is provided, that function
// is run and returned value is used instead.
func (m *MDRenderer) RenderPageHeader3(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdHeadingThreePattern, b.Text)
}

// RenderParagraph for MDRenderer takes a client's the text object present in
// the Block and returns it. If an override is provided, that function
// is run and returned value is used instead.
func (m *MDRenderer) RenderParagraph(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return b.Text
}

// RenderParagraph for MDRenderer returns "---" representing a mardown divider.
// If an override is provided, that function is run and returned value is used
// instead.
func (m *MDRenderer) RenderDivider(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return mdDividerPattern
}

// RenderNumberedList for MDRenderer takes a client's the text object present
// in the Block and returns it prepended with "1. ". If an override is
// provided, that function is run and returned value is used instead.
func (m *MDRenderer) RenderNumberedList(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdNumItemPattern, b.Text)
}

// RenderBulletedList for MDRenderer takes a client's the text object present
// in the Block and returns it prepended with "* ". If an override is
// provided, that function is run and returned value is used instead.
func (m *MDRenderer) RenderBulletedList(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdListItemPattern, b.Text)
}

// The first row of cells retrieved is always treated as a row header. While
// Notion supports tables without row headers, many markdown parsers do not:
// (https://stackoverflow.com/questions/17536216). Similarlly, many markdown
// parsers do not support column headers, thus they are not respected here.
func (m *MDRenderer) RenderTableRow(cells []tableCell, o ...rowOverride) string {
	// when a rowOverride function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](cells)
	}

	var row string
	var currentRow int
	for _, c := range cells {
		currentRow = c.tableRef.currentRow
		row += fmt.Sprintf(mdTableElementPattern, c.rowTxt)
	}
	row += "|"
	// when row is the first, it's a header
	if currentRow == 0 {
		var rowHeader string
		for range cells {
			rowHeader += "| --- "
		}
		rowHeader += "|"
		row += "\n" + rowHeader
	}
	return row
}

func (m *MDRenderer) RenderTodoList(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	var tb *na.ToDoBlock
	if b.BlockRef.GetType() == "to_do" {
		tb = b.BlockRef.(*na.ToDoBlock)
	}
	if tb.ToDo.Checked {
		return fmt.Sprintf(mdTodoCheckedPattern, b.Text)
	}
	return fmt.Sprintf(mdTodoUncheckedPattern, b.Text)
}

func (m *MDRenderer) RenderCallout(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	// quote pattern used here as callouts are treated as markdown quotes
	return fmt.Sprintf(mdQuotePattern, b.Text)
}

func (m *MDRenderer) RenderQuote(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	return fmt.Sprintf(mdQuotePattern, b.Text)
}

func (m *MDRenderer) RenderImage(b *Block, o ...imageOverride) (string, error) {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	if b.BlockRef.GetType() != "image" {
		return "", fmt.Errorf("RenderImage was passed a %s but expected an ImageBlock", b.BlockRef.GetType())
	}

	config := resolveRenderConfig(b.Opts...)
	ib := b.BlockRef.(*na.ImageBlock)

	// image was not uploaded to Notion, but is referenced from an
	// external URL.
	if ib.Image.External != nil {
		// TODO(joshrosso): Friendly name is currently "image". Should think
		// about how to make this more eloquent.
		return fmt.Sprintf(MdImagePattern, "image", ib.Image.External.URL), nil
	}
	// image was uploaded to Notion, need to download to local
	// filesystem.
	var filePath string
	var err error
	if ib.Image.File != nil {
		filePath, err = SaveNotionImageToFilesystem(ib.Image.File.URL, config.ImageOpts)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf(MdImagePattern, "image", filePath), nil
}

func (m *MDRenderer) RenderCode(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	var cb *na.CodeBlock
	if b.BlockRef.GetType() == "code" {
		cb = b.BlockRef.(*na.CodeBlock)
	}

	r := mdCodeBlockDelimiter + ResolveLanguageForCodeBlock(cb.Code.Language) +
		"\n" + b.Text + "\n" + mdCodeBlockDelimiter

	return r
}

// RenderText takes the RichText object from the Notion API and parses it to
// rewrite all formatting requied. Examples are text that is bold, italicised,
// or a hyperlink.
func (m *MDRenderer) RenderText(rt []na.RichText, o ...richTextOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](rt)
	}

	var parsed string
	for _, t := range rt {
		switch {
		// text is a hyperlink
		case t.Href != "":
			parsed += fmt.Sprintf(mdLinkPattern, t.Text.Content, t.Href)

		// text is bolded
		case t.Annotations.Bold:
			parsed += fmt.Sprintf(mdBoldPattern, t.Text.Content)

		// text is italicised
		case t.Annotations.Italic:
			parsed += fmt.Sprintf(mdItalicPattern, t.Text.Content)

		// text is strikethrough
		case t.Annotations.Strikethrough:
			parsed += fmt.Sprintf(mdStrikeThroughPattern, t.Text.Content)

		// text is code
		case t.Annotations.Code:
			parsed += fmt.Sprintf(mdInlineCodePattern, t.Text.Content)

		// text is plain
		default:
			parsed += fmt.Sprintf(t.Text.Content)
		}
	}

	return parsed
}

func (m *MDRenderer) AddPadding(b *Block, o ...blockOverride) string {
	// when an override function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](b)
	}

	// when at root (depth: 0) do no padding processing
	if b.Depth == 0 {
		return b.Text
	}

	padding := ""
	for i := 0; i < b.Depth*4; i++ {
		padding += " "
	}

	paddedTxt := padding + b.Text
	// When there are line breaks in the block (e.g. code); pad the next line.
	paddedTxt = strings.ReplaceAll(paddedTxt, "\n", "\n"+padding)

	return paddedTxt
}

// ResolveLanguageForCodeBlock takes a Notion code block's language type as
// input and returns a representation more friendly for markdown parsers. For
// example, Notion uses 'plain text' for Plain Text codeblocks, however most
// markdown parsers expects this to be specified as 'txt'. For some languages,
// Notion uses the correct (for Markdown) name. In this cause, the language
// name passed is returned. If the language is entirely unknown, the language
// name passed is returned.
func ResolveLanguageForCodeBlock(language string) string {
	if val, ok := languages[language]; ok {
		return val
	}
	return language
}

func createPathIfNonExistent(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveNotionImageToFilesystem takes the URL of a Notion-hosted image. The URL
// is typically an S3 address. ImageSaveOptions can be optinally provided. If
// multiple options are provided, only the first is respected. By default the
// image is save in a ./images directory. If successful, the path the image was
// saved is returned. An error is returned if the image can not be returned or
// saved to the filesystem.
func SaveNotionImageToFilesystem(address string,
	opts ...ImageSaveOptions) (string, error) {

	// establish config for image save from options
	config := ResolveImageSaveOptions(opts...)
	createPathIfNonExistent(config.SavePath)

	// determine name of image using UUID created by notion
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	resources := strings.Split(u.Path, "/")
	if len(resources) < 2 {
		return "", fmt.Errorf("Path from Notion Image URL was invalid. Path was: %s", address)
	}
	fileName := resources[2]
	filePath := filepath.Join(config.SavePath, fileName) + notionImageExtension

	// if file exists, do no more and return the existing file's path
	if !config.OverwriteExisting {
		_, err := os.Stat(filePath)
		if !os.IsNotExist(err) {
			return filePath, nil
		}
	}

	// download the image from the Notion-provided URL
	// TODO(joshrosso): Don't rely on default HTTP client; need better control
	// of timeouts.
	resp, err := http.Get(address)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Non 200 status code returned when retrieveing."+
			"Code was: %d", resp.StatusCode)
	}

	// persist the downloaded image to the filesystem
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (m *MDRenderer) AddSectionSeperation(previousType string, currentType string, o ...seperationOverride) string {
	// when a rowOverride function is passed, call it and return its output
	if len(o) > 0 && o[0] != nil {
		return o[0](previousType, currentType)
	}

	// special conditions for single break
	if previousType == "table_row" && currentType == "table_row" {
		return "\n"
	}
	if previousType == "to_do" && currentType == "to_do" {
		return "\n"
	}
	if previousType == "numbered_list_item" && currentType == "numbered_list_item" {
		return "\n"
	}
	if previousType == "bulleted_list_item" && currentType == "bulleted_list_item" {
		return "\n"
	}

	// if now special condition, ensure currentType is known and will be
	// rendered
	switch currentType {
	case "heading_1":
		return "\n\n"

	case "heading_2":
		return "\n\n"

	case "heading_3":
		return "\n\n"

	case "table_row":
		return "\n\n"

	case "to_do":
		return "\n\n"

	case "numbered_list_item":
		return "\n\n"

	case "bulleted_list_item":
		return "\n\n"

	case "paragraph":
		return "\n\n"

	case "divider":
		return "\n\n"

	case "code":
		return "\n\n"

	case "quote":
		return "\n\n"

	case "callout":
		return "\n\n"

	case "image":
		return "\n\n"
	}

	// currentType won't be rendered, so don't bother with break.
	return ""
}

// createPadding takes the depth of a block (ie child) and calculates what the
// appropraite left padding is. It returns a string of spaces representing this
// padding.
func createPadding(depth int) string {
	padding := ""
	for i := 0; i < depth*4; i++ {
		padding += " "
	}
	return padding
}

// ResolveImageSaveOptions takes a list of ImageSaveOptions and sets defaults,
// overwritting them with any options specified. While it takes multiple
// arguments, it only respects the first option passed.
func ResolveImageSaveOptions(opts ...ImageSaveOptions) ImageSaveOptions {
	// setup default
	config := ImageSaveOptions{
		SavePath:     defaultImageSaveLocation,
		IgnoreImages: false,
	}

	// No options were provided; return the default
	if len(opts) < 1 {
		return config
	}

	if opts[0].SavePath != "" {
		config.SavePath = opts[0].SavePath
	}

	if opts[0].IgnoreImages {
		config.IgnoreImages = opts[0].IgnoreImages
	}

	return config
}
