package bearnotes

import "strings"

// TagOptions specifies how to convert notes having this tag.
type TagOptions struct {
	// count is used in the discover phase to count notes having this tag

	count int `yaml:"-"`
	// When true, Ignore specifies that this tag is not relevant.
	// It can be useful when a tag is wrongly identified.
	Ignore bool `yaml:"ignore"`

	// HandlingStrategy specifies how notes will be saved on the filesystem
	// - same-folder:         all notes having this tag are stored in the TargetDirectory
	//                        along with their embedded images and file attachments.
	// - one-note-per-folder: each note will get a sub-folder in the TargetDirectory
	// - "" (empty string):   no handling specified for this tag
	HandlingStrategy string `yaml:"handling_strategy"`

	// TargetDirectory specifies where to store notes, along with their images and files
	TargetDirectory string `yaml:"target_directory"`

	// TargetTagName specifies the new tag name. Since Bear supports nested tags (#foo/bar)
	// but Zettlr does not, by default the target is the last component of the Bear tag (#bar).
	//
	// If TargetTagName is the empty string, the tag is removed from the note.
	TargetTagName string `yaml:"target_tag_name"`
}

// NewTagOptions initializes a new TagOptions from a Tag object, with sane defaults
// and counter == 1
func NewTagOptions(tag Tag) TagOptions {
	tagComponents := strings.Split(tag.Name, "/")
	lastComponent := tagComponents[len(tagComponents)-1]
	return TagOptions{count: 1, HandlingStrategy: "same-folder", TargetDirectory: tag.Name, TargetTagName: lastComponent}
}
