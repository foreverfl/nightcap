package claude

import (
	"cmp"
	"encoding/json"
	"strings"
)

const screenLineWidth = 80

// ScreenLines turns one stream-json line into what nightcap prints while claude works:
// `Edit internal/feed/feed.go` per tool call and the quoted first line per text of an assistant message, nothing for any other line.
// A path under worktreeDir is shown relative to it.
func ScreenLines(raw []byte, worktreeDir string) []string {
	var streamLine struct {
		Type    string `json:"type"`
		Message struct {
			Content []struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Name  string `json:"name"`
				Input struct {
					FilePath string `json:"file_path"`
					Command  string `json:"command"`
					Pattern  string `json:"pattern"`
				} `json:"input"`
			} `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &streamLine); err != nil || streamLine.Type != "assistant" {
		return nil
	}
	var screenLines []string
	for _, content := range streamLine.Message.Content {
		switch content.Type {
		case "tool_use":
			target := cmp.Or(content.Input.FilePath, content.Input.Command, content.Input.Pattern)
			target = strings.TrimPrefix(target, worktreeDir+"/")
			screenLines = append(screenLines, firstLineCut(strings.TrimSpace(content.Name+" "+target)))
		case "text":
			screenLines = append(screenLines, `"`+firstLineCut(content.Text)+`"`)
		}
	}
	return screenLines
}

// firstLineCut returns the first line of text, cut to screenLineWidth characters.
func firstLineCut(text string) string {
	firstLine, _, _ := strings.Cut(text, "\n")
	characters := []rune(firstLine)
	if len(characters) > screenLineWidth {
		return string(characters[:screenLineWidth])
	}
	return firstLine
}
