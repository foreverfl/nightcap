## 무엇을
{{.Summary}}

## 읽는 순서
{{range .Units}}- {{.What}}: {{.Done}}
{{end}}
## 위험
{{range .Units}}{{if .ReadCodeReason}}- 코드까지 읽어야 함: {{.ReadCodeReason}}
{{end}}{{end}}
## 확인한 것
```
{{.CheckTail}}
```
