/*
Copyright © 2020 Nicolas Massé <nicolas.masse@itix.fr>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"log"

	"github.com/nmasse-itix/bearnotes"
	"github.com/spf13/cobra"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discovers your notes to extract tags",
	Long:  `Parses your notes to extract tags.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := bearnotes.DiscoverNotes(fromDir, tagFile)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	discoverCmd.Flags().StringVar(&fromDir, "from", "", "directory holding your Bear notes")
	discoverCmd.Flags().StringVar(&tagFile, "tag-file", "", "filename for the generated tag file")
	discoverCmd.MarkFlagRequired("from")
	discoverCmd.MarkFlagRequired("tag-file")
	rootCmd.AddCommand(discoverCmd)
}
