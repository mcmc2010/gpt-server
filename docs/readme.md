```
add_header Access-Control-Allow-Origin   $http_origin; # '*';
add_header Access-Control-Allow-Methods 'PUT,POST,GET,DELETE,OPTIONS';
add_header Access-Control-Allow-Headers 'accept,authorization,content-type,content-encoding,cache-control,transfer-encoding,openai-organization';
add_header Access-Control-Expose-Headers 'authorization,content-type,content-encoding,cache-control,transfer-encoding';
add_header Access-Control-Allow-Credentials 'true';

if ($request_method = 'OPTIONS') {
   return 200;
}

```

```
# 开启gzip
gzip on;

# 启用gzip压缩的最小文件，小于设置值的文件将不会压缩
gzip_min_length 1k;

# gzip 压缩级别，1-9，数字越大压缩的越好，也越占用CPU时间，后面会有详细说明
gzip_comp_level 1;

# 进行压缩的文件类型。JavaScript有多种形式。其中的值可以在 mime.types 文件中找到。
gzip_types text/plain application/json application/xml 
```


```
    location ~ ^/api/(.*)$ {
		add_header           "Access-Control-Allow-Headers" "accept,authorization,content-type,content-encoding,cache-control;transfer-encoding,openai-organization";
		#add_header          "Access-Control-Allow-Credentials" "true";
		add_header           "Access-Control-Expose-Headers" "authorization,content-type,content-encoding,cache-control,transfer-encoding";
		#
		resolver             114.114.114.114 8.8.8.8;
		#
		#proxy_http_version   1.1; 
		proxy_pass           https://api.openai.com/$1$is_args$args;
		#proxy_pass          https://www.google.com/$1;
		proxy_redirect       off;
		#proxy_set_header    Host $host;
		#proxy_set_header    X-Real-IP $remote_addr;
		#proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header     user-agent $http_user_agent;
		proxy_cookie_domain  google.com <domain>;
		proxy_set_header     Cookie $http_cookie;
		proxy_connect_timeout        20s;
		proxy_read_timeout   20s;
		proxy_send_timeout   20s;
		proxy_ssl_session_reuse off;
		proxy_ssl_server_name        on;
		#proxy_ssl_name              $proxy_host;
		#Linux remove SSLv2 SSLv3
		proxy_ssl_protocols         TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
		#Windows Add SSLv2 SSLv3
		#proxy_ssl_protocols         TLSv1 TLSv1.1 TLSv1.2 TLSv1.3 SSLv2 SSLv3;
    }
```

```
# Fedora or Other Linux Nginx Fixed
#
## Error : (13: Permission denied) while connecting to upstream
#
sudo cat /var/log/audit/audit.log | grep nginx | grep denied

sudo setsebool httpd_can_network_connect on -P
sudo setsebool httpd_can_network_relay on -P

getsebool -a | grep httpd
```


