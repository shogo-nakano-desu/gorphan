package parser

import (
	"reflect"
	"testing"
)

func TestExtractLocalMarkdownLinks(t *testing.T) {
	content := `
[inline](./guide/intro.md)
[ext](https://example.com/a.md)
[mail](mailto:test@example.com)
[fragment](./guide/intro.md#part)
[query](./guide/intro.md?raw=1)
[noext](./guide/intro)
[escaped](./guide/my\ file.md)

[ref-doc][doc]
[doc]: ./ref/readme.markdown#top
[[WikiPage]]
[[nested/Guide|Guide Page]]
`

	got := ExtractLocalMarkdownLinks(content, []string{".md", ".markdown"})
	want := []string{
		"guide/intro.md",
		"guide/my file.md",
		"ref/readme.markdown",
		"WikiPage.md",
		"nested/Guide.md",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected links\nwant: %#v\n got: %#v", want, got)
	}
}

func TestExtractLocalMarkdownLinks_IgnoresAbsoluteAndUnknownExt(t *testing.T) {
	content := `
[root](/docs/root.md)
[pdf](./docs/manual.pdf)
[[#Section]]
`

	got := ExtractLocalMarkdownLinks(content, []string{".md"})
	if len(got) != 0 {
		t.Fatalf("expected no links, got: %#v", got)
	}
}
