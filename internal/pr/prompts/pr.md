## 무엇을
{{.Summary}}

## 읽는 순서
{{- if .ReviewOrder}}
{{.ReviewOrder}}
{{- else}}
{{- range .Units}}
- {{.What}}: {{.Done}}
{{- end}}
{{- end}}

## 위험
{{- if .IsMissingEnglish}}
- 영문 없음
{{- end}}
{{- if .IsMissingReviewOrder}}
- 읽는 순서 없음
{{- end}}
{{- range .Units}}
  {{- if .ReadCodeReason}}
- 코드까지 읽어야 함: {{.ReadCodeReason}}
  {{- end}}
{{- end}}
{{- range .Risks}}
- {{.}}
{{- end}}
{{- if .HasNoRisk}}
- 없음
{{- end}}

## 확인한 것
```
{{.CheckTail}}
```
