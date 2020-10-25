# Migrate your notes from Bear to Zettlr!

## Background

I recently switch from Bear to Zettlr to manage my notes.
However, I was sad to discover that Bear exports notes in a non-standard way:

- File attachments are HTML links that don't even point to the correct file
- All notes are stored in the same directory, regardless of your tag hierarchy
- The tag format used by Bear (#foo/bar) is incompatible with Zettlr format (#bar)
- Bonus: some file attachment were missing...

So, I decided to write this tool in order to save my hundreds of notes!

## Usage

First, export all your notes from Bear.

* Go to **Notes**
* Hit **Cmd+A** to select all your notes
* Go to **File** > **Export Notes...**
* Select **Markdown** and check **Export attachments**
* Select a directory to store your exported notes
* Click **Export notes**

Then, install **git** and **go**.

```sh
brew install git golang
```

Checkout this repository.

```sh
git clone https://github.com/nmasse-itix/bearnotes.git
cd bearnotes/cli
```

Start a discovery of your notes.

```sh
go run main.go discover --from /path/to/bear-notes --tag-file /tmp/tags.yaml
```

If everything goes well, it should display a count of your exported notes
along with the discovered tag list.

You can review the generated tag configuration file.

```sh
open /tmp/tags.yaml
```

Finally, launch the proper migration phase.

```sh
go run main.go migrate --from /path/to/bear-notes --to /path/to/zettlr-notes --tag-file /tmp/tags.yaml
```

Review the migrated notes.

If you want to change the default folder hierarchy, read the next section.

## Configuration

You can configure how the migration tool stores your notes, in which folder and even rewrite the tags to match Zettlr's format.

The default configuration for a tag is:

```yaml
foo/bar:
    ignore: false
    handling_strategy: same-folder
    target_directory: foo/bar
    target_tag_name: bar
```

It defines that any note having this tag will go to the **foo/bar** directory.
The **#foo/bar** tag will be rewritten as **#bar**.
All the notes having the **#foo/bar** tag, will be stored in the same directory, along with their embedded images and file attachments.

If you think the migration tool wrongly identified a tag, you can switch the **ignore** option to **true**.

```yaml
foo/bar:
    ignore: true
```

If you want to rewrite the **#foo/bar** tag as **#foo-bar**, you can change the **target_tag_name**. 

```yaml
foo/bar:
    ignore: false
    handling_strategy: same-folder
    target_directory: foo/bar
    target_tag_name: foo-bar
```

Note: If you want to remove the tag from the migrated note, use `target_tag_name: ""`.

The `target_directory` option is straightforward: it defines where to store the notes having this tag.

The `handling_strategy` option specifies how notes will be saved on the filesystem

- **same-folder**: all notes having this tag are stored in the **target_directory** along with their embedded images and file attachments.
- **one-note-per-folder**: each note will get a sub-folder in the **target_directory**

Note: given that a document can have multiple tags, it is perfectly valid for a tag to specify no target directory or no handling strategy if you know that another tag will provide them. 

If by any chance, for a note the tool cannot determine a target directory or an handling strategy, the note will be stored in the root of the target directory.

And if a note receives different configurations by two different tags, the first one wins (by order of tag appearance in the document).

## License

MIT
