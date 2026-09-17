package chunker

import (
	"strings"
)

type Splitter struct {
	ChunkSize    int
	ChunkOverlap int
	Separators   []string
}

func NewDefaultSplitter() *Splitter {
	return &Splitter{
		ChunkSize:    1000,
		ChunkOverlap: 200,
		Separators:   []string{"\n\n", "\n", ". ", " ", ""},
	}
}

// SplitText splits large documents recursively by paragraph, newline, sentence, and word
func (s *Splitter) SplitText(text string) []string {
	return s.split(text, s.Separators)
}

func (s *Splitter) split(text string, separators []string) []string {
	var finalChunks []string
	separator := separators[len(separators)-1]
	var newSeparators []string

	for i, sep := range separators {
		if sep == "" || strings.Contains(text, sep) {
			separator = sep
			newSeparators = separators[i+1:]
			break
		}
	}

	var splits []string
	if separator != "" {
		splits = strings.Split(text, separator)
	} else {
		// Character by character
		for _, r := range text {
			splits = append(splits, string(r))
		}
	}

	var currentDoc []string
	total := 0

	for _, piece := range splits {
		pieceLen := len(piece)
		if total+pieceLen > s.ChunkSize {
			if total > 0 {
				doc := strings.Join(currentDoc, separator)
				if strings.TrimSpace(doc) != "" {
					finalChunks = append(finalChunks, strings.TrimSpace(doc))
				}
				// Handle overlap
				for total > s.ChunkOverlap && len(currentDoc) > 0 {
					total -= len(currentDoc[0])
					if len(separator) > 0 && len(currentDoc) > 1 {
						total -= len(separator)
					}
					currentDoc = currentDoc[1:]
				}
			}
		}

		// If a single piece is still larger than chunk size, recurse with next separator
		if pieceLen > s.ChunkSize && len(newSeparators) > 0 {
			subChunks := s.split(piece, newSeparators)
			finalChunks = append(finalChunks, subChunks...)
		} else {
			currentDoc = append(currentDoc, piece)
			total += pieceLen
			if len(currentDoc) > 1 {
				total += len(separator)
			}
		}
	}

	if len(currentDoc) > 0 {
		doc := strings.Join(currentDoc, separator)
		if strings.TrimSpace(doc) != "" {
			finalChunks = append(finalChunks, strings.TrimSpace(doc))
		}
	}

	return finalChunks
}
