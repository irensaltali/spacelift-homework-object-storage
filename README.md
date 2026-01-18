# spacelift-homework-object-storage


## Docker commands

```bash
docker ps --filter name=amazin-object-storage-node --format {{.ID}}

0eac241250cf
86039758f40a
61fcf510aae8

docker inspect 0eac241250cf

[
    {
        "Id": "0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee",
        "Created": "2026-01-17T13:57:41.843449549Z",
        "Path": "/usr/bin/docker-entrypoint.sh",
        "Args": [
            "server",
            "--console-address",
            ":9001",
            "/tmp/data"
        ],
        "State": {
            "Status": "running",
            "Running": true,
            "Paused": false,
            "Restarting": false,
            "OOMKilled": false,
            "Dead": false,
            "Pid": 659,
            "ExitCode": 0,
            "Error": "",
            "StartedAt": "2026-01-18T09:28:51.585832877Z",
            "FinishedAt": "2026-01-17T16:43:24.725022303Z"
        },
        "Image": "sha256:14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e",
        "ResolvConfPath": "/var/lib/docker/containers/0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee/resolv.conf",
        "HostnamePath": "/var/lib/docker/containers/0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee/hostname",
        "HostsPath": "/var/lib/docker/containers/0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee/hosts",
        "LogPath": "/var/lib/docker/containers/0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee/0eac241250cfe0e919ede5fa7b9f70fb3c919043659a5b3e1ef1f2384bc21fee-json.log",
        "Name": "/homework-object-storage-amazin-object-storage-node-3-1",
        "RestartCount": 0,
        "Driver": "overlayfs",
        "Platform": "linux",
        "MountLabel": "",
        "ProcessLabel": "",
        "AppArmorProfile": "",
        "ExecIDs": null,
        "HostConfig": {
            "Binds": null,
            "ContainerIDFile": "",
            "LogConfig": {
                "Type": "json-file",
                "Config": {}
            },
            "NetworkMode": "homework-object-storage_amazin-object-storage",
            "PortBindings": {
                "9001/tcp": [
                    {
                        "HostIp": "",
                        "HostPort": "9003"
                    }
                ]
            },
            "RestartPolicy": {
                "Name": "no",
                "MaximumRetryCount": 0
            },
            "AutoRemove": false,
            "VolumeDriver": "",
            "VolumesFrom": null,
            "ConsoleSize": [
                0,
                0
            ],
            "CapAdd": null,
            "CapDrop": null,
            "CgroupnsMode": "private",
            "Dns": [],
            "DnsOptions": [],
            "DnsSearch": [],
            "ExtraHosts": [],
            "GroupAdd": null,
            "IpcMode": "private",
            "Cgroup": "",
            "Links": null,
            "OomScoreAdj": 0,
            "PidMode": "",
            "Privileged": false,
            "PublishAllPorts": false,
            "ReadonlyRootfs": false,
            "SecurityOpt": null,
            "UTSMode": "",
            "UsernsMode": "",
            "ShmSize": 67108864,
            "Runtime": "runc",
            "Isolation": "",
            "CpuShares": 0,
            "Memory": 0,
            "NanoCpus": 0,
            "CgroupParent": "",
            "BlkioWeight": 0,
            "BlkioWeightDevice": null,
            "BlkioDeviceReadBps": null,
            "BlkioDeviceWriteBps": null,
            "BlkioDeviceReadIOps": null,
            "BlkioDeviceWriteIOps": null,
            "CpuPeriod": 0,
            "CpuQuota": 0,
            "CpuRealtimePeriod": 0,
            "CpuRealtimeRuntime": 0,
            "CpusetCpus": "",
            "CpusetMems": "",
            "Devices": null,
            "DeviceCgroupRules": null,
            "DeviceRequests": null,
            "MemoryReservation": 0,
            "MemorySwap": 0,
            "MemorySwappiness": null,
            "OomKillDisable": null,
            "PidsLimit": null,
            "Ulimits": null,
            "CpuCount": 0,
            "CpuPercent": 0,
            "IOMaximumIOps": 0,
            "IOMaximumBandwidth": 0,
            "MaskedPaths": [
                "/proc/acpi",
                "/proc/asound",
                "/proc/interrupts",
                "/proc/kcore",
                "/proc/keys",
                "/proc/latency_stats",
                "/proc/sched_debug",
                "/proc/scsi",
                "/proc/timer_list",
                "/proc/timer_stats",
                "/sys/devices/virtual/powercap",
                "/sys/firmware"
            ],
            "ReadonlyPaths": [
                "/proc/bus",
                "/proc/fs",
                "/proc/irq",
                "/proc/sys",
                "/proc/sysrq-trigger"
            ]
        },
        "GraphDriver": {
            "Data": null,
            "Name": ""
        },
        "Mounts": [
            {
                "Type": "volume",
                "Name": "3167c7494ae0f3c94f0c90a6d72218f6a6cefc7da006a7ae2271bb85f9e7b635",
                "Source": "/var/lib/docker/volumes/3167c7494ae0f3c94f0c90a6d72218f6a6cefc7da006a7ae2271bb85f9e7b635/_data",
                "Destination": "/data",
                "Driver": "local",
                "Mode": "",
                "RW": true,
                "Propagation": ""
            }
        ],
        "Config": {
            "Hostname": "0eac241250cf",
            "Domainname": "",
            "User": "",
            "AttachStdin": false,
            "AttachStdout": true,
            "AttachStderr": true,
            "ExposedPorts": {
                "9000/tcp": {},
                "9001/tcp": {}
            },
            "Tty": false,
            "OpenStdin": false,
            "StdinOnce": false,
            "Env": [
                "MINIO_ACCESS_KEY=rendezvous",
                "MINIO_SECRET_KEY=bluegreen",
                "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
                "MINIO_ACCESS_KEY_FILE=access_key",
                "MINIO_SECRET_KEY_FILE=secret_key",
                "MINIO_ROOT_USER_FILE=access_key",
                "MINIO_ROOT_PASSWORD_FILE=secret_key",
                "MINIO_KMS_SECRET_KEY_FILE=kms_master_key",
                "MINIO_UPDATE_MINISIGN_PUBKEY=RWTx5Zr1tiHQLwG9keckT0c45M3AGeHD6IvimQHpyRywVWGbP1aVSGav",
                "MINIO_CONFIG_ENV_FILE=config.env",
                "MC_CONFIG_DIR=/tmp/.mc"
            ],
            "Cmd": [
                "server",
                "--console-address",
                ":9001",
                "/tmp/data"
            ],
            "Image": "minio/minio",
            "Volumes": {
                "/data": {}
            },
            "WorkingDir": "/",
            "Entrypoint": [
                "/usr/bin/docker-entrypoint.sh"
            ],
            "OnBuild": null,
            "Labels": {
                "architecture": "aarch64",
                "build-date": "2025-08-06T08:17:03",
                "com.docker.compose.config-hash": "6c38ff1b17d711c3bccbefd24673298c2c2d78caa7d1c07a421d578b43a57257",
                "com.docker.compose.container-number": "1",
                "com.docker.compose.depends_on": "",
                "com.docker.compose.image": "sha256:14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e",
                "com.docker.compose.oneoff": "False",
                "com.docker.compose.project": "homework-object-storage",
                "com.docker.compose.project.config_files": "/Users/irensaltali/source/irensaltali/homework-object-storage/docker-compose.yml",
                "com.docker.compose.project.working_dir": "/Users/irensaltali/source/irensaltali/homework-object-storage",
                "com.docker.compose.service": "amazin-object-storage-node-3",
                "com.docker.compose.version": "2.40.3",
                "com.redhat.component": "ubi9-micro-container",
                "com.redhat.license_terms": "https://www.redhat.com/en/about/red-hat-end-user-license-agreements#UBI",
                "description": "MinIO object storage is fundamentally different. Designed for performance and the S3 API, it is 100% open-source. MinIO is ideal for large, private cloud environments with stringent security requirements and delivers mission-critical availability across a diverse range of workloads.",
                "distribution-scope": "public",
                "io.buildah.version": "1.40.1",
                "io.k8s.description": "Very small image which doesn't install the package manager.",
                "io.k8s.display-name": "Red Hat Universal Base Image 9 Micro",
                "io.openshift.expose-services": "",
                "maintainer": "MinIO Inc \u003cdev@min.io\u003e",
                "name": "MinIO",
                "release": "RELEASE.2025-09-07T16-13-09Z",
                "summary": "MinIO is a High Performance Object Storage, API compatible with Amazon S3 cloud storage service.",
                "url": "https://catalog.redhat.com/en/search?searchType=containers",
                "vcs-ref": "32d8e5209f029a1fb7235308ab32253bface1001",
                "vcs-type": "git",
                "vendor": "MinIO Inc \u003cdev@min.io\u003e",
                "version": "RELEASE.2025-09-07T16-13-09Z"
            },
            "StopTimeout": 1
        },
        "NetworkSettings": {
            "Bridge": "",
            "SandboxID": "4ce57990b6eca6f08ee66d1c64b164c2347f690573ef6c13adc1d6e9639dddd8",
            "SandboxKey": "/var/run/docker/netns/4ce57990b6ec",
            "Ports": {
                "9001/tcp": [
                    {
                        "HostIp": "0.0.0.0",
                        "HostPort": "9003"
                    },
                    {
                        "HostIp": "::",
                        "HostPort": "9003"
                    }
                ]
            },
            "HairpinMode": false,
            "LinkLocalIPv6Address": "",
            "LinkLocalIPv6PrefixLen": 0,
            "SecondaryIPAddresses": null,
            "SecondaryIPv6Addresses": null,
            "EndpointID": "",
            "Gateway": "",
            "GlobalIPv6Address": "",
            "GlobalIPv6PrefixLen": 0,
            "IPAddress": "",
            "IPPrefixLen": 0,
            "IPv6Gateway": "",
            "MacAddress": "",
            "Networks": {
                "homework-object-storage_amazin-object-storage": {
                    "IPAMConfig": {
                        "IPv4Address": "169.253.0.4"
                    },
                    "Links": null,
                    "Aliases": [
                        "homework-object-storage-amazin-object-storage-node-3-1",
                        "amazin-object-storage-node-3"
                    ],
                    "MacAddress": "b2:20:6d:c8:7c:c8",
                    "DriverOpts": null,
                    "GwPriority": 0,
                    "NetworkID": "0715396c22293a74c06eaafa69e24842b8ab5dbadb7b3e551ce95e2534e1993f",
                    "EndpointID": "fe95b1055399bb5d654ea5ae0b45b86706ed71ae6fa04e687f6547426ee7e1e4",
                    "Gateway": "169.253.0.1",
                    "IPAddress": "169.253.0.4",
                    "IPPrefixLen": 24,
                    "IPv6Gateway": "",
                    "GlobalIPv6Address": "",
                    "GlobalIPv6PrefixLen": 0,
                    "DNSNames": [
                        "homework-object-storage-amazin-object-storage-node-3-1",
                        "amazin-object-storage-node-3",
                        "0eac241250cf"
                    ]
                }
            }
        },
        "ImageManifestDescriptor": {
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "digest": "sha256:9966a92a734f9411e32f4f41d7d9d826fcdc0f68c4e20b70295bd4e7c11f8a2f",
            "size": 2081,
            "platform": {
                "architecture": "arm64",
                "os": "linux"
            }
        }
    }
]
```