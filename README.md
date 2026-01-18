# IKuai Prometheus Exporter

## 部署

- 只提供 Metrics
```
python manager.py --restart --ikuai_ip {ikuai_ip} --username {username} --password {password}
curl localhost:12695/metrics
```

- 提供 Metrics 并自动 Push 到 Pushgateway
```
python manager.py --restart --ikuai_ip {ikuai_ip} --username {username} --password {password} --pushgateway_url {pushgateway_url} --pushgateway_username {pushgateway_username} --pushgateway_password {pushgateway_password}
curl localhost:12695/metrics
```
