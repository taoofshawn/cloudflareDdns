# DRAFT

# cloudflareDdns
Keep cloudflare dns records up to date with your origin dynamic IP.  

For example:  

Home Web Server > Router/Cable Modem (dynamic IP) > Cloudflare > World  


## Notes
- This only contacts cloudflare on start and if it detects a change in the dynamic IP.

- It will not detect a change done on the cloudflare side.

- Will need to be restarted if it is missing a change on the cf side.

- Can't decide if I want to keep it like this or contact cf on every loop




