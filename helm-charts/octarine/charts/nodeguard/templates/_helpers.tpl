{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "nodeguard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nodeguard.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nodeguard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Nodeguard labels template, to be used by nodeguard components.

This takes an array of two values:
- the top context
- the name of the component
*/}}
{{- define "nodeguard.labels.tpl" -}}
{{- $context := first . | default . -}}
{{- $name := index . 1 | default (include "nodeguard.name" $context) -}}
{{- with $context -}}
app.kubernetes.io/name: {{ $name }}
helm.sh/chart: {{ include "nodeguard.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
{{- end -}}

{{/*
Nodeguard labels
*/}}
{{- define "nodeguard.labels" -}}
{{- template "nodeguard.labels.tpl" (list . (include "nodeguard.name" .)) -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "nodeguard.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "nodeguard.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}