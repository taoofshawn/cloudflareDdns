# cloudflareDdns
Keep cloudflare dns records up to date with your origin dynamic IP.  

My use case:  

Home Web Server > Router/Cable Modem (dynamic IP) > Cloudflare > World  

## Usage  
### APITOKEN  
- go here  https://dash.cloudflare.com/profile/api-tokens
- create token
- permission the zones you would like to manage

### NAMELIST 
comma separated list of records that will be updated or created. A record will be created if it does not exist and the zone is permissioned.

### POLLTIME 
optional minutes to wait between checking in with cloudflare. defaults to 60 (one hour)
```
docker run -d \
    --name=cloudflareddns \
    -e APITOKEN=<api token from cloudflare> \
    -e NAMELIST="www.example.com,example.com" \
    -e TZ=America/New_York \
    -e POLLTIME=360 \
    taoofshawn/cloudflareddns
```
Or use docker-compose.yml:
```yaml
version: '3'
services:
    cloudflareddns:
        build: .
        container_name: cloudflareddns
        image: taoofshawn/cloudflareddns:testing
        environment:
            - APITOKEN=<token from cloudflare>
            - POLLTIME=360
            - TZ=America/New_York
            - NAMELIST=>
                www.example.com,
                example.com,
                www.example2.com,
                example2.com
```
use docker logs to monitor:
```
I0809 20:56:42.887544 3911846 main.go:20] Mornin' Ralph
I0809 20:56:42.887586 3911846 cloudflareClient.go:35] verifying token
I0809 20:56:42.887595 3911846 cloudflareClient.go:55] connecting to Cloudflare: https://api.cloudflare.com/client/v4/user/tokens/verify
I0809 20:56:43.211159 3911846 cloudflareClient.go:43] This API Token is valid and active
I0809 20:56:43.211182 3911846 cloudflareClient.go:76] getting zone list
I0809 20:56:43.211194 3911846 cloudflareClient.go:55] connecting to Cloudflare: https://api.cloudflare.com/client/v4/zones
I0809 20:56:43.388844 3911846 cloudflareClient.go:87] getting zone records
```
