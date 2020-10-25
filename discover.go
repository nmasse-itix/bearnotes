package bearnotes

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// DiscoverNotes walk through recursively the Bear notes directory to find notes.
// It generates a tag configuration file, suitable for migration.
func DiscoverNotes(notesDir string, tagFile string) error {
	var tags map[string]TagOptions = make(map[string]TagOptions)
	var imageCount int
	var fileCount int
	var noteCount int

	fmt.Printf("Looking for Bear notes into %s...\n", notesDir)

	err := filepath.Walk(notesDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("stat: %s: %s\n", path, err)
				return nil
			}

			if strings.HasSuffix(info.Name(), ".md") && !info.IsDir() { // it's a Markdown file!
				content, err := ioutil.ReadFile(path)
				if err != nil {
					log.Printf("open: %s: %s\n", path, err)
					return nil
				}
				note := LoadNote(string(content))
				imageCount += len(note.Images)
				fileCount += len(note.Files)
				noteCount++

				for _, tag := range note.Tags {
					// just to be safe, normalize the tag name since it is used
					// afterwards to generate paths and filenames
					tag.Name = norm.NFC.String(tag.Name)

					// all tags are lowercase in Bear
					tagName := strings.ToLower(tag.Name)

					tagEntry, ok := tags[tagName]
					if !ok {
						tags[tagName] = NewTagOptions(tag)
					} else {
						tagEntry.count++
						tags[tagName] = tagEntry
					}
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d notes, %d embedded images, %d attachments and %d unique tags.\n", noteCount, imageCount, fileCount, len(tags))
	fmt.Println("")

	// Displays all tags, sorted by their name
	fmt.Println("Tag list:")
	tagNames := make([]string, len(tags))
	i := 0
	for k := range tags {
		tagNames[i] = k
		i++
	}
	sort.Strings(tagNames)
	for _, tagName := range tagNames {
		fmt.Printf("#%s\n", tagName)
	}

	// Write the tag configuration file
	fmt.Println("")
	fmt.Printf("Writing all tags into %s...\n", tagFile)
	fileContent, err := yaml.Marshal(tags)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tagFile, fileContent, 0644)
	if err != nil {
		return err
	}

	return nil
}
