// Package bearnotes provides tools to read Markdown files generated
// by the Bear app. It can also convert those files to a format suitable
// for Zettlr.
//
// It handles notes, embedded images and file attachments.
//
// Note: there are some Unicode normalization issues between the filenames
// in the filesystem and paths in the Markdown file. It is up to the caller
// to normalize strings when required.
package bearnotes

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Regular expression to detect Bear tags.
// Examples:
//  - #foo
//  - #bar/baz
var reTag *regexp.Regexp

// Regular expression to detect file attachments.
// Example: <a href='my%20file.pdf'>my file.pdf</a>
var reFile *regexp.Regexp

// Regular expression to detect embedded images.
// Example: ![](note/my-image.png)
var reImage *regexp.Regexp

func init() {
	// This regex has a catch: it matches a leading and trailing extra character.
	// This is because Go does not support look-ahead/look-behind markers.
	// So we need to implement look-ahead/look-behind by ourself.
	reTag = regexp.MustCompile(`(^|.?)#([\p{L}][-\p{L}\p{N}/$_§%=+°({[\\@]*)(.?|$)`)

	// Those two regex are straightforward
	reFile = regexp.MustCompile(`<a +href=['"]([^'"]+)['"]>([^<]+)</a>`)
	reImage = regexp.MustCompile(`!\[([^\]]*)]\(([^(]+)\)`)
}

// Tag represents a Bear tag (#foo)
type Tag struct {
	// The name of the tag (without the leading hashtag)
	Name string
	// Position of this tag in the Markdown file
	position []int
	// The character before the tag (for look-ahead, see Regex description above)
	before string
	// The character after the tag (for look-behind, see Regex description above)
	after string
}

// NewTag creates a Tag from its content (including leading and trailing
// characters) and position in file.
func NewTag(content string, position []int) Tag {
	var tag Tag
	parts := reTag.FindStringSubmatch(content)
	if len(parts) > 0 {
		beforeIsEmpty := len(parts[1]) == 0
		before, _ := utf8.DecodeRuneInString(parts[1])
		beforeIsSpace := unicode.IsSpace(before)
		afterIsEmpty := len(parts[3]) == 0
		after, _ := utf8.DecodeRuneInString(parts[3])
		afterIsSpace := unicode.IsSpace(after)

		// A valid tag is surrounded by either a space character or nothing
		if (beforeIsEmpty || beforeIsSpace) && (afterIsEmpty || afterIsSpace) {
			tag.position = position
			tag.before = parts[1]
			tag.Name = parts[2]
			tag.after = parts[3]
		}
	}
	return tag
}

// String converts the Tag back to string.
func (tag *Tag) String() string {
	if len(tag.Name) == 0 {
		return fmt.Sprintf("%s%s", tag.before, tag.after)
	}

	return fmt.Sprintf("%s#%s%s", tag.before, tag.Name, tag.after)
}

// File represents a file attachment in a note.
type File struct {
	Location string // The path to the file attachment
	Name     string // The name of the file
	position []int  // The position in the Markdown file
}

// NewFile creates a File from the Markdown content and position in file.
func NewFile(content string, position []int) File {
	var file File
	parts := reFile.FindStringSubmatch(content)
	if len(parts) > 0 {
		file.Location, _ = url.PathUnescape(parts[1])
		file.Name = parts[2]
		file.position = position
	}
	return file
}

// URL encode a path, component by component so that slashes do not go
// through URL encoding.
func escapePath(path string) string {
	pathComponents := strings.Split(path, "/")
	var escapedPath strings.Builder
	for i, pathComponent := range pathComponents {
		if i > 0 {
			escapedPath.WriteString("/")
		}
		escapedPath.WriteString(url.PathEscape(pathComponent))
	}
	return escapedPath.String()
}

// String converts a file attachment back to Markdown syntax suitable for Zettlr.
func (file *File) String() string {
	return fmt.Sprintf("[%s](%s)", file.Name, escapePath(file.Location))
}

// Image represents an embedded image in a note.
type Image struct {
	Location    string // The path to the embedded image
	Description string // The alternative text for the image
	position    []int  // The position in the Markdown file
}

// NewImage creates an Image from the Markdown content and position in file.
func NewImage(content string, position []int) Image {
	var image Image
	parts := reImage.FindStringSubmatch(content)
	if len(parts) > 0 {
		image.Location, _ = url.PathUnescape(parts[2])
		image.Description = parts[1]
		image.position = position
	}
	return image
}

// String converts an image back to Markdown syntax suitable for Zettlr.
func (image *Image) String() string {
	return fmt.Sprintf("![%s](%s)", image.Description, escapePath(image.Location))
}

// Note represents a Bear note with its tags, file attachments and embedded images.
type Note struct {
	Tags    []Tag   // All the tags
	Files   []File  // All the file attachments
	Images  []Image // All the embedded images
	content string  // The full note content
}

// LoadNote parses a Bear note in Markdown format and returns a Note object.
func LoadNote(content string) *Note {
	var note Note
	note.content = content
	for _, match := range reTag.FindAllStringIndex(content, -1) {
		tag := NewTag(content[match[0]:match[1]], match)
		if len(tag.Name) > 0 {
			note.Tags = append(note.Tags, tag)
		}
	}
	for _, match := range reFile.FindAllStringIndex(content, -1) {
		note.Files = append(note.Files, NewFile(content[match[0]:match[1]], match))
	}
	for _, match := range reImage.FindAllStringIndex(content, -1) {
		note.Images = append(note.Images, NewImage(content[match[0]:match[1]], match))
	}
	return &note
}

// updatedItem is used to sort tags, images and files by their order
// of appearance in the file.
type updatedItem struct {
	content  string // tag, file or image content
	position []int  // position in file
}

// WriteNote converts the note back into a format suitable for Zettlr.
func (note *Note) WriteNote() string {
	// Tags, Images and Files are all stored into a common list
	var items []updatedItem
	for _, item := range note.Tags {
		items = append(items, updatedItem{item.String(), item.position})
	}
	for _, item := range note.Files {
		items = append(items, updatedItem{item.String(), item.position})
	}
	for _, item := range note.Images {
		items = append(items, updatedItem{item.String(), item.position})
	}
	// And sorted by their order of appearance in the file
	// Note: this only works when items do not overlap (which hopefully
	// is the case in most, if not all, markdown files).
	sort.Slice(items, func(i, j int) bool {
		return items[i].position[0] < items[j].position[1]
	})

	// Go through all items and copy the updated version of the item along
	// with the interleaved original excerpts
	var current int
	var newContent strings.Builder
	for _, item := range items {
		newContent.WriteString(note.content[current:item.position[0]])
		newContent.WriteString(item.content)
		current = item.position[1]
	}
	newContent.WriteString(note.content[current:len(note.content)])

	return newContent.String()
}
