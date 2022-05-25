Name: {{.serviceName}}.rpc
ListenOn: 127.0.0.1:8080
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: {{.serviceName}}.rpc

Log:
  Path: "/data/log/rpc/{{.serviceName}}/logs"
  Level: "info"
  Encoding: plain
  Mode: "file"
  ServiceName: "{{.serviceName}}-rpc"
  TimeFormat: "2006-01-02 15:04:05"
  KeepDays: 15