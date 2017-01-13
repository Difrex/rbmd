# RBMD

RBD mount wrapper cluster

**NOT FOR PRODUCTION**

Current status: *development*, *testing*

**NOT FOR PRODUCTION**

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [RBMD](#rbmd)
    - [Dependencies](#dependencies)
    - [Usage](#usage)
        - [Example](#example)
    - [Build](#build)
    - [API](#api)
        - [GET /status](#get-status)
            - [Example](#example)
        - [GET /node](#get-node)
            - [Example](#example)
        - [GET /health](#get-health)
            - [Example](#example)
        - [POST /mount](#post-mount)
            - [Example](#example)
        - [POST /umount](#post-umount)
            - [Example](#example)
        - [POST /resolve](#post-resolve)
            - [Example](#example)
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
    	HTTP API listen address (default "127.0.0.1:9076")
  -tick int
    	Tick time loop (default 5)
  -version
    	Show version info and exit
  -ws string
    	Websockets listen address (default "127.0.0.1:7690")
  -zk string
    	Zookeeper servers comma separated (default "127.0.0.1:2181")
  -zkPath string
    	Zookeeper path (default "/rbmd")
```

### Example

```./rbmd -listen 127.0.0.1:9908 -zkPath /rbmd-ru-dc3-rack5```

## Build

Required Go > 1.6

```
git clone https://github.com/rbmd/rbmd.git && cd rbmd
GOPATH=$(pwd) go get github.com/gorilla/websocket
GOPATH=$(pwd) go get github.com/samuel/go-zookeeper/zk
GOPATH=$(pwd) go build
```

## API

### GET /status

Return JSON of quorum status

#### Example

```json
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
```json
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
  * resizing. Match regexp: ```^resizing\.$``` -- One or more nodes goind down
  * deadly. Match regexp: ```^deadly\.$``` -- One or more nodes is down and they has mapped images
  

#### Example
```
curl 127.0.0.1:9076/health
alive.
```

### POST /mount

Map rbd image and mount it

Allowed mount options:
 * ro
 * noatime
 * relatime
 * nosuid
 * noexec
 * nodiratime

#### Example

Accept JSON
```json
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
```json
{
    "state": "OK",
    "message": "OK"
}
```

On failure
```json
{
    "state": "FAIL",
    "message": "mount: /dev/null not a block device"
}
```

### POST /umount

Unmount filesystem and unmap RBD device

#### Example
 
Accept JSON
```json
{
    "node": "node.example.com",
    "mountpoint": "/var/www/pictures",
    "block": "rbd0"
}
```

Return JSON.

On success
```json
{
    "state": "OK",
    "message": "OK"
}
```

On failure
```json
{
    "state": "FAIL"
    "message": "Not found"
}
```

### POST /resolve

Remove deadly node from quorum.

#### Example

Accept JSON
```json
{
    "node": "node.example.com"
}
```

# AUTHORS

Denis Zheleztsov <difrex.punk@gmail.com>

# LICENSE

GPL version 3 see [LICENSE](LICENSE)
