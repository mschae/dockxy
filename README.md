# Dockxy

Dockxy is a nginx config file generator for docker files.

My biggest pain when I started with docker was working with a SOA.
Docker assigns random ports to exposed ports of containers. Therefore
talking to those endpoints required me to look up the ports using `docker ps -a`,
feeding them into other Apps, etc.

With Dockxy you can listen to docker events (e.g. creation or deletion of new
containers) and automatically create nginx config files to reverse-proxy from
nginx to those containers.

The port and the IP as well as the name of the container will automatically be
filled in.

## Running

You can download the pre-built binary. Or you can build it yourself by running
`go build`.

The following runtime flags are available:

Flag         | Meaning | Default
-------------|-------- | -------
dockerIP     | The IP address of the docker host                | 192.168.59.103 (boot2docker default)
dockerURL    | The address to docker (e.g. tcp://10.0.0.1:1234) | tcp://192.168.59.103:2375
templatePath | Path to the nginx template                       | templates/site.tmpl
outDir       | Directory for the generated config files         | out

They are all optional. If you're using `boot2docker` there is a good chance all
you have to do is `sudo ./dockxy`

Finally, add the following to your `nginx.conf` (ideally near the last closing
bracket):

```
include /path/to/your/outDir/*;
```

## Caveats

* The outDir will be cleaned on each run. This means all files inside will be deleted.
* Since the program reloads nginx, you will most likely have to run it as root.

## Example

Given dockxy is running.

Running

```
docker up --name foo -P some/container
docker up --name bar -P some/other_container
```

Will automatically generate the following config files:

foo.conf


```
upstream foo.dev {
  server 192.168.59.103:49158;
}

server {
  listen 80;
  server_name foo.dev *.foo.dev;

  client_max_body_size 50M;
  error_page 500 502 503 504 /50x.html;

  location = /50x.html {
    root html;
  }

  try_files $uri/index.html $uri @foo.dev;
  location @foo.dev {
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header Host $http_host;
    proxy_redirect off;
    proxy_pass http://foo.dev;
    add_header Access-Control-Allow-Origin *;
  }
}
```

bar.conf

```
upstream bar.dev {
  server 192.168.59.103:49159;
}

server {
  listen 80;
  server_name bar.dev *.bar.dev;

  client_max_body_size 50M;
  error_page 500 502 503 504 /50x.html;

  location = /50x.html {
    root html;
  }

  try_files $uri/index.html $uri @bar.dev;
  location @bar.dev {
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header Host $http_host;
    proxy_redirect off;
    proxy_pass http://bar.dev;
    add_header Access-Control-Allow-Origin *;
  }
}
```

So we can just do `curl foo.bar` and it'll direct us inside the container.

Cool huh?
