package cmd

import (
    "path/filepath"
    "github.com/spf13/cobra"
    "sakura/internal/gallery"
)

var dir, outputDir, templateDir, prevLabel, nextLabel string
var photosPerPage int
var verbose bool
var quiet bool

const defaultTemplateDir = ""

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Generate the photo gallery",
    Long: `Generate a photo gallery from Instagram backup JSON files.
You can specify the directory containing the Instagram backup, the output directory for HTML pages, the number of photos per page, and optional labels.`,
    Run: func(cmd *cobra.Command, args []string) {
        absDir := gallery.GetAbsolutePath(dir)
        absOutputDir := gallery.GetAbsolutePath(outputDir)
        absTemplateDir := gallery.GetAbsolutePath(templateDir)
        if templateDir == "" {
            absTemplateDir = defaultTemplateDir
        }

        if quiet {
            gallery.SetLogLevel(gallery.LOG_LEVEL_QUIET)
        } else if verbose {
            gallery.SetLogLevel(gallery.LOG_LEVEL_VERBOSE)
        } else {
            gallery.SetLogLevel(gallery.LOG_LEVEL_NORMAL)
        }

        gallery.CreateOutputDir(absOutputDir)

        gallery.Generate(
            filepath.Join(absDir, "your_instagram_activity/content"),
            absOutputDir,
            absTemplateDir,
            photosPerPage,
            prevLabel,
            nextLabel,
        )
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
    serveCmd.Flags().StringVarP(&dir, "dir", "d", ".", "Path to the Instagram backup root folder")
    serveCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Path to the output folder for HTML pages")
    serveCmd.Flags().StringVarP(&templateDir, "template", "t", "", "Path to the template folder (default: templates)")
    serveCmd.Flags().IntVarP(&photosPerPage, "photos-per-page", "p", 8, "Number of photos per page")
    serveCmd.Flags().StringVarP(&prevLabel, "prev-label", "P", "Previous", "Label for the 'Previous' button")
    serveCmd.Flags().StringVarP(&nextLabel, "next-label", "N", "Next", "Label for the 'Next' button")
    serveCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
    serveCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet mode (no logging)")
}
