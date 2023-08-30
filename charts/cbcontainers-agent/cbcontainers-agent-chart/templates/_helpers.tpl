{{/* Get the name of the secret that contains the access token */}}
{{- define "cbcontainers-agent.access-token-name" -}}
{{- $secret := . -}}
{{- if $secret -}}
"{{- $secret -}}"
{{- else -}}
"cbcontainers-access-token"
{{- end -}}
{{- end -}}

{{/* Get the name of the secret that contains the company code */}}
{{- define "cbcontainers-agent.company-code-name" -}}
{{- $secret := . -}}
{{- if $secret -}}
"{{- $secret -}}"
{{- else -}}
"cbcontainers-company-code"
{{- end -}}
{{- end -}}


