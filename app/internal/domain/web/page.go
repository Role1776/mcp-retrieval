package web

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Role1776/mcp-retrieval/app/internal/domain"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/textcut"
	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
)

var markdownLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)

const truncationSuffix = "\n\n[truncated]"

type Document struct {
	id        uuid.UUID
	title     string
	byline    string
	markdown  string
	length    int
	excerpt   string
	siteName  string
	mainImage string
	allImages []string
	favicon   string
	language  string
	truncated bool
}

type DocumentProps struct {
	Title     string   `validate:"required"`
	Byline    string   `validate:"omitempty"`
	Markdown  string   `validate:"required"`
	Length    int      `validate:"gte=0"`
	Excerpt   string   `validate:"omitempty"`
	SiteName  string   `validate:"omitempty"`
	MainImage string   `validate:"omitempty,url"`
	AllImages []string `validate:"omitempty,dive,url"`
	Favicon   string   `validate:"omitempty,url"`
	Language  string   `validate:"omitempty"`
}

func NewDocument(props DocumentProps) (Document, error) {
	if err := validator.Validate(props); err != nil {
		return Document{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Document{}, err
	}

	return Document{
		id:        id,
		title:     props.Title,
		byline:    props.Byline,
		markdown:  props.Markdown,
		length:    props.Length,
		excerpt:   props.Excerpt,
		siteName:  props.SiteName,
		mainImage: props.MainImage,
		allImages: props.AllImages,
		favicon:   props.Favicon,
		language:  props.Language,
	}, nil
}

type DocumentView struct {
	ID        uuid.UUID
	Title     string
	Byline    string
	Markdown  string
	Length    int
	Excerpt   string
	SiteName  string
	MainImage string
	AllImages []string
	Favicon   string
	Language  string
	Truncated bool
}

func (d Document) View() DocumentView {
	var imagesCopy []string
	if d.allImages != nil {
		imagesCopy = make([]string, len(d.allImages))
		copy(imagesCopy, d.allImages)
	}

	return DocumentView{
		ID:        d.id,
		Title:     d.title,
		Byline:    d.byline,
		Markdown:  d.markdown,
		Length:    d.length,
		Excerpt:   d.excerpt,
		SiteName:  d.siteName,
		MainImage: d.mainImage,
		AllImages: imagesCopy,
		Favicon:   d.favicon,
		Language:  d.language,
		Truncated: d.truncated,
	}
}

func (d *Document) IsEmpty() bool {
	return strings.TrimSpace(d.markdown) == ""
}

func (d *Document) RemoveAllLinks() {
	if strings.Contains(d.markdown, "](") {
		d.markdown = markdownLinkRegex.ReplaceAllString(d.markdown, "$1")
		d.length = utf8.RuneCountInString(d.markdown)
	}
}

func (d *Document) Truncate(limit int) error {
	if limit <= 0 || utf8.RuneCountInString(d.markdown) <= limit {
		return nil
	}

	chunks, err := textcut.SplitText(d.markdown, limit, 0)
	if err != nil {
		return err
	}

	if len(chunks) == 0 {
		return domain.ErrNoChunks
	}

	d.markdown = chunks[0] + truncationSuffix
	d.length = utf8.RuneCountInString(d.markdown)
	d.truncated = true

	return nil
}

func (d *Document) Text() string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(d.title)

	if d.siteName != "" {
		b.WriteString("\nsource: ")
		b.WriteString(d.siteName)
	}

	if d.byline != "" {
		b.WriteString("\nauthor: ")
		b.WriteString(d.byline)
	}

	b.WriteString("\n\n")
	b.WriteString(d.markdown)

	return b.String()
}
