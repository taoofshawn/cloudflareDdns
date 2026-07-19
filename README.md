# [cloudflareDdns](https://github.com/taoofshawn/cloudflareDdns)
Keep cloudflare dns records up to date with your origin dynamic IP.

My use case:

Home Web Server > Router/Cable Modem (dynamic IP) > Cloudflare > World

## Environment Variables

### APITOKEN (required)
- go here  https://dash.cloudflare.com/profile/api-tokens
- create token
- permission the zones you would like to manage

### NAMELIST (required)
Comma separated list of records that will be updated or created. An A/AAAA record will be created if it does not exist and the zone is permissioned.

### POLLTIME (optional, default: 60)
Minutes to wait between checking in with cloudflare.

### PROXIED (optional, default: true)
Whether created records should be proxied through Cloudflare. Set to `false` for DNS-only records.

### RECORD_TYPE (optional, default: A)
DNS record type to manage. `A` for IPv4, `AAAA` for IPv6.

### TZ (optional)
Container timezone (e.g. `America/New_York`).

## Running

### docker run
```
docker run -d \
    --name=cloudflareddns \
    -e APITOKEN=<api token from cloudflare> \
    -e NAMELIST="www.example.com,example.com" \
    -e TZ=America/New_York \
    -e POLLTIME=360 \
    -e PROXIED=false \
    -e RECORD_TYPE=A \
    taoofshawn/cloudflareddns
```

### docker-compose.yml
```yaml
version: '3'
services:
    cloudflareddns:
        build: .
        container_name: cloudflareddns
        image: taoofshawn/cloudflareddns
        environment:
            - APITOKEN=<token from cloudflare>
            - POLLTIME=360
            - TZ=America/New_York
            - NAMELIST=www.example.com,
                        example.com,
                        www.example2.com,
                        example2.com
```

Dictionary style with a folded scalar also works:
```yaml
version: '3'
services:
    cloudflareddns:
        build: .
        container_name: cloudflareddns
        image: taoofshawn/cloudflareddns
        environment:
            APITOKEN: <token from cloudflare>
            POLLTIME: 360
            TZ: America/New_York
            NAMELIST: >-
                www.example.com,
                example.com,
                www.example2.com,
                example2.com
```

## Logs

Use `docker logs` to monitor. The log format is structured key=value:

```
time=2026/07/19 12:44:43 level=INFO msg="Mornin' Ralph"
time=2026/07/19 12:44:43 level=INFO msg="verifying token"
time=2026/07/19 12:44:43 level=INFO msg="This API Token is valid and active"
time=2026/07/19 12:44:43 level=INFO msg="fetching zones"
time=2026/07/19 12:44:43 level=INFO msg="found zone" name=example.com
time=2026/07/19 12:44:43 level=INFO msg="refreshing zone records"
time=2026/07/19 12:44:43 level=INFO msg="found a record" name=www.example.com ip=1.2.3.4
time=2026/07/19 12:44:43 level=INFO msg="checking current dynamic (origin) IP address"
time=2026/07/19 12:44:43 level=INFO msg="current dynamic IP" ip=203.0.113.42
time=2026/07/19 12:44:43 level=INFO msg="no update needed" name=www.example.com
```
