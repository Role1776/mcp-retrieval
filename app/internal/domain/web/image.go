package web

import (
	"strings"

	"github.com/google/uuid"

	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
)

type Image struct {
	id          uuid.UUID
	url         string
	pageURL     string
	description string
}

type ImageProps struct {
	URL         string `validate:"required,url"`
	PageURL     string `validate:"required,url"`
	Description string `validate:"omitempty"`
}

func NewImage(props ImageProps) (Image, error) {
	if err := validator.Validate(props); err != nil {
		return Image{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Image{}, err
	}

	return Image{
		id:          id,
		url:         props.URL,
		pageURL:     props.PageURL,
		description: props.Description,
	}, nil
}

type ImageView struct {
	ID          uuid.UUID
	URL         string
	PageURL     string
	Description string
}

func (i Image) View() ImageView {
	return ImageView{
		ID:          i.id,
		URL:         i.url,
		PageURL:     i.pageURL,
		Description: i.description,
	}
}

func (i Image) Markdown() string {
	alt := i.description
	if alt == "" {
		alt = "image"
	}

	return "![" + alt + "](" + i.url + ")"
}

type Images []Image

func (i Images) Len() int {
	return len(i)
}

func (i Images) IsEmpty() bool {
	return len(i) == 0
}

func (i Images) Views() []ImageView {
	views := make([]ImageView, 0, len(i))
	for _, image := range i {
		views = append(views, image.View())
	}

	return views
}

func (i Images) Limit(n int) Images {
	if n <= 0 || len(i) == 0 {
		return Images{}
	}

	result := make(Images, min(len(i), n))
	copy(result, i)

	return result
}

func (i Images) Dedupe() Images {
	seen := make(map[string]struct{}, len(i))
	result := make(Images, 0, len(i))

	for _, image := range i {
		if _, ok := seen[image.url]; ok {
			continue
		}
		seen[image.url] = struct{}{}
		result = append(result, image)
	}

	return result
}

func (i Images) Markdown() string {
	parts := make([]string, 0, len(i))
	for _, image := range i {
		parts = append(parts, image.Markdown())
	}

	return strings.Join(parts, "\n")
}
