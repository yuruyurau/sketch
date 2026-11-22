# release.nu
# 使い方: nu release.nu

def main [] {
    if not ("go.mod" | path exists) {
        error make {
            msg: "go.mod が見つかりません。プロジェクトのルートで実行してください。"
        }
    }

    let version = (try {
        (git describe --tags --abbrev=0 | str trim)
    } catch {
        "0.0.0"
    } | str replace -r "^v" "")

    print $"Version: ($version)"

    mkdir release
    ls release | each {|item| rm -rf $item.name }

    let targets = [
        {os: "windows", arch: "amd64", bin: "sketch.exe"}
        {os: "linux", arch: "amd64", bin: "sketch"}
        {os: "linux", arch: "arm64", bin: "sketch"}
        {os: "darwin", arch: "amd64", bin: "sketch"}
        {os: "darwin", arch: "arm64", bin: "sketch"}
    ]

    for t in $targets {
        let folder_name = $"sketch_($version)_($t.os)_($t.arch)"
        let dir = $"release/($folder_name)"
        mkdir $dir

        with-env {GOOS: $t.os, GOARCH: $t.arch, CGO_ENABLED: "0"} {
            go build -trimpath -ldflags="-s -w" -o $"($dir)/($t.bin)"
        }

        cp LICENSE $"($dir)/LICENSE"
        cp config.json $"($dir)/config.json"

        do {
            cd release
            ^7z a -tzip $"($folder_name).zip" $folder_name | ignore
            rm -rf $folder_name
        }

        print $"Created: ($dir).zip"
    }
}
