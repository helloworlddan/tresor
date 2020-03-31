package cmd

import (
	"fmt"
	"strings"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xlab/treeprint"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "List remote directory in a tree-like structure.",
	Long:  `List remote directory in a tree-like structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		prefixFilter := ""
		if len(args) == 1 {
			prefixFilter = args[0]
		}

		attrs, err := tresor.QueryStorage(viper.Get("bucket").(string), prefixFilter, false)
		if err != nil {
			fail(err)
		}

		root := treeprint.New()
		for _, attr := range attrs {
			attach(root, attr.Name)
		}
		fmt.Println(root.String())
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)
}

func attach(tree treeprint.Tree, path string) treeprint.Tree {
	if path == "" {
		return tree
	}
	pieces := strings.Split(path, "/")
	sub := tree.FindByValue(pieces[0])
	if sub == nil {
		sub = tree.AddBranch(pieces[0])
	}
	return attach(sub, strings.Join(pieces[1:], "/"))
}
