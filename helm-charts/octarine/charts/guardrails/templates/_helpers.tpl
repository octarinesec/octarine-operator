{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "guardrails.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "guardrails.fullname" -}}
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
{{- define "guardrails.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "guardrails.labels" -}}
app.kubernetes.io/name: {{ include "guardrails.name" . }}
helm.sh/chart: {{ include "guardrails.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "guardrails.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "guardrails.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create service name for guardrails service.
*/}}
{{- define "guardrails.service.name" -}}
{{- default (include "guardrails.fullname" .) .Values.service.name -}}
{{- end -}}

{{/*
Create service fullname for guardrails service. Used also as the webhook name.
*/}}
{{- define "guardrails.service.fullname" -}}
{{- default ( printf "%s.%s.svc" (include "guardrails.service.name" .) .Release.Namespace ) -}}
{{- end -}}

{{/*
Create name for guardrails validating webhook.
*/}}
{{- define "guardrails.validatingwebhook.name" -}}
{{- default (include "guardrails.fullname" .) .Values.admissionController.name -}}
{{- end -}}

{{/*
Create name for guardrails TLS secret.
*/}}
{{- define "guardrails.tls.secret.name" -}}
{{- default ( printf "%s-tls" (include "guardrails.fullname" .) ) .Values.admissionController.secret.name -}}
{{- end -}}

{{/*
Create webhook fullname for guardrails namespaces webhook.
*/}}
{{- define "guardrails-namespaces.webhook.fullname" -}}
{{- default ( printf "%s-namespaces.%s.svc" (include "guardrails.fullname" .) .Release.Namespace ) -}}
{{- end -}}

{{/*
Create configmap name for guardrails env vars.
*/}}
{{- define "guardrails.configmap.env.fullname" -}}
{{-  printf "%s-env" (include "guardrails.fullname" .) -}}
{{- end -}}

{{/*
Generate env vars for guardrails.
Const env vars are taken from the values, dynamic env vars are generated here.
*/}}
{{- define "guardrails.env" -}}
{{ toYaml .Values.env }}
OCTARINE_GUARDRAIL_SERVICE_PORT: {{ .Values.service.port | quote }}
OCTARINE_GUARDRAIL_SERVICE_PROMETHEUS_PORT: {{ .Values.prometheus.port | quote }}
OCTARINE_GUARDRAIL_SERVICE_PROBES_PORT: {{ .Values.probes.port | quote }}
{{- end -}}

{{- define "guardrails.webhook.timeout" -}}
{{- if semverCompare ">=1.14" .Capabilities.KubeVersion.GitVersion -}}
timeoutSeconds: {{ .Values.admissionController.timeoutSeconds }}
{{- end -}}
{{- end -}}

{{/*
Generate certificates for admission-controller webhooks
*/}}
{{- define "guardrails.gen-certs" -}}
{{- $expiration := (.Values.admissionController.CA.expiration | int) -}}
{{- if (or (empty .Values.admissionController.CA.cert) (empty .Values.admissionController.CA.key)) -}}
{{- $ca :=  genCA "guardrails-ca" $expiration -}}
{{- template "guardrails.gen-client-tls" (dict "RootScope" . "CA" $ca) -}}
{{- else -}}
{{- $ca :=  buildCustomCert (.Values.admissionController.CA.cert | b64enc) (.Values.admissionController.CA.key | b64enc) -}}
{{- template "guardrails.gen-client-tls" (dict "RootScope" . "CA" $ca) -}}
{{- end -}}
{{- end -}}

{{/*
Generate client key and cert from CA
*/}}
{{- define "guardrails.gen-client-tls" -}}
{{- $altNames := list ( include "guardrails.service.fullname" .RootScope) -}}
{{- $expiration := (.RootScope.Values.admissionController.CA.expiration | int) -}}
{{- $cert := genSignedCert ( include "guardrails.fullname" .RootScope) nil $altNames $expiration .CA -}}
{{- $clientCert := default $cert.Cert .RootScope.Values.admissionController.secret.cert | b64enc -}}
{{- $clientKey := default $cert.Key .RootScope.Values.admissionController.secret.key | b64enc -}}
caCert: {{ .CA.Cert | b64enc }}
clientCert: {{ $clientCert }}
clientKey: {{ $clientKey }}
{{- end -}}