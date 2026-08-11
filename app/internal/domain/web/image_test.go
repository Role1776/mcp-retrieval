package web

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewImage(t *testing.T) {
	cases := []struct {
		name  string
		props ImageProps

		want    Image
		wantErr bool
	}{
		{
			name: "full",
			props: ImageProps{
				URL:         "https://example.com/a.png",
				PageURL:     "https://example.com/page",
				Description: "a picture",
			},
			want: Image{
				URL:         "https://example.com/a.png",
				PageURL:     "https://example.com/page",
				Description: "a picture",
			},
		},
		{
			name:  "description is optional",
			props: ImageProps{URL: "https://example.com/a.png", PageURL: "https://example.com/page"},
			want:  Image{URL: "https://example.com/a.png", PageURL: "https://example.com/page"},
		},
		{
			name:    "missing url",
			props:   ImageProps{PageURL: "https://example.com/page", Description: "a picture"},
			wantErr: true,
		},
		{
			name:    "missing page url",
			props:   ImageProps{URL: "https://example.com/a.png", Description: "a picture"},
			wantErr: true,
		},
		{name: "empty props", props: ImageProps{}, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			image, err := NewImage(tc.props)

			// assert
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, Image{}, image)

				return
			}

			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, image.ID)

			image.ID = uuid.Nil
			assert.Equal(t, tc.want, image)
		})
	}
}

func TestImageMarkdown(t *testing.T) {
	cases := []struct {
		name  string
		image Image

		want string
	}{
		{
			name:  "description becomes the alt text",
			image: Image{URL: "https://example.com/a.png", Description: "a picture"},
			want:  "![a picture](https://example.com/a.png)",
		},
		{
			name:  "empty description falls back to a placeholder",
			image: Image{URL: "https://example.com/a.png"},
			want:  "![image](https://example.com/a.png)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.image.Markdown())
		})
	}
}

func TestImagesLen(t *testing.T) {
	cases := []struct {
		name   string
		images Images

		want int
	}{
		{name: "nil", images: nil, want: 0},
		{name: "empty", images: Images{}, want: 0},
		{name: "two", images: Images{{URL: "a"}, {URL: "b"}}, want: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.images.Len())
		})
	}
}

func TestImagesIsEmpty(t *testing.T) {
	cases := []struct {
		name   string
		images Images

		want bool
	}{
		{name: "nil", images: nil, want: true},
		{name: "empty", images: Images{}, want: true},
		{name: "one", images: Images{{URL: "a"}}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.images.IsEmpty())
		})
	}
}

func TestImagesLimit(t *testing.T) {
	three := Images{{URL: "a"}, {URL: "b"}, {URL: "c"}}

	cases := []struct {
		name   string
		images Images
		n      int

		want Images
	}{
		{name: "keeps the first n", images: three, n: 2, want: Images{{URL: "a"}, {URL: "b"}}},
		{name: "n above the length keeps everything", images: three, n: 10, want: three},
		{name: "n equal to the length keeps everything", images: three, n: 3, want: three},
		{name: "zero n", images: three, n: 0, want: Images{}},
		{name: "negative n", images: three, n: -1, want: Images{}},
		{name: "empty receiver", images: Images{}, n: 2, want: Images{}},
		{name: "nil receiver", images: nil, n: 2, want: Images{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			limited := tc.images.Limit(tc.n)

			// assert
			assert.Equal(t, tc.want, limited)

			if len(limited) > 0 {
				limited[0].URL = "mutated"
				assert.NotEqual(t, "mutated", tc.images[0].URL, "the result must not alias the receiver")
			}
		})
	}
}

func TestImagesDedupe(t *testing.T) {
	cases := []struct {
		name   string
		images Images

		want Images
	}{
		{
			name:   "keeps the first occurrence of a url",
			images: Images{{URL: "a", Description: "first"}, {URL: "b"}, {URL: "a", Description: "second"}},
			want:   Images{{URL: "a", Description: "first"}, {URL: "b"}},
		},
		{
			name:   "the same url on different pages collapses",
			images: Images{{URL: "a", PageURL: "p1"}, {URL: "a", PageURL: "p2"}},
			want:   Images{{URL: "a", PageURL: "p1"}},
		},
		{
			name:   "nothing to dedupe",
			images: Images{{URL: "a"}, {URL: "b"}},
			want:   Images{{URL: "a"}, {URL: "b"}},
		},
		{name: "empty receiver", images: Images{}, want: Images{}},
		{name: "nil receiver", images: nil, want: Images{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.images.Dedupe())
		})
	}
}

func TestImagesMarkdown(t *testing.T) {
	cases := []struct {
		name   string
		images Images

		want string
	}{
		{
			name: "entries are separated by a newline",
			images: Images{
				{URL: "https://example.com/a.png", Description: "a"},
				{URL: "https://example.com/b.png"},
			},
			want: "![a](https://example.com/a.png)\n![image](https://example.com/b.png)",
		},
		{
			name:   "single entry has no separator",
			images: Images{{URL: "https://example.com/a.png", Description: "a"}},
			want:   "![a](https://example.com/a.png)",
		},
		{name: "empty receiver", images: Images{}, want: ""},
		{name: "nil receiver", images: nil, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// act & assert
			assert.Equal(t, tc.want, tc.images.Markdown())
		})
	}
}
