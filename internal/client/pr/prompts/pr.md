## What
{{.Summary}}

## Review order
{{- if .ReviewOrder}}
{{.ReviewOrder}}
{{- else}}
{{- range .Commits}}
- {{.What}}: {{.Done}}
{{- end}}
{{- end}}

## Risks
{{- if .IsMissingEnglish}}
- 영어 블록 없음(번역 실패)
{{- end}}
{{- if .IsMissingReviewOrder}}
- 읽는 순서 없음
{{- end}}
{{- range .Commits}}
  {{- if .ReadCodeReason}}
- {{.ReadCodeReason}}
  {{- end}}
{{- end}}
{{- range .Risks}}
- {{.}}
{{- end}}
{{- if .HasNoRisk}}
- 없음
{{- end}}

## Verified
```
{{.TestTail}}
```
