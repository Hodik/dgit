package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dgit",
	Short: "dgit is a Git implementation in Go",
	Long: `A Git implementation built from scratch in Go
to understand the internals of Git version control system.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new dgit repository",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Initializing repository...")
		initDgit()
	},
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Record changes to the repository",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No message provided")
			return
		}
		cmd.Println(commit(args[0]))
	},
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Compute object ID and optionally creates a blob from a file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No file provided")
			return
		}

		data, err := os.ReadFile(args[0])
		if err != nil {
			cmd.Println("Error reading file:", err)
			return
		}
		cmd.Println(hashObject(data, "blob"))
	},
}

var catObjectCmd = &cobra.Command{
	Use:   "cat-object",
	Short: "Provide content of repository object",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No hash provided")
			return
		}

		cmd.Println(catObject(args[0], ""))
	},
}

var writeTreeCmd = &cobra.Command{
	Use:   "write-tree",
	Short: "Create a tree object from the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(writeTree("."))
	},
}

var readTreeCmd = &cobra.Command{
	Use:   "read-tree",
	Short: "Read a tree object",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("No hash provided")
			return
		}

		readTree(args[0])
	},
}

var logCmd = &cobra.Command{
  Use:   "log",
  Short: "Show commit logs",
  Run: func(cmd *cobra.Command, args []string) {
    log()
  },
}

var checkoutCmd = &cobra.Command{
  Use:   "checkout",
  Short: "Checkout a commit",
  Run: func(cmd *cobra.Command, args []string) {
    if len(args) == 0 {
      cmd.Println("No hash provided")
      return
    }

    checkout(args[0])
  },
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(hashObjectCmd)
	rootCmd.AddCommand(catObjectCmd)
	rootCmd.AddCommand(writeTreeCmd)
	rootCmd.AddCommand(readTreeCmd)
  rootCmd.AddCommand(commitCmd)
  rootCmd.AddCommand(logCmd)
  rootCmd.AddCommand(checkoutCmd)
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
