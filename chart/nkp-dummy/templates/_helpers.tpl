{{- define "nkp-dummy.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nkp-dummy.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "nkp-dummy.labels" -}}
{{ include "nkp-dummy.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "nkp-dummy.secretName" -}}
{{- default (include "nkp-dummy.fullname" .) .Values.secret.existingSecret -}}
{{- end -}}
