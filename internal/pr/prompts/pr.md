## What
{{.Summary}}

## Review order
{{- if .ReviewOrder}}
{{.ReviewOrder}}
{{- else}}
{{- range .Units}}
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
{{- range .Units}}
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
{{.CheckTail}}
```
