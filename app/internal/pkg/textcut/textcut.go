package textcut

import "github.com/tmc/langchaingo/textsplitter"

var separators = []string{
	"\n\n",
	"\n",
	". ",
	"? ",
	"! ",
	", ",
	" ",
	"",
}

func SplitText(text string, chunkSize, chunkOverlap int) ([]string, error) {
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
		textsplitter.WithSeparators(separators),
	)

	return splitter.SplitText(text)
}
