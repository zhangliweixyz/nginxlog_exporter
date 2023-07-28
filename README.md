# nginxlog_exporter
A exporter to parse Nginx access log for prometheus metrics.

## Features
- Various log formats support.
- Request path can be rewrited.
- Status code can be mergeable.

## Installation
1. go get `github.com/zhangliweixyz/nginxlog_exporter`
2. Or use [binary](https://github.com/zhangliweixyz/nginxlog_exporter/releases) release

## Usage
```
nginxlog_exporter -h 

Usage of:
  -config.file string
    	Nginx log exporter configuration file name. (default "config.yml")
  -web.listen-address string
    	Address to listen on for the web interface and API. (default ":9999")
exit status 2
```

## Configuration
```
- name: nginx
  format: "$time_iso\t$http_x_forwarded_for\t$remote_addr\t$request\t$server_port\t$status\t$body_bytes_sent\t$http_referer\t\"$http_user_agent\"\t$request_time\t$upstream_addr\t$upstream_status\t$upstream_response_time\t$http_host\t$http_cookie"
  source_files:
    - ./logs/access.log
  static_labels:
    region: qa
  relabel_config:
    source_labels:
      - http_host
      - request
      - status
      - request_time
      - upstream_addr
      - upstream_response_time
    replacement:
      request:
        trim: "?"
      status:
        replaces:
          - target: 4.+
            value: 4xx
          - target: 5.+
            value: 5xx
  histogram_buckets: [0.1, 0.3, 0.5, 1, 2]
```

- format: your nginx `log_format` regular expression.
- name: service name, metric will be `{name}_http_request_count_total`.
- source_files: service nginx log, support multiple files.
- static_labels: all metrics will add this labelsets.
- relabel_config:
    * source_labels: what's labels should be use.
    * replacement: source_labels value format rule, it supports regrex.
- histogram_buckets: configure histogram metrics buckets.

## Thanks
- Inspired by [nginx-log-exporter](https://github.com/songjiayang/nginx-log-exporter)

## 补充说明
本项目主要是参考开源项目nginx-log-exporter，按照自己实际需求做了简化和修改。