# buildx bake

```text
docker buildx bake [OPTIONS] [TARGET...]
```

<!---MARKER_GEN_START-->
Build from a file

### Aliases

`docker buildx bake`, `docker buildx f`

### Options

| Name                             | Type          | Default | Description                                                                              |
|:---------------------------------|:--------------|:--------|:-----------------------------------------------------------------------------------------|
| [`--builder`](#builder)          | `string`      |         | Override the configured builder instance                                                 |
| [`-f`](#file), [`--file`](#file) | `stringArray` |         | Build definition file                                                                    |
| `--load`                         |               |         | Shorthand for `--set=*.output=type=docker`                                               |
| `--metadata-file`                | `string`      |         | Write build result metadata to the file                                                  |
| [`--no-cache`](#no-cache)        |               |         | Do not use cache when building the image                                                 |
| [`--print`](#print)              |               |         | Print the options without building                                                       |
| [`--progress`](#progress)        | `string`      | `auto`  | Set type of progress output (`auto`, `plain`, `tty`). Use plain to show container output |
| [`--provenance`](#provenance)    | `string`      |         | Shorthand for `--set=*.attest=type=provenance`                                           |
| [`--pull`](#pull)                |               |         | Always attempt to pull all referenced images                                             |
| `--push`                         |               |         | Shorthand for `--set=*.output=type=registry`                                             |
| [`--sbom`](#sbom)                | `string`      |         | Shorthand for `--set=*.attest=type=sbom`                                                 |
| [`--set`](#set)                  | `stringArray` |         | Override target value (e.g., `targetpattern.key=value`)                                  |


<!---MARKER_GEN_END-->

## Description

Bake is a high-level build command. Each specified target runs in parallel
as part of the build.

Read [High-level build options with Bake](https://docs.docker.com/build/bake/)
guide for introduction to writing bake files.

> **Note**
>
> `buildx bake` command may receive backwards incompatible features in the future
> if needed. We are looking for feedback on improving the command and extending
> the functionality further.

## Examples

### <a name="builder"></a> Override the configured builder instance (--builder)

Same as [`buildx --builder`](buildx.md#builder).

### <a name="file"></a> Specify a build definition file (-f, --file)

Use the `-f` / `--file` option to specify the build definition file to use.
The file can be an HCL, JSON or Compose file. If multiple files are specified,
all are read and the build configurations are combined.

You can pass the names of the targets to build, to build only specific target(s).
The following example builds the `db` and `webapp-release` targets that are
defined in the `docker-bake.dev.hcl` file:

```hcl
# docker-bake.dev.hcl
group "default" {
  targets = ["db", "webapp-dev"]
}

target "webapp-dev" {
  dockerfile = "Dockerfile.webapp"
  tags = ["docker.io/username/webapp"]
}

target "webapp-release" {
  inherits = ["webapp-dev"]
  platforms = ["linux/amd64", "linux/arm64"]
}

target "db" {
  dockerfile = "Dockerfile.db"
  tags = ["docker.io/username/db"]
}
```

```console
$ docker buildx bake -f docker-bake.dev.hcl db webapp-release
```

See the [Bake file reference](https://docs.docker.com/build/bake/reference/)
for more details.

### <a name="no-cache"></a> Don't use cache when building the image (--no-cache)

Same as `build --no-cache`. Don't use cache when building the image.

### <a name="print"></a> Print the options without building (--print)

Prints the resulting options of the targets desired to be built, in a JSON
format, without starting a build.

```console
$ docker buildx bake -f docker-bake.hcl --print db
{
  "group": {
    "default": {
      "targets": [
        "db"
      ]
    }
  },
  "target": {
    "db": {
      "context": "./",
      "dockerfile": "Dockerfile",
      "tags": [
        "docker.io/tiborvass/db"
      ]
    }
  }
}
```

### <a name="progress"></a> Set type of progress output (--progress)

Same as [`build --progress`](buildx_build.md#progress).

### <a name="provenance"></a> Create provenance attestations (--provenance)

Same as [`build --provenance`](buildx_build.md#provenance).

### <a name="pull"></a> Always attempt to pull a newer version of the image (--pull)

Same as `build --pull`.

### <a name="sbom"></a> Create SBOM attestations (--sbom)

Same as [`build --sbom`](buildx_build.md#sbom).

### <a name="set"></a> Override target configurations from command line (--set)

```
--set targetpattern.key[.subkey]=value
```

Override target configurations from command line. The pattern matching syntax
is defined in https://golang.org/pkg/path/#Match.

```console
$ docker buildx bake --set target.args.mybuildarg=value
$ docker buildx bake --set target.platform=linux/arm64
$ docker buildx bake --set foo*.args.mybuildarg=value # overrides build arg for all targets starting with 'foo'
$ docker buildx bake --set *.platform=linux/arm64     # overrides platform for all targets
$ docker buildx bake --set foo*.no-cache              # bypass caching only for targets starting with 'foo'
```

You can override the following fields:

* `args`
* `cache-from`
* `cache-to`
* `context`
* `dockerfile`
* `labels`
* `no-cache`
* `no-cache-filter`
* `output`
* `platform`
* `pull`
* `push`
* `secrets`
* `ssh`
* `tags`
* `target`
