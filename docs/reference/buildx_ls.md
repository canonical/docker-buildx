# buildx ls

```text
docker buildx ls
```

<!---MARKER_GEN_START-->
List builder instances


<!---MARKER_GEN_END-->

## Description

Lists all builder instances and the nodes for each instance

```console
$ docker buildx ls
NAME/NODE       DRIVER/ENDPOINT             STATUS  BUILDKIT PLATFORMS
elated_tesla *  docker-container
  elated_tesla0 unix:///var/run/docker.sock running v0.10.3  linux/amd64
  elated_tesla1 ssh://ubuntu@1.2.3.4        running v0.10.3  linux/arm64*, linux/arm/v7, linux/arm/v6
default         docker
  default       default                     running v0.8.2   linux/amd64
```

Each builder has one or more nodes associated with it. The current builder's
name is marked with a `*` in `NAME/NODE` and explicit node to build against for
the target platform marked with a `*` in the `PLATFORMS` column.
