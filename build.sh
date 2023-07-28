# !/bin/sh

set -e

# prepare folders
rm -rf nginxlog-exporter && mkdir nginxlog-exporter
cp config.yml nginxlog-exporter && cp -r logs nginxlog-exporter/logs

echo "--> building nginxlog-exporter_linux_amd64 <--"

GOOS=linux GOARCH=amd64 go build -a -o nginxlog-exporter/nginxlog-exporter main.go
tar -czvf "nginxlog-exporter_linux_amd64.tar.gz" nginxlog-exporter