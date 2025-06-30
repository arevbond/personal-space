package processor

import (
	"fmt"
	"log/slog"
	"regexp"
)

type ImageProcessor struct {
	log *slog.Logger
}

func NewImageProcessor(log *slog.Logger) *ImageProcessor {
	return &ImageProcessor{log: log}
}

func (ir *ImageProcessor) AddPrefix(content []byte, prefix string) ([]byte, error) {
	pattern := `\!\[(?:([^\]]*)\]\(.*?([^/\)]+)\)|\[([^\]]+)\]\])`

	mdImage, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("can't parse pattern: %w", err)
	}

	text := string(content)

	result := mdImage.ReplaceAllStringFunc(text, func(match string) string {
		submatches := mdImage.FindStringSubmatch(match)

		if len(submatches) > 1 {
			if submatches[2] != "" {
				// Standard markdown: ![alt](path/file)
				altText := submatches[1] // Could be empty
				filename := submatches[2]

				return fmt.Sprintf("![%s](%s%s)", altText, prefix, filename)
			} else if submatches[3] != "" {
				// Wiki format: ![[file]]
				filename := submatches[3]

				return fmt.Sprintf("![](%s%s)", prefix, filename)
			}
		}

		return match
	})

	return []byte(result), nil
}
