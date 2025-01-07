package notion

import (
	"fmt"

	gn "github.com/dstotijn/go-notion"
)

type HeadingBlock struct {
	Tag      string
	RichText []gn.RichText
}

func ProcessParagraphBlock(b *gn.ParagraphBlock) string {
	html := "<p>"
	for _, text := range b.RichText {
		html += fmt.Sprintf("<p>%s</p>", text.Text.Content)
	}
	html += "</p>"

	return html
}

func ProcessHeadingBlock(b HeadingBlock) string {
	html := fmt.Sprintf("<%s>", b.Tag)
	for _, text := range b.RichText {
		html += text.Text.Content
	}
	html += fmt.Sprintf("</%s>", b.Tag)

	return html
}

func ProcessBulletedListItemBlock(b *gn.BulletedListItemBlock) string {
	html := "<ul><li>"
	for _, text := range b.RichText {
		html += text.Text.Content
	}
	html += "</li></ul>"

	return html
}

func ProcessNumberedListItemBlock(b *gn.NumberedListItemBlock) string {
	html := "<ol><li>"
	for _, text := range b.RichText {
		html += text.Text.Content
	}
	html += "</li></ol>"

	return html
}

func ProcessImageBlock(b *gn.ImageBlock) string {
	if b.File != nil {
		return fmt.Sprintf(`<img src="%s" alt="Image" />`, b.File.URL)
	} else if b.External != nil {
		return fmt.Sprintf(`<img src="%s" alt="Image" />`, b.External.URL)
	}

	return ""
}

func ProcessTableOfContentsBlock() string {
	return "<div><strong>Table of Contents</strong></div>"
}

func ProcessQuoteBlock(b *gn.QuoteBlock) string {
	html := ""
	for _, text := range b.RichText {
		html += fmt.Sprintf("<blockquote>%s</blockquote>", text.Text.Content)
	}

	return html
}

func ProcessDividerBlock() string {
	return "<hr/>"
}

func ProcessBlock(block gn.Block) string {
	switch b := block.(type) {
	case *gn.ParagraphBlock:
		return ProcessParagraphBlock(b)
	case *gn.Heading1Block:
		return ProcessHeadingBlock(HeadingBlock{Tag: "h1", RichText: b.RichText})
	case *gn.Heading2Block:
		return ProcessHeadingBlock(HeadingBlock{Tag: "h2", RichText: b.RichText})
	case *gn.Heading3Block:
		return ProcessHeadingBlock(HeadingBlock{Tag: "h3", RichText: b.RichText})
	case *gn.BulletedListItemBlock:
		return ProcessBulletedListItemBlock(b)
	case *gn.NumberedListItemBlock:
		return ProcessNumberedListItemBlock(b)
	case *gn.ImageBlock:
		return ProcessImageBlock(b)
	case *gn.TableOfContentsBlock:
		return ProcessTableOfContentsBlock()
	case *gn.QuoteBlock:
		return ProcessQuoteBlock(b)
	case *gn.DividerBlock:
		return ProcessDividerBlock()
	default:
		fmt.Printf("Unsupported block type: %T\n", b)
		return ""
	}
}
