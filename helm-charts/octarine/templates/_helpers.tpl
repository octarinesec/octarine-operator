{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
Using a const name as it will be used by the sub-charts as well (and .Chart params aren't available to sub-charts).
*/}}
{{- define "octarine.name" -}}
{{- "octarine" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "octarine.fullname" -}}
{{- $name := include "octarine.name" . -}}
{{- if contains $name $.Release.Name -}}
{{- $.Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" $.Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "octarine.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "octarine.labels" -}}
app.kubernetes.io/name: {{ include "octarine.name" . }}
helm.sh/chart: {{ include "octarine.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create secret name for access token.
If provided by the user - use it, otherwise generate.
*/}}
{{- define "octarine.accesstoken.secret.name" -}}
{{- if .Values.global.octarine.accessTokenSecret }}
{{- .Values.global.octarine.accessTokenSecret -}}
{{- else }}
{{- printf "%s-accesstoken" (include "octarine.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Create configmap name for octarine global env vars.
*/}}
{{- define "octarine.configmap.env.fullname" -}}
{{- printf "%s-env" (include "octarine.fullname" .) -}}
{{- end -}}

{{/*
Generate global env vars for octarine dataplane components.
*/}}
{{- define "octarine.env" -}}
OCTARINE_ACCOUNT: {{ required "A valid .Values.global.octarine.account is required" .Values.global.octarine.account | quote }}
OCTARINE_DOMAIN: {{ required "A valid .Values.global.octarine.domain is required" .Values.global.octarine.domain | quote }}
OCTARINE_API_ADAPTER_NAME: {{ required "A valid .Values.global.octarine.api.adapterName is required" .Values.global.octarine.api.adapterName | quote }}
OCTARINE_API_HOST: {{ required "A valid .Values.global.octarine.api.host is required" .Values.global.octarine.api.host | quote }}
OCTARINE_API_PORT: {{ required "A valid .Values.global.octarine.api.port is required" .Values.global.octarine.api.port | quote }}
OCTARINE_MESSAGEPROXY_HOST: {{ required "A valid .Values.global.octarine.messageproxy.host is required" .Values.global.octarine.messageproxy.host | quote }}
OCTARINE_MESSAGEPROXY_PORT: {{ required "A valid .Values.global.octarine.messageproxy.port is required" .Values.global.octarine.messageproxy.port | quote }}
{{- end -}}

{{/*
envFrom value for common octarine env config
*/}}
{{- define "octarine.common.envFrom" -}}
- configMapRef:
    name: {{ include "octarine.configmap.env.fullname" . }}
{{- end -}}

{{/*
env value for common octarine env config
*/}}
{{- define "octarine.common.env" -}}
- name: OCTARINE_ACCESS_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "octarine.accesstoken.secret.name" . }}
      key: accessToken
{{- end -}}

{{/*
Image pull secret name.
If provided by the user - use it, otherwise generate.
*/}}
{{- define "octarine.imagePullSecret.name" -}}
{{- if .Values.global.imagePullSecret }}
{{- .Values.global.imagePullSecret -}}
{{- else }}
{{- printf "%s-registry-secret" (include "octarine.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Template for generating a docker image pull secret.
*/}}
{{- define "imagePullSecret.tpl" }}
{{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" (required "A valid .Values.imageCredentials.registry entry required!" .Values.imageCredentials.registry) (printf "%s:%s" (required "A valid .Values.imageCredentials.username entry required!" .Values.imageCredentials.username) (required "A valid .Values.imageCredentials.password entry required!" .Values.imageCredentials.password) | b64enc) | b64enc }}
{{- end }}

{{/*
Priority class name
*/}}
{{- define "octarine.priorityClass.name" -}}
{{- printf "%s-priority" (include "octarine.fullname" .) -}}
{{- end }}

{{/*
Determine priority class apiVersion by K8s version
*/}}
{{- define "octarine.priorityClass.apiVersion" -}}
{{- if semverCompare ">=1.14.0-0" .Capabilities.KubeVersion.GitVersion -}}
scheduling.k8s.io/v1
{{- else if semverCompare ">=1.11.0-0" .Capabilities.KubeVersion.GitVersion -}}
scheduling.k8s.io/v1beta1
{{- else -}}
scheduling.k8s.io/v1alpha1
{{- end -}}
{{- end -}}