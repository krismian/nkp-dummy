{{- define "nkp-dummy.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nkp-dummy.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "nkp-dummy.labels" -}}
{{ include "nkp-dummy.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.common.image.tag | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.environmentname }}
site: {{ . }}
{{- end }}
{{- end -}}

{{- define "nkp-dummy.configMapName" -}}
{{- printf "configmap-%s-%s" .Release.Name .Values.common.environment -}}
{{- end -}}

{{- define "nkp-dummy.credentialName" -}}
{{- printf "secret-%s-%s" .Release.Name .Values.common.environment -}}
{{- end -}}
