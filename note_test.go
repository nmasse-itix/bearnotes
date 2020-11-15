package bearnotes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTag(t *testing.T) {
	tagContent := " #test/123 "
	tag := NewTag(tagContent, []int{0, len(tagContent)})
	assert.Equal(t, "test/123", tag.Name, "tag name must be equal")

	// Back to string
	assert.Equal(t, " #test/123 ", tag.String(), "tag content must be equal")

	tag.Name = ""
	assert.Equal(t, "  ", tag.String(), "tag content must be empty")
}

func TestNewTagLookAround(t *testing.T) {
	testCases := [][]string{{" #test/123 ", "test/123"}, {"/#trap ", ""}, {" #trap#", ""}, {"#ok", "ok"}}
	for _, testCase := range testCases {
		tagContent := testCase[0]
		expected := testCase[1]
		tag := NewTag(tagContent, []int{0, len(tagContent)})
		assert.Equal(t, expected, tag.Name, "tag name must be equal")
	}
}

func TestNewFile(t *testing.T) {
	fileContent := `<a href='note/my%20file.pdf'>my file.pdf</a>`
	file := NewFile(fileContent, []int{0, len(fileContent)})
	assert.Equal(t, "note/my file.pdf", file.Location, "file location must be equal")
	assert.Equal(t, "my file.pdf", file.Name, "file name must be equal")

	// Back to string
	assert.Equal(t, "[my file.pdf](note/my%20file.pdf)", file.String(), "file content must be equal")
}

func TestNewImage(t *testing.T) {
	imageContent := `![my image](note/image%202.jpg)`
	image := NewImage(imageContent, []int{0, len(imageContent)})
	assert.Equal(t, "note/image 2.jpg", image.Location, "image location must be equal")
	assert.Equal(t, "my image", image.Description, "image description must be equal")

	// Back to string
	assert.Equal(t, "![my image](note/image%202.jpg)", image.String(), "image content must be equal")
}

func TestLoadNote(t *testing.T) {
	md := `# Sample Markdown title (not a tag)

## Files

<a href='note/my%20file.pdf'>my file.pdf</a>
<a href='note/my%20other%20file.pdf'>my other file.pdf</a>

## Images

![an image](note/image%202.jpg)
![](note/no-alt.jpg)
![](note_with_nested(parenthesis)/test.jpg)

## Tags

this is a paragraph with a #tag

And some tags in a list

- #foo
- #foo/bar@baz
- #and-a_very%special$one/avec/des/éèà

[it's a trap](https://www.perdu.com/#trap)

Another trap: https://www.perdu.com/#trap 

another trap: world #1

Traps, traps, traps... #trap#trap 

#two-tags #one-after-another

#end`

	note := LoadNote(md)

	// Tags
	assert.Len(t, note.Tags, 7, "There must be 7 tags")
	assert.Equal(t, "tag", note.Tags[0].Name, "first tag must be 'tag'")
	assert.Equal(t, "foo", note.Tags[1].Name, "second tag must be 'foo'")
	assert.Equal(t, "foo/bar@baz", note.Tags[2].Name, "third tag must be 'foo/bar@baz'")
	assert.Equal(t, "and-a_very%special$one/avec/des/éèà", note.Tags[3].Name, "fourth tag must be 'and-a_very%special$one/avec/des/éèà'")
	assert.Equal(t, "two-tags", note.Tags[4].Name, "fifth tag must be 'two-tags'")
	assert.Equal(t, "one-after-another", note.Tags[5].Name, "sixth tag must be 'one-after-another'")
	assert.Equal(t, "end", note.Tags[6].Name, "seventh tag must be 'end'")

	// Files
	assert.Len(t, note.Files, 2, "There must be 2 files")
	assert.Equal(t, "my file.pdf", note.Files[0].Name, "first file must be 'my file.pdf'")
	assert.Equal(t, "my other file.pdf", note.Files[1].Name, "second file must be 'my other file.pdf'")

	// Images
	assert.Len(t, note.Images, 3, "There must be 3 images")
	assert.Equal(t, "note/image 2.jpg", note.Images[0].Location, "first image must be 'note/image 2.jpg'")
	assert.Equal(t, "note/no-alt.jpg", note.Images[1].Location, "second image must be 'note/no-alt.jpg'")
	assert.Equal(t, "note_with_nested(parenthesis)/test.jpg", note.Images[2].Location, "third image must be 'note_with_nested(parenthesis)/test.jpg'")

	// Alter tags, files and images
	note.Tags[1].Name = ""
	note.Tags[6].Name = "not-really"
	note.Files[0].Location = "note2/my file.pdf"
	note.Files[1].Location = "note2/my other file.pdf"
	note.Images[0].Location = "note2/image 2.jpg"
	note.Images[1].Location = "note2/no-alt.jpg"

	// Write back the resulting note
	expectedMd := `# Sample Markdown title (not a tag)

## Files

[my file.pdf](note2/my%20file.pdf)
[my other file.pdf](note2/my%20other%20file.pdf)

## Images

![an image](note2/image%202.jpg)
![](note2/no-alt.jpg)
![](note_with_nested%28parenthesis%29/test.jpg)

## Tags

this is a paragraph with a #tag

And some tags in a list

- 
- #foo/bar@baz
- #and-a_very%special$one/avec/des/éèà

[it's a trap](https://www.perdu.com/#trap)

Another trap: https://www.perdu.com/#trap 

another trap: world #1

Traps, traps, traps... #trap#trap 

#two-tags #one-after-another

#not-really`
	newNote := note.WriteNote()
	assert.Equal(t, expectedMd, newNote, "notes must be equal")
}
