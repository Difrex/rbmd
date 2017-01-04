RBD mount wrapper cluster

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [-](#-)
- [Usage](#usage)
    - [Example](#example)
    - [API](#api)
        - [GET /status](#get-status)
            - [Example](#example)
        - [GET /node](#get-node)
            - [Example](#example)
        - [GET /health](#get-health)
            - [Example](#example)
        - [POST /mount](#post-mount)
        - [POST /umount](#post-umount)
- [AUTHORS](#authors)
- [LICENSE](#license)

<!-- markdown-toc end -->


## Dependencies

* Zookeeper
* Access to ceph cluster(for map/unmap images)

## Usage

```
Usage of ./rbmd:
  -listen string
    	HTTP API listen address (default "0.0.0.0:9076")
  -tick int
    	Tick time loop (default 5)
  -ws string
    	Websockets listen address (default "0.0.0.0:7690")
  -zk string
    	Zookeeper servers comma separated (default "127.0.0.1:2181")
  -zkPath string
    	Zookeeper path (default "/rbmd")
```

### Example

```./rbmd -listen 127.0.0.1:9908 -zkPath /rbmd-ru-dc3-rack5```

## API

### GET /status

Return JSON of quorum status

#### Example

```
 curl 127.0.0.1:9076/status | jq
{
  "quorum": {
    "node.example.com": {
      "node": "node.example.com",
      "ip": {
        "v4": [
          "10.0.3.1"
        ],
        "v6": [
          "fe80::f869:d0ff:fea3:3c0a"
        ]
      },
      "updated": 1483428452,
      "mounts": null
    }
  },
  "health": "alive.",
  "leader": "node.example.com"
}
```

### GET /node

Return JSON of node stats 

#### Example
```
{
  "node": "difrex-mac.wargaming.net",
  "ip": {
    "v4": [
      "169.254.156.1"
    ],
    "v6": [
      "fe80::108d:fcff:fe77:3df6"
    ]
  },
  "updated": 1483095493,
  "mounts": null
}
```

### GET /health

Return string with quorum health check result

Statuses:
  * alive. Match regexp: ```^alive\.$``` -- all is good
  * resizing. Match regexp: ```^resizing\. (.+)``` -- One or more nodes goind down
  * deadly. Match regexp: ```^deadly\. (.+)``` -- One or more nodes is down and they has mapped images. Return string with \n
  

#### Example
```
curl 127.0.0.1:9076/health
alive.
```

### POST /mount

Map rbd image and mount it

#### Example

Accept JSON
```
{
    "node": "node.example.com",
    "pool": "web",
    "image": "pictures",
    "mountpoint": "/var/www/pictures",
    "mountopts": "noatime,nodiratime",
    "fstype": "xfs"
}
```

Return JSON.

On success
```
{
    "state": "OK",
    "message": "OK"
}
```

On failure
```
{
    "state": "FAIL"
    "message": "mount: /dev/null not a block device"
}
```

### POST /umount

Accept JSON. Not implemented yet.

# AUTHORS

Denis Zheleztsov <difrex.punk@gmail.com>

# LICENSE

GPL version 3 see LICENSE
