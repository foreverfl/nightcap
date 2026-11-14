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
- no English block
{{- end}}
{{- if .IsMissingReviewOrder}}
- no review order
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
- none
{{- end}}

## Verified
```
{{.TestTail}}
```
