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
Guardrails labels template, to be used by Guardrails components.

This takes an array of two values:
- the top context
- the name of the component
*/}}
{{- define "guardrails.labels.tpl" -}}
{{- $context := first . | default . -}}
{{- $name := index . 1 | default (include "guardrails.name" $context) -}}
{{- with $context -}}
app.kubernetes.io/name: {{ $name }}
helm.sh/chart: {{ include "guardrails.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
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
Create guardrails enforcer name
*/}}
{{- define "guardrails.enforcer.name" -}}
{{- printf "%s-enforcer" (include "guardrails.name" .) -}}
{{- end -}}

{{/*
Create guardrails enforcer fullname
*/}}
{{- define "guardrails.enforcer.fullname" -}}
{{- printf "%s-enforcer" (include "guardrails.fullname" .) -}}
{{- end -}}

{{/*
Enforcer labels
*/}}
{{- define "guardrails.enforcer.labels" -}}
{{- template "guardrails.labels.tpl" (list . (include "guardrails.enforcer.name" .)) -}}
{{- end -}}

{{/*
Create configmap name for guardrails enforcer env vars.
*/}}
{{- define "guardrails.enforcer.configmap.env.fullname" -}}
{{-  printf "%s-env" (include "guardrails.enforcer.fullname" .) -}}
{{- end -}}

{{/*
Generate env vars for enforcer.
Const env vars are taken from the values, dynamic env vars are generated here.
*/}}
{{- define "guardrails.enforcer.env" -}}
{{ toYaml .Values.enforcer.env }}
{{- end -}}

{{/*
Create service name for guardrails enforcer service.
*/}}
{{- define "guardrails.enforcer.service.name" -}}
{{- include "guardrails.enforcer.fullname" . -}}
{{- end -}}

{{/*
Create service fullname for guardrails enforcer service. Used also as the webhook name.
*/}}
{{- define "guardrails.enforcer.service.fullname" -}}
{{- default ( printf "%s.%s.svc" (include "guardrails.enforcer.service.name" .) .Release.Namespace ) -}}
{{- end -}}

{{/*
Create name for guardrails enforcer validating webhook.
*/}}
{{- define "guardrails.enforcer.validatingwebhook.name" -}}
{{- default (include "guardrails.enforcer.fullname" .) .Values.enforcer.admissionController.name -}}
{{- end -}}

{{/*
Create webhook fullname for guardrails namespaces webhook.
*/}}
{{- define "guardrails.enforcer.namespaces.webhook.fullname" -}}
{{- default ( printf "%s-namespaces.%s.svc" (include "guardrails.enforcer.fullname" .) .Release.Namespace ) -}}
{{- end -}}

{{/*
Create name for guardrails enforcer TLS secret.
*/}}
{{- define "guardrails.enforcer.tls.secret.name" -}}
{{- default ( printf "%s-tls" (include "guardrails.enforcer.fullname" .) ) .Values.enforcer.admissionController.secret.name -}}
{{- end -}}

{{- define "guardrails.webhook.timeout" -}}
{{- if semverCompare ">=1.14" .Capabilities.KubeVersion.GitVersion -}}
timeoutSeconds: {{ .Values.admissionController.timeoutSeconds }}
{{- end -}}
{{- end -}}

{{/*
Generate certificates for admission-controller webhooks
*/}}
{{- define "guardrails.enforcer.gen-certs" -}}
{{- $expiration := (.Values.enforcer.admissionController.CA.expiration | int) -}}
{{- if (or (empty .Values.enforcer.admissionController.CA.cert) (empty .Values.enforcer.admissionController.CA.key)) -}}
{{- $ca :=  genCA "guardrails-enforcer-ca" $expiration -}}
{{- template "guardrails.enforcer.gen-client-tls" (dict "RootScope" . "CA" $ca) -}}
{{- else -}}
{{- $ca :=  buildCustomCert (.Values.enforcer.admissionController.CA.cert | b64enc) (.Values.enforcer.admissionController.CA.key | b64enc) -}}
{{- template "guardrails.enforcer.gen-client-tls" (dict "RootScope" . "CA" $ca) -}}
{{- end -}}
{{- end -}}

{{/*
Generate client key and cert from CA
*/}}
{{- define "guardrails.enforcer.gen-client-tls" -}}
{{- $altNames := list ( include "guardrails.enforcer.service.fullname" .RootScope) -}}
{{- $expiration := (.RootScope.Values.enforcer.admissionController.CA.expiration | int) -}}
{{- $cert := genSignedCert ( include "guardrails.enforcer.fullname" .RootScope) nil $altNames $expiration .CA -}}
{{- $clientCert := default $cert.Cert .RootScope.Values.enforcer.admissionController.secret.cert | b64enc -}}
{{- $clientKey := default $cert.Key .RootScope.Values.enforcer.admissionController.secret.key | b64enc -}}
caCert: {{ .CA.Cert | b64enc }}
clientCert: {{ $clientCert }}
clientKey: {{ $clientKey }}
{{- end -}}

{{/*
Create guardrails state-reporter name
*/}}
{{- define "guardrails.state-reporter.name" -}}
{{- printf "%s-reporter" (include "guardrails.name" .) -}}
{{- end -}}

{{/*
Create guardrails reporter fullname
*/}}
{{- define "guardrails.state-reporter.fullname" -}}
{{- printf "%s-state-reporter" (include "guardrails.fullname" .) -}}
{{- end -}}

{{/*
reporter labels
*/}}
{{- define "guardrails.state-reporter.labels" -}}
{{- template "guardrails.labels.tpl" (list . (include "guardrails.state-reporter.name" .)) -}}
{{- end -}}

{{/*
Create configmap name for guardrails.state-reporter.env vars.
*/}}
{{- define "guardrails.state-reporter.configmap.env.fullname" -}}
{{-  printf "%s-env" (include "guardrails.state-reporter.fullname" .) -}}
{{- end -}}

{{/*
Generate env vars for reporter.
Const env vars are taken from the values, dynamic env vars are generated here.
*/}}
{{- define "guardrails.state-reporter.env" -}}
{{ toYaml .Values.stateReporter.env }}
{{- end -}}