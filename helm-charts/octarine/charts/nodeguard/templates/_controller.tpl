{{/*
Create nodeguard controller name
*/}}
{{- define "nodeguard.controller.name" -}}
{{- printf "%s-controller" (include "nodeguard.name" .) -}}
{{- end -}}

{{/*
Create nodeguard controller fullname
*/}}
{{- define "nodeguard.controller.fullname" -}}
{{- printf "%s-controller" (include "nodeguard.fullname" .) -}}
{{- end -}}

{{/*
Controller labels
*/}}
{{- define "nodeguard.controller.labels" -}}
{{- template "nodeguard.labels.tpl" (list . (include "nodeguard.controller.name" .)) -}}
{{- end -}}

{{/*
Create service full address for nodeguard controller service.
*/}}
{{- define "nodeguard.controller.service.address" -}}
{{- default ( printf "%s.%s.svc.cluster.local" (include "nodeguard.controller.fullname" .) .Release.Namespace ) -}}
{{- end -}}

{{/*
Create configmap name for nodeguard controller env vars.
*/}}
{{- define "nodeguard.controller.configmap.env.fullname" -}}
{{-  printf "%s-env" (include "nodeguard.controller.fullname" .) -}}
{{- end -}}

{{/*
Generate env vars for nodeguard controller.
Const env vars are taken from the values, dynamic env vars are generated here.
*/}}
{{- define "nodeguard.controller.env" -}}
{{ toYaml .Values.controller.env }}
OCTARINE_NODEGUARD_PROMETHEUS_PORT: {{ .Values.controller.prometheus.port | quote }}
OCTARINE_NODEGUARD_PROBES_PORT: {{ .Values.controller.probes.port | quote }}
{{- end -}}