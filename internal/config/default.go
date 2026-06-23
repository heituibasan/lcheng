package config

const DefaultYAML = `mixed-port: 7890
allow-lan: false
mode: rule
log-level: info
external-controller: 127.0.0.1:9090
secret: ""

dns:
  enable: true
  enhanced-mode: fake-ip
  nameserver:
    - 223.5.5.5
    - 119.29.29.29

proxies: []

proxy-groups:
  - name: PROXY
    type: select
    proxies:
      - DIRECT

rules:
  - MATCH,PROXY
`
