package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "sakura",
    Short: "A tool to generate HTML photo galleries from Instagram downloads",
    Long:  `Sakura is a command line tool to generate static HTML photo galleries from downloaded Instagram photos.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Welcome to the Sakura Command Line Interface.")
        fmt.Println("Use the '--help' flag to see available commands and options.")
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    // No need to initialize configuration
}
