# SNITCH
## VERY BETA

Retrieves certain server info - Lists docker containers and NPM hosts. Made for Linux but works for Windows too.

Wrote this because I wanted to be able to snapshot the current status of the server rather than running a bunch of commands.

```
####### SNITCH v0.01 #######

Hostname: my-linux-server
OS: ubuntu 22.04 (6.5.0-28-generic)
Architecture: amd64
CPU: Intel(R) Xeon(R) CPU E5-2676 v3 @ 2.40GHz (4 cores)
Memory: 16.00 GB total, 5.24 GB used (32.75%)
Uptime: 72h15m32.123456s
IP Addresses:
 - Interface: lo, IP: 127.0.0.1
 - Interface: lo, IP: ::1
 - Interface: eth0, IP: 192.168.10.100
 - Interface: eth0, IP: fe80::a00:27ff:fe4e:66a1
Docker Installed: Yes
  Docker Version: 24.0.7
  Docker Containers:
    nginx-proxy-manager | jc21/nginx-proxy-manager:latest | Up 4 days
    redis-server | redis:7.2 | Exited (0) 2 days ago
    mysql-db | mysql:8.0 | Up 4 days
UFW Firewall Status: Active
NPM Records: 
  Searching for database.sqlite starting at: /opt
  Found database: /opt/nginx_proxy/data/database.sqlite
  Reading table: proxy_host
    - ["example.com"] -> 192.168.1.10:80 [DELETED]
    - ["api.example.com"] -> 10.0.0.5:8080
    - ["test-site.org"]-> 172.16.5.100:443 [DELETED]
  Reading table: redirection_host
    - ["redirect.com"] -> www.redirect.com
    - ["old-site.org"] -> older-site.org [DELETED]
```  

Download the latest release or try something like for Linux.

curl -L https://github.com/xavier-hernandez/snitch/releases/download/v.0.01/snitch -o /tmp/snitch && chmod +x /tmp/snitch && /tmp/snitch