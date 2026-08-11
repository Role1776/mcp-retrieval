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
	ID        uuid.UUID `json:"id" validate:"required"`
	Title     string    `json:"title" validate:"required"`
	Byline    string    `json:"byline" validate:"omitempty"`
	Markdown  string    `json:"markdown" validate:"required"`
	Length    int       `json:"length" validate:"gte=0"`
	Excerpt   string    `json:"excerpt" validate:"omitempty"`
	SiteName  string    `json:"site_name" validate:"omitempty"`
	MainImage string    `json:"main_image" validate:"omitempty"`
	AllImages []string  `json:"all_images" validate:"omitempty"`
	Favicon   string    `json:"favicon" validate:"omitempty"`
	Language  string    `json:"language" validate:"omitempty"`
	Truncated bool      `json:"truncated"`
}

type DocumentProps struct {
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
}

func NewDocument(props DocumentProps) (Document, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		ID:        id,
		Title:     props.Title,
		Byline:    props.Byline,
		Markdown:  props.Markdown,
		Length:    props.Length,
		Excerpt:   props.Excerpt,
		SiteName:  props.SiteName,
		MainImage: props.MainImage,
		AllImages: props.AllImages,
		Favicon:   props.Favicon,
		Language:  props.Language,
	}

	if err := validator.Validate(doc); err != nil {
		return Document{}, err
	}

	return doc, nil
}

func (d *Document) IsEmpty() bool {
	return strings.TrimSpace(d.Markdown) == ""
}

func (d *Document) RemoveAllLinks() {
	if strings.Contains(d.Markdown, "](") {
		d.Markdown = markdownLinkRegex.ReplaceAllString(d.Markdown, "$1")
		d.Length = utf8.RuneCountInString(d.Markdown)
	}
}

func (d *Document) Truncate(limit int) error {
	if limit <= 0 || utf8.RuneCountInString(d.Markdown) <= limit {
		return nil
	}

	chunks, err := textcut.SplitText(d.Markdown, limit, 0)
	if err != nil {
		return err
	}

	if len(chunks) == 0 {
		return domain.ErrNoChunks
	}

	d.Markdown = chunks[0] + truncationSuffix
	d.Length = utf8.RuneCountInString(d.Markdown)
	d.Truncated = true

	return nil
}

func (d *Document) Text() string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(d.Title)

	if d.SiteName != "" {
		b.WriteString("\nsource: ")
		b.WriteString(d.SiteName)
	}

	if d.Byline != "" {
		b.WriteString("\nauthor: ")
		b.WriteString(d.Byline)
	}

	b.WriteString("\n\n")
	b.WriteString(d.Markdown)

	return b.String()
}
