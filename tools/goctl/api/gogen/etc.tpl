# tools/goctl/api/gogen/etc.tpl
Name: {{.serviceName}}
Host: {{.host}}
Port: {{.port}}

Log:
  Path: "/data/log/api/{{.serviceName}}/logs"
  Level: "info"
  Encoding: plain
  Mode: "file"
  ServiceName: "{{.serviceName}}-api"
  TimeFormat: "2006-01-02 15:04:05"
  KeepDays: 15

Prometheus:
  Host: 0.0.0.0
  Port: 8103
  Path: /metrics