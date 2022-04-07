{{/* Get the desired version of the api extension based on the Kubernetes version used */}}
{{- define "cbcontainers-agent.api-version" -}}
{{ if lt (atoi .Capabilities.KubeVersion.Minor) 16 -}}
v1beta1
{{- else -}}
v1
{{- end }}
{{- end }}
