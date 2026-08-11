package web

import (
	"strings"

	"github.com/google/uuid"

	"github.com/Role1776/mcp-retrieval/app/internal/pkg/validator"
)

type Image struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	URL         string    `json:"url" validate:"required"`
	PageURL     string    `json:"page_url" validate:"required"`
	Description string    `json:"description" validate:"omitempty"`
}

type ImageProps struct {
	URL         string
	PageURL     string
	Description string
}

func NewImage(props ImageProps) (Image, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Image{}, err
	}

	image := Image{
		ID:          id,
		URL:         props.URL,
		PageURL:     props.PageURL,
		Description: props.Description,
	}

	if err := validator.Validate(image); err != nil {
		return Image{}, err
	}

	return image, nil
}

func (i Image) Markdown() string {
	alt := i.Description
	if alt == "" {
		alt = "image"
	}

	return "![" + alt + "](" + i.URL + ")"
}

type Images []Image

func (i Images) Len() int {
	return len(i)
}

func (i Images) IsEmpty() bool {
	return len(i) == 0
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
		if _, ok := seen[image.URL]; ok {
			continue
		}
		seen[image.URL] = struct{}{}
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
