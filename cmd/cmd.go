package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joshrosso/nexp/config"
	ne "github.com/joshrosso/nexp/export"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.Flags().StringP("to-file", "o", "", "Write export content to file specified instead of standard out.")
	exportCmd.Flags().StringP("format", "f", "markdown", "Export format for page.")
	exportCmd.Flags().StringP("token", "t", "", "Define an API token to use for"+
		" operations. By default the env var NOTION_TOKEN is used or the token value"+
		" in ${HOME}/.config/nexp.yaml")
	exportCmd.Flags().StringP("image-directory", "d", "images", "Location to store Notion-hosted images.")
	exportCmd.Flags().Bool("disable-images", false, "Skips all images found in pages.")
	exportCmd.Flags().Bool("skip-empty-paragraphs", false, "Omit any empty paragraph blocks from the output.")
	exportCmd.Flags().Bool("overwrite-existing-images", false, "Redownloads images even existing copies are found on the filesytem.")
}

var rootCmd = &cobra.Command{
	Use:   "nexp",
	Short: "Download notion Pages and export them to a target format.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export notion page to target format.",
	Run:   RunExport,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save Notion token for use with nexp.",
	Run:   RunLogin,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func SetupCommands() *cobra.Command {
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(loginCmd)
	return rootCmd
}

func RunExport(cmd *cobra.Command, args []string) {
	// ignore the error here as no format flag should result in an empty
	// string.
	f, _ := cmd.Flags().GetString("format")

	eopts := ne.ExporterOptions{
		NotionToken: "",
		Format:      f,
		Renderer:    nil,
	}
	e, err := ne.NewExporter(eopts)
	if err != nil {
		fmt.Printf("Failed creating exporter. Error: %s", err)
		os.Exit(1)
	}
	e.Renderer, err = ne.NewRenderer(f)
	if err != nil {
		fmt.Printf("Failed attaching renderer to exporter. Error: %s", err)
		os.Exit(1)
	}

	if len(args) < 1 {
		fmt.Println("A proper page identifier was not provided.")
		os.Exit(1)
	}
	// TODO(joshrosso): Needs more robust UUID detection
	// notion expects all page references to be UUID 4 with dashes removed.
	// this can be validated by looking at the argument passed for a 32
	// character alphanumeric string.
	reg, err := regexp.Compile("[a-z0-9]{32}$")
	if err != nil {
		fmt.Println("Unexpected error compiling regex for page ID detection")
		os.Exit(1)
	}
	pageID := reg.FindString(args[0])
	if pageID == "" {
		fmt.Printf("Could not detect valid page UUID for %s\n", args[0])
		os.Exit(1)
	}

	savePath, _ := cmd.Flags().GetString("image-directory")
	ignoreImages, _ := cmd.Flags().GetBool("disable-images")
	overwriteExistingImages, _ := cmd.Flags().GetBool("overwrite-existing-images")
	skipEmptyParagraphs, _ := cmd.Flags().GetBool("skip-empty-paragraphs")
	ropts := ne.RenderOptions{
		ImageOpts: ne.ImageSaveOptions{
			SavePath:          savePath,
			IgnoreImages:      ignoreImages,
			OverwriteExisting: overwriteExistingImages,
		},
		SkipEmptyParagraphs: skipEmptyParagraphs,
	}

	out, err := e.Render(pageID, ropts)
	if err != nil {
		fmt.Printf("Page exporting failed. Error: %s\n", err)
		os.Exit(1)
	}

	// check whether an output file was specified. If it was, write to the file
	// as opposed to printing output to standard out.
	toFile, _ := cmd.Flags().GetString("to-file")
	if toFile == "" {
		fmt.Printf("%s\n", out)
		// TODO(joshrosso): Refactor this function. Not great that there's a
		// random, non-error "return" here.
		os.Exit(0)
	}
	err = os.WriteFile(toFile, out, 0666)
	if err != nil {
		fmt.Printf("Failed to write file to %s, error: %s", toFile, err)
	}
}

func RunLogin(cmd *cobra.Command, args []string) {
	c, err := config.LoadNexpConfig()
	if err != nil {
		fmt.Println("Failed to load configuration file")
		os.Exit(1)
	}
	if len(args) < 1 {
		fmt.Println("Must provide login token.")
		os.Exit(1)
	}
	c.Token = args[0]

	err = config.SaveNexpConfig(*c)
	if err != nil {
		fmt.Printf("Failed to update config with token. Error: %s\n", err)
		os.Exit(1)
	}
}
