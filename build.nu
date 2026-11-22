# build.nu
# 使い方: nu build.nu

def main [] {
    if not ("go.mod" | path exists) {
        error make {
            msg: "go.mod が見つかりません。プロジェクトのルートで実行してください。"
        }
    }

    mkdir bin
    ls bin | each {|item| rm -rf $item.name }

    let dev_bin = if $nu.os-info.name == "windows" { "bin/sketch.exe" } else { "bin/sketch" }

    with-env {CGO_ENABLED: "0"} {
        go build -trimpath -ldflags="-s -w" -o $dev_bin
    }

    print $"Built: ($dev_bin)"
}
