---
sidebar_position: 200
---

# Install

github-comment is written in Go. So you only have to install a binary in your `PATH`.

There are some ways to install github-comment.

1. [Homebrew](#homebrew)
1. [Scoop](#scoop)
1. [aqua](#aqua)
1. [mise](#mise)
1. [GitHub Releases](#github-releases)
1. [Build an executable binary from source code yourself using Go](#build-an-executable-binary-from-source-code-yourself-using-go)

## Homebrew

You can install github-comment using [Homebrew](https://brew.sh/).

```sh
brew install suzuki-shunsuke/github-comment/github-comment
```

## Scoop

You can install github-comment using [Scoop](https://scoop.sh/).

```sh
scoop bucket add suzuki-shunsuke https://github.com/suzuki-shunsuke/scoop-bucket
scoop install github-comment
```

## aqua

You can install github-comment using [aqua](https://aquaproj.github.io/).

```sh
aqua g -i suzuki-shunsuke/github-comment
```

## mise

You can install github-comment v6.4.2 or later using [mise](https://mise.jdx.dev/) with the [packslip](https://packslip.dev/) backend.

```sh
mise use packslip:github.com/suzuki-shunsuke/github-comment
```

## Build an executable binary from source code yourself using Go

```sh
go install github.com/suzuki-shunsuke/github-comment/v6/cmd/github-comment@latest
```

## GitHub Releases

You can download an asset from [GitHub Releases](https://github.com/suzuki-shunsuke/github-comment/releases).
Please unarchive it and install a pre built binary into `$PATH`. 

### Verify downloaded assets from GitHub Releases

You can verify downloaded assets using some tools.

1. [GitHub CLI](https://cli.github.com/)
1. [Cosign](https://github.com/sigstore/cosign)
1. [Packslip](https://packslip.dev/) `github-comment >= v6.4.2`
1. [slsa-verifier](https://github.com/slsa-framework/slsa-verifier) `github-comment <= v6.4.1`

### 1. GitHub CLI

You can install GitHub CLI by aqua.

```sh
aqua g -i cli/cli
```

```sh
version=v6.4.2
asset=github-comment_${version#v}_darwin_arm64.tar.gz
gh release download -R suzuki-shunsuke/github-comment "$version" -p "$asset"
gh attestation verify "$asset" \
  -R suzuki-shunsuke/github-comment \
  --signer-workflow suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

### 2. Cosign

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

```sh
version=v6.4.2
checksum_file="github-comment_${version#v}_checksums.txt"
asset=github-comment_${version#v}_darwin_arm64.tar.gz
gh release download "$version" \
  -R suzuki-shunsuke/github-comment \
  -p "$asset" \
  -p "$checksum_file" \
  -p "${checksum_file}.sigstore.json"
cosign verify-blob \
  --bundle "${checksum_file}.sigstore.json" \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$checksum_file"
cat "$checksum_file" | sha256sum --ignore-missing -c -
```

For v6.4.1 or older, the signature and certificate are published as `${checksum_file}.sig` and `${checksum_file}.pem` instead of `${checksum_file}.sigstore.json`.
Download them and pass `--signature "${checksum_file}.sig" --certificate "${checksum_file}.pem"` to `cosign verify-blob` instead of `--bundle`.

### 3. Packslip

`github-comment >= v6.4.2`

You can install packslip by aqua.

```sh
aqua g -i jdx/packslip
```

```sh
version=v6.4.2
asset=github-comment_${version#v}_darwin_arm64.tar.gz
gh release download -R suzuki-shunsuke/github-comment "$version" -p "$asset" -p packslip.sigstore.json
packslip verify packslip.sigstore.json \
    --pin ps1_4vy6kb5zu2owi6ez56hmilp2iy \
    --identity-prefix https://github.com/suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml@ \
    --artifact "$asset"
```

- `--pin` requires the release to be signed from the repository `suzuki-shunsuke/github-comment`, identified by its repository ID rather than its name
- `--identity-prefix` requires the release to be signed by the release workflow of [suzuki-shunsuke/go-release-workflow](https://github.com/suzuki-shunsuke/go-release-workflow)

On success, packslip prints a line starting with `ok:`.
Please make sure it shows the project and version you meant, e.g. `github.com/suzuki-shunsuke/github-comment 6.4.2`.

### 4. slsa-verifier

:::note
This method works only for v6.4.1 or older.
v6.4.2 or later don't publish `multiple.intoto.jsonl`, so please use GitHub CLI, Cosign, or Packslip instead.
:::

You can install slsa-verifier by aqua.

```sh
aqua g -i slsa-framework/slsa-verifier
```

```sh
version=v6.4.1
asset=github-comment_${version#v}_darwin_arm64.tar.gz
gh release download -R suzuki-shunsuke/github-comment "$version" -p "$asset" -p multiple.intoto.jsonl
slsa-verifier verify-artifact "$asset" \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/suzuki-shunsuke/github-comment \
  --source-tag "$version"
```
