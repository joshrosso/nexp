# nexp

⚠️:: I'm currently wrapping up the first release of this library and it's not
ready for use.

A CLI and Go Library for exporting [Notion](https://www.notion.so/product)
pages to formats such as Markdown or HTML.

`nexp` leverages [joemi/notionapi](https://github.com/jomei/notionapi) to query [the official
Notion API](https://developers.notion.com/). It then parses each Notion block and renders a text
representation based on your desired format. You can download `nexp` as a statically compiled
binary to quicky export pages or include in scripts. Additionally, `nexp` can be used as a library
in your Go code where, along with `joemi/notionapi`, you can query and setup advanced rules
exporting pages.

You may consider `nexp` to:

* [Export to different formats via pipelines/scripts.](TODO)
* [Keep local copies of your Notion pages.](TODO)
* [Utilize Notion as a CMS for your website.](TODO)

## Documentation

* [nexp.joshrosso.com](nexp.joshrosso.com)
	* Official documentation with usage examples.
	* Managed in Notion and rendered using `nexp`.
* [pkg.go.dev/github.com/joshrosso/nexp](https://pkg.go.dev/github.com/joshrosso/nexp)
	* Go API documentation.

## Install and Basic Usage

Visit [the documentation](nexp.joshrosso.com) for details beyond this section.

### As a CLI

1. Download and add `nexp` to your `$PATH`.

	a. [Download from GitHub releases.](https://github.com/joshrosso/nexp/releases)
	and move it to your path.

	b. Use `go install github.com/joshrosso/nexp`.

2. Add your Notion integration token to `nexp`.

	```sh
	nexp login
	token: <paste token here and hit enter>
	```

	> You can create an integration token at
	> ([https://www.notion.so/my-integrations](https://www.notion.so/my-integrations)). Ensure
	> whatever page you'd like to access you've shared with this integration (can be done in the
	> Notion GUI by clicking share on a page.)

3. Export a page to Markdown.

	```sh
	nexp export 71ad7bd4cbae457f809dd313aa595b4a
	```

	>  Alternatively, you can use the page URL as the argument above, for
	>  example, the full URL of the above is
	>  https://www.notion.so/joshrosso/Days-of-Future-Passed-71ad7bd4cbae457f809dd313aa595b4a

4. The resulting Markdown will be printed to standard out and can be piped accordingly.

	> Images that are hosted in Notion are saved to `./images/<image-name>`.

### As a Library

To use `nexp` as a library, you will import the `nexp/export` package.

1. Add `nexp/export` to your Go project.

	```sh
	go get github.com/joshrosso/nexp/export
	```

1. Import `nexp`.

	```go
	package main

	import (
		"fmt"
		"github.com/joshrosso/nexp/export"
	)

	const (
		pageID = "71ad7bd4cbae457f809dd313aa595b4a"
	)

	func main() {
		fmt.Println("hello world")

		e, err := export.NewExporter()
		if err != nil {
			panic(err)
		}
		r, err := export.NewRenderer("markdown")
		if err != nil {
			panic(err)
		}

		output, err := e.Render(pageID, r)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", output)
	}

	```

1. Run the above, and view the Markdown output.

	```sh
	go run main.go
	```

	> Images that are hosted in Notion are saved to `./images/<image-name>`.

## Development & Contributing

* Run `make` to get help text on development tasks.
* Issues and Pull Requests welcome.
	* Consider opening an issue for non-trivial changes before implementing and opening a PR.
* Line width for code contributions must be <= 99 columns.
