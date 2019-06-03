# Versions
Atom 2.4.0
MySQL 5 (Google Cloud SQL)

# Running commands on the instance
## General

 - ssh in to instance (gcloud compute ssh daa-server)
 - get root: sudo su root
 - cd /usr/share/nginx/atom

# Regenerate slugs

 - php symfony propel:generate-slugs --delete
 - php symfony cc & php symfony search:populate

## Redirects

Use broken-links to build a new redirects map and copy to /etc/nginx/redirects.map

Include this block:
`
map $uri $new_uri {
    include /etc/nginx/redirects.map;
}

server {
  listen 80 default_server;
  listen [::]:80 default_server;
  server_name directory.archivists.org.au;

  if ($new_uri) {
     return 301 https://$host$new_uri;
  }
  return 301 https://$host$request_uri;
}
`

And remove listen 80 from the first line of the atom server block


## HTTPS

	sudo apt-get update
	sudo apt-get install software-properties-common
	sudo add-apt-repository universe
	sudo add-apt-repository ppa:certbot/certbot
	sudo apt-get update
	sudo apt-get install certbot python-certbot-nginx 

    sudo certbot --nginx

    Test autorenewal cron job: sudo certbot renew --dry-run

Add: 

server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name directory.archivists.org.au;
    return 301 https://$host$request_uri;
}

And remove listen 80 from main server block.

Cetbox includes 


