# How to cut a new release

## Process

To release a version _“v0.W.Z”_ of the `konfluxctl` in GitHub, follow these steps:

1. Create annotated and GPG-signed tag

```sh
git tag -s v0.W.Z -m "v0.W.Z"
git push origin v0.W.Z
```

2. In Github, [create release](https://github.com/eguzki/konfluxctl/releases/new).

* Pick recently pushed git tag
* Automatically generate release notes from previous released tag
* Set as the latest release

3. Verify that the build [Release workflow](https://github.com/eguzki/konfluxctl/actions/workflows/release.yaml) is triggered and completes for the new tag

### Verify new release is available

Download `konfluxctl` binary from [releases](https://github.com/eguzki/konfluxctl/releases) page.
The binary is available in multiple `OS` and `arch`. Pick your option.

```sh
wget https://github.com/eguzki/konfluxctl/releases/download/v0.W.Z/konfluxctl-v0.W.Z-{OS}-{arch}.tar.gz

tar -zxf konfluxctl-v0.W.Z-{OS}-{arch}.tar.gz
```

2. Verify version, it should be:

```sh
./konfluxctl version
```

The output should be the expected v0.W.Z and commitID. For example

```
konfluxctl v0.3.0 (eec318b2e11e7ea5add5e550ff872bde64555d8f)
```
