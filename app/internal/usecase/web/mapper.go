package web

import (
	"github.com/Role1776/mcp-retrieval/app/internal/domain/web"
	dto "github.com/Role1776/mcp-retrieval/app/internal/dto/web"
)

func toDocument(doc web.Document) dto.Document {
	v := doc.View()

	return dto.Document{
		ID:        v.ID,
		Title:     v.Title,
		Byline:    v.Byline,
		Markdown:  v.Markdown,
		Length:    v.Length,
		Excerpt:   v.Excerpt,
		SiteName:  v.SiteName,
		MainImage: v.MainImage,
		AllImages: v.AllImages,
		Favicon:   v.Favicon,
		Language:  v.Language,
		Truncated: v.Truncated,
	}
}

func toSnippets(snippets web.Snippets) []dto.Snippet {
	result := make([]dto.Snippet, 0, snippets.Len())
	for _, v := range snippets.Views() {
		result = append(result, dto.Snippet{
			ID:      v.ID,
			Link:    v.Link,
			Title:   v.Title,
			Rank:    v.Rank,
			Source:  v.Source,
			Snippet: v.Snippet,
			Favicon: v.Favicon,
		})
	}

	return result
}

func toImages(images web.Images) []dto.Image {
	result := make([]dto.Image, 0, images.Len())
	for _, v := range images.Views() {
		result = append(result, dto.Image{
			ID:          v.ID,
			URL:         v.URL,
			PageURL:     v.PageURL,
			Description: v.Description,
		})
	}

	return result
}
