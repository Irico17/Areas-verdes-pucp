#!/bin/sh
set -eu
mkdir -p /etc/nginx/snippets
cp /opt/campus/nginx-headers.conf /etc/nginx/snippets/campus-headers.conf
if [ -f /etc/nginx/certs/fullchain.pem ] && [ -f /etc/nginx/certs/privkey.pem ]; then
  cp /opt/campus/nginx.tls.conf /etc/nginx/conf.d/default.conf
else
  cp /opt/campus/nginx.http.conf /etc/nginx/conf.d/default.conf
fi
exec nginx -g 'daemon off;'
