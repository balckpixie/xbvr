# Windows ビルドガイド

Windows 環境で `xbvr.exe` をビルドするための環境要件、スクリプト構成、およびビルド手順をまとめたドキュメントです。

## 1. ビルド環境（前提ツール）

以下のツールが必要です（かっこ内は動作確認済みのバージョン）。

- **Go** (`go 1.24.0` 以上 / 確認済み: `go1.27.4 windows/amd64`)
- **MinGW-w64 (GCC / G++)**（確認済み: `WinLibs` / `BrechtSanders.WinLibs.POSIX.UCRT` `16.2.0`、`x86_64-w64-mingw32-gcc.exe`）
  - `CGO_ENABLED=1` でのビルド（`go-sqlite3` やトレイアイコン機能等）に必要です。
- **Node.js**（確認済み: `v24.19.0`）
- **Yarn**（確認済み: `1.22.22`）

### ローカル依存パッケージの配置

`package.json` および `yarn.lock` はローカルパッケージ `videojs-thumbnail-sprite-custom` を相対パス `../libs/videojs-thumbnail-sprite-custom/` で参照しています。  
リポジトリが `D:\Dev\Repos\xbvr` にある場合、以下の位置に配置されている必要があります。

```text
D:\Dev\Repos\
├── libs\
│   └── videojs-thumbnail-sprite-custom\
└── xbvr\
```

## 2. ビルドスクリプトの構成

`ui/fs.go` が `//go:embed dist` でフロントエンドのビルド成果物（`ui/dist`）を埋め込む構成になっているため、**フロントエンド → Go バイナリ** の順でビルドする必要があります。

| ファイル | 役割 |
| :--- | :--- |
| `build_all.bat` | 一括ビルド用バッチ（推奨）。ツールパスの補完、依存関係インストール、`vbuild.bat`・`build.bat` の順次実行を行います。 |
| `vbuild.bat` | フロントエンド（Vue UI）のビルドを実行し、`ui/dist` を生成します（`call yarn run build`）。 |
| `build.bat` | `xbvr.exe` をビルドします（`GOOS=windows`, `GOARCH=amd64`, `CGO_ENABLED=1`, `-ldflags "-H windowsgui"`, `-tags=json1`, 対象: `pkg/tray/main.go`）。`PATH` 上に `x86_64-w64-mingw32-gcc` があればそれを使用し、無い場合は `gcc` / `g++` にフォールバックします。 |

## 3. ビルド手順

### 一括ビルド（推奨）

リポジトリのルート（`d:\Dev\Repos\xbvr`）で以下を実行します。

```bat
build_all.bat
```

`build_all.bat` は内部で以下を順番に実行します。

1. Go・WinLibs（MinGW-w64）・Yarn のパスを現在のセッションの `PATH` に補完
2. `node_modules` が存在しない場合のみ `yarn install --ignore-engines` を実行
3. `vbuild.bat` でフロントエンドをビルド（`ui/dist` を生成）
4. `build.bat` で `xbvr.exe` を生成

### 個別ビルド

ツールチェインに `PATH` が通っている状態で、用途に応じて個別に実行することもできます。

- **フロントエンドのみ再ビルドする場合**:
  ```bat
  vbuild.bat
  ```
- **Go 側のみ再ビルドする場合**（`ui/dist` が既に存在する場合）:
  ```bat
  build.bat
  ```

## 4. トラブルシューティング

### `yarn install` で Node バージョンエラーが出る場合

Node.js v24 などの新しい環境では、依存パッケージ `@achrinza/node-ipc@9.2.9` のエンジン制約により通常の `yarn install` が失敗します。  
手動でインストールする場合は `--ignore-engines` を付けて実行してください（`build_all.bat` には組み込み済みです）。

```bat
yarn install --ignore-engines
```

### `cgo: C compiler "gcc" not found` が出る場合

Go の CGO が C コンパイラを見つけられていません。`build_all.bat` 経由で実行するか、MinGW-w64 の `bin` ディレクトリ（例: `%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin`）を `PATH` に通してから `build.bat` を実行してください。
