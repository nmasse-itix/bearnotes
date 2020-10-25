package bearnotes

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v2"
)

// MigrateNotes takes a source directory (from), a destination directory (to),
// a tag configuration file (tagFile) and performs a Bear to Zettlr migration.
func MigrateNotes(from string, to string, tagFile string) error {
	var tags map[string]TagOptions = make(map[string]TagOptions)

	fmt.Printf("Reading the tag file from %s...\n", tagFile)
	fileContent, err := ioutil.ReadFile(tagFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(fileContent, &tags)
	if err != nil {
		return err
	}

	fmt.Printf("Migrating Bear notes from %s to %s...\n", from, to)
	var success int = 0
	var allNotes int = 0
	err = filepath.Walk(from,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("stat: %s: %s\n", p, err)
				return nil
			}

			// If it's not a markdown file, skip it.
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			log.Printf("Processing %s...\n", info.Name())
			allNotes++

			// Load the note
			content, err := ioutil.ReadFile(p)
			if err != nil {
				log.Printf("open: %s: %s\n", p, err)
				return nil
			}
			note := LoadNote(string(content))

			// Iterate over the note's tags to compute the target directory & handling strategy.
			// Since a note can have multiple tags, the first tag that defines a valid (non-empty)
			// target directory and/or handling strategy sets the value.
			// If another one specifies a different value, we issue a warning.
			var targetDir string
			var handlingStrategy string
			for i, tag := range note.Tags {
				// Normalize tag names to prevent file not found errors because of Unicode encoding.
				tag.Name = norm.NFC.String(tag.Name)
				// And make it lowercase since all tags are lower-case in Bear.
				tagName := strings.ToLower(tag.Name)

				tagOption, ok := tags[tagName]
				if !ok {
					log.Printf("ERROR: Unknown tag name '%s' in %s! Re-run the discover command!\n", tagName, info.Name())
					return nil
				}

				if tagOption.Ignore {
					continue
				}

				// Rewrite the tag name as instructed
				note.Tags[i].Name = tagOption.TargetTagName

				if tagOption.TargetDirectory != "" && targetDir != "" && targetDir != tagOption.TargetDirectory {
					log.Printf("WARNING: Target directory '%s' for tag '%s' conflict with directives (%s) from another tag. Continuing with existing value.\n", tagOption.TargetDirectory, tagName, targetDir)
				} else {
					targetDir = tagOption.TargetDirectory
				}

				if tagOption.HandlingStrategy != "" && handlingStrategy != "" && handlingStrategy != tagOption.HandlingStrategy {
					log.Printf("WARNING: Handling strategy '%s' for tag '%s' conflict with directives (%s) from another tag. Continuing with existing value.\n", tagOption.HandlingStrategy, tagName, handlingStrategy)
				} else {
					if tagOption.HandlingStrategy == "same-folder" || tagOption.HandlingStrategy == "one-note-per-folder" || tagOption.HandlingStrategy == "" {
						handlingStrategy = tagOption.HandlingStrategy
					} else {
						log.Printf("WARNING: Unknown handling strategy '%s' for tag '%s'.\n", tagOption.HandlingStrategy, tagName)
					}
				}
			}

			// Compute the final target directory, based on the handling strategy
			noteName := strings.TrimSuffix(info.Name(), ".md")
			if handlingStrategy == "one-note-per-folder" {
				targetDir = path.Join(to, targetDir, noteName)
			} else if handlingStrategy == "same-folder" {
				targetDir = path.Join(to, targetDir)
			} else {
				// If no tag set an handling strategy or if the note has no tag,
				// then it goes at the root of the target directory
				targetDir = to
			}

			// Creates all the directory hierarchy
			err = os.MkdirAll(targetDir, 0755)
			if err != nil {
				log.Printf("mkdir: %s: %s\n", targetDir, err)
				return nil
			}

			// Migrate embedded images
			for i, image := range note.Images {
				// Normalize filenames to prevent 'file not found' errors
				imageFileName := filepath.Base(norm.NFC.String(image.Location))
				source := filepath.Join(from, norm.NFC.String(image.Location))

				destination := filepath.Join(targetDir, imageFileName)
				_, err := os.Stat(destination)
				if os.IsNotExist(err) {
					// Copy the image only if we don't overwrite an existing one
					err = copyFile(source, destination)
					if os.IsNotExist(err) {
						log.Printf("WARNING: source image '%s' in note %s cannot be found!\n", imageFileName, noteName)
					} else if err != nil {
						log.Printf("copy: %s -> %s: %s\n", source, destination, err)
						return nil
					}
				} else if err != nil {
					log.Printf("stat: %s: %s\n", destination, err)
					return nil
				} else {
					log.Printf("WARNING: embedded image '%s' of note %s already exists in the target directory %s!\n", imageFileName, noteName, destination)
				}
				note.Images[i].Location = imageFileName
			}

			// Migrate file attachments
			for i, file := range note.Files {
				// Normalize filenames to prevent 'file not found' errors
				fileName := filepath.Base(norm.NFC.String(file.Location))
				source := filepath.Join(from, noteName, norm.NFC.String(file.Location))

				destination := filepath.Join(targetDir, fileName)
				_, err := os.Stat(destination)
				if os.IsNotExist(err) {
					// Copy the file attachment if we don't overwrite an existing one
					err = copyFile(source, destination)
					if os.IsNotExist(err) {
						log.Printf("WARNING: source file '%s' in note %s cannot be found!\n", fileName, noteName)
					} else if err != nil {
						log.Printf("copy: %s -> %s: %s\n", source, destination, err)
						return nil
					}
				} else if err != nil {
					log.Printf("stat: %s: %s\n", destination, err)
					return nil
				} else {
					log.Printf("WARNING: file attachment '%s' of note %s already exists in the target directory %s!\n", fileName, noteName, destination)
				}
				note.Files[i].Location = fileName
			}

			// Write back the updated note
			newNote := note.WriteNote()
			targetNoteFileName := filepath.Join(targetDir, info.Name())
			fd, err := os.Create(targetNoteFileName)
			if err != nil {
				log.Printf("open: %s: %s\n", targetNoteFileName, err)
				return nil
			}
			defer fd.Close()
			fd.WriteString(newNote)
			success++

			return nil
		})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Processed %d notes with %d successes and %d failures\n", allNotes, success, allNotes-success)

	return nil
}

// from https://opensource.com/article/18/6/copying-files-go
func copyFile(src string, dest string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}
