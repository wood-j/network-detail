# 1. Monitor Network Detail

A simple command line tool to monitor network workload of linux server.

And save history data to `sqlite3` file `network_detail.db`.

## 1.1. env

Develop enviroment:

- golang: 1.20.6

mod:

```bash
go mod tidy
```

## 1.2. compile

Use following command to compile binary exec.

```bash
go build ./network_detail.go
```

## 1.3. run

Run cli with:

- replace with your network device name, example `ens160`

```bash
./network_detail -i ens160
```

## 1.4. systemd

For long term running, we suggest to use systemd service

Create or edit file `/usr/lib/systemd/system/network-detail.service`:

- replace with your upload path, example `/data/deploy/network_detail`

```ini
[Unit]
Description=Monitor network detail
After=network.target

[Service]
ExecStart=/data/deploy/network_detail/network_detail 
WorkingDirectory=/data/deploy/network_detail/
Restart=always

[Install]
WantedBy=multi-user.target
```

Apply changes and restart/status service:

```bash
systemctl daemon-reload
systemctl restart network-detail
systemctl enable network-detail
systemctl status network-detail
```

## 1.5. Query

Just simple sql query to analysis.

### 1.5.1. Examples

Query detail:

```sql
SELECT * FROM  network_detail ORDER BY time DESC;
```

Query by minute:

```sql
SELECT strftime('%Y-%m-%d %H:%M', time) AS minute, sum(send_bytes) as sum_send_bytes, sum(receive_bytes) as sum_receive_bytes
FROM network_detail
GROUP BY minute
ORDER BY minute DESC;
```

Filter with condition:

```sql
SELECT * FROM (
  SELECT strftime('%Y-%m-%d %H:%M', time) AS minute, sum(send_bytes) as sum_send_bytes, sum(receive_bytes) as sum_receive_bytes
  FROM network_detail
  GROUP BY minute
  ORDER BY minute DESC
)
WHERE sum_send_bytes > 1000000 or sum_receive_bytes > 1000000;
```
