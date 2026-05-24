# Implementation

## Layering

実装は小さなレイヤーに分けます。

```text
outer
  CLI / local filesystem / query engine / presentation
    -> application
      -> domain
inner
```

依存方向は外側から内側へ向けます。内側の domain は外側の実装詳細を知りません。
Application は ports を定義し、外側の adapters がそれを実装します。

## Directory Structure

Go 実装では、Clean Architecture の境界を `internal/` 配下に置きます。
初回リリースでは過度に細かく分けず、外側の adapter と内側の domain/application が見分けられる粒度に留めます。

```text
.
├── cmd/
│   └── shelvia/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── book.go
│   │   ├── allowed_values.go
│   │   ├── author.go
│   │   ├── edition.go
│   │   ├── file_path.go
│   │   ├── genre.go
│   │   ├── imprint.go
│   │   ├── optional_text.go
│   │   ├── publisher.go
│   │   ├── rating.go
│   │   ├── read_date.go
│   │   └── title.go
│   ├── application/
│   │   ├── commands.go
│   │   ├── ports.go
│   │   ├── shelf_service.go
│   │   └── query_service.go
│   └── adapter/
│       ├── cli/
│       │   ├── parser.go
│       │   └── runner.go
│       ├── localfs/
│       │   ├── shelf_repository.go
│       │   └── templates.go
│       ├── queryengine/
│       │   └── query_store.go
│       └── presentation/
│           └── output.go
├── docs/
│   ├── architecture.md
│   ├── shelf.md
│   ├── implementation.md
│   └── testing.md
├── fixtures/
│   └── shelves/
│       ├── valid/
│       │   └── minimal/
│       │       ├── config.toml
│       │       ├── example.toml
│       │       └── Some Book.toml
│       └── invalid/
│           ├── nested-config/
│           ├── missing-required/
│           ├── unknown-config-value/
│           └── unsafe-filename/
├── README-ja.md
├── README.md
└── AGENTS.md
```

各ディレクトリの責務は次のとおりです。

- `cmd/shelvia`: 実行可能ファイルの入口。依存の組み立てだけを行い、業務ロジックを持たない
- `internal/domain`: `Book` aggregate/entity、設定由来の許可値、値オブジェクト、純粋な検証ルール
- `internal/application`: use case と port 定義。domain と port interface にだけ依存する
- `internal/adapter/cli`: コマンドライン引数を application command に変換する
- `internal/adapter/localfs`: OS のファイルシステム上にある shelf の discovery、`config.toml` 読み込み、テンプレート作成
- `internal/adapter/queryengine`: 検証済み書籍から一時的な検索ビューを作り、query port を実装する
- `internal/adapter/presentation`: 成功・失敗・検索結果の表示整形
- `docs`: エージェントと実装者向けの設計文書
- `fixtures/shelves`: tests、README、手動確認で読み込ませる shelf fixtures

依存方向は次の形に固定します。

```text
cmd/shelvia
  -> internal/adapter/*
    -> internal/application
      -> internal/domain
```

禁止する依存は次のとおりです。

- `internal/domain` から `internal/application` や `internal/adapter` へ依存しない
- `internal/application` から `internal/adapter` へ依存しない
- `internal/adapter/queryengine` の内部型を domain や application の公開型に混ぜない
- CLI parser の都合を domain の型や validation に持ち込まない

小さいうちは package を増やしすぎません。新しい adapter や service は、重複を減らすか依存境界を守る必要が出た時点で追加します。

## Fixture Shelves

`fixtures/` は、Shelvia が実際に読み込む shelf の fixture を置くために使います。mock implementation とは用途を分け、`fixtures/shelves/` 配下に shelf root 単位で配置します。

`fixtures/shelves/valid/` には成功する shelf を置きます。`minimal/` は README と同じ基本構成を保ち、`config.toml`、`example.toml`、追加の書籍 TOML を含めます。

`fixtures/shelves/invalid/` には失敗を確認する shelf を置きます。各ディレクトリは 1 つの失敗理由だけを表し、テスト名から期待する診断が分かる名前にします。

Fixture は production code から特別扱いしません。tests や手動確認では、通常の `shelf root` と同じように `fixtures/shelves/...` の各ディレクトリを読み込ませます。

## Layer Responsibilities

### CLI

CLI レイヤーは、引数の解析、`shelf root` の解決、application へのコマンド委譲だけを担当します。書籍の意味的な検証や、TOML・SQLite の構造には踏み込みません。

### Application

Application レイヤーは、コマンドの流れを調停します。

- `config.toml` を読み込む
- 書籍ファイルを読み込む
- domain ルールを検証する
- 一時的な検索ビューを作る
- コマンド出力を返す

`init` が既存ファイルを上書きしない、といったコマンド単位の判断は application が持ちます。将来 `--force` のようなオプションを追加する場合も、この層で扱います。

Application は local filesystem や query engine の具象実装には依存しません。必要な操作は port として定義し、外側の adapter が実装します。

### Domain

Domain レイヤーは、書籍と設定ファイル由来の許可値に関するルールを所有します。

`Book` は書籍記録の aggregate root/entity として扱います。`Book` 自体は value object ではありません。`Title`、`Author`、`Rating`、`ReadDate`、`Genre`、`Publisher`、`Edition`、`Imprint`、`OptionalText`、`FilePath` などを value object として分け、各ファイルに不変条件を閉じ込めます。

- 必須項目が存在する
- `rating` は `0` から `100` の整数である
- `read_date` は TOML の日付である
- `genre`、`publisher`、`edition`、`imprint` は設定ファイルの許可値と照合できる
- `imprint` は `edition` なしでは指定できない
- `edition.imprint_required = true` の場合は `imprint` が必須になる
- 任意テキスト項目の空文字は未指定として扱う

Domain は、ファイルシステム、SQLite、端末表示、コマンドライン引数に依存しません。

### Local Filesystem Adapter

Local filesystem adapter は、OS のファイルシステム上にある shelf の読み書きを担当します。`.toml` 探索と `config.toml` の扱いは [Shelf and Commands](./shelf.md) に従います。

- `init` のために `config.toml` と `example.toml` を作る
- `new` のために 1 冊分の TOML テンプレートを作る
- `new` のファイル名を OS 非依存に検証し、既存ファイルを上書きしない

ファイル由来の診断では、`Some Book.toml` のようなユーザーに見えるパスを保持します。

### Query Engine Adapter

Query engine adapter は、検証済みの書籍からインメモリの検索ビューを作ります。初回実装では内部実装として SQLite を使ってよいですが、package 名には永続化方式やライブラリ名を出しません。SQLite は検索を助けるための一時ビューであり、永続化は行いません。

`query --where` は、最終的な SQL を組み立てる前に、[Shelf and Commands](./shelf.md) で定義した範囲に収まるかを検証します。初回リリースでは保守的にし、未対応の SQL fragment は adapter 境界で拒否します。

### Presentation

Presentation レイヤーは、ユーザー向けの表示を担当します。`validate` の成功・失敗は、短く読みやすくします。

成功例:

```text
Validated 1 book, 1 config file.
```

失敗例:

```text
Some Book.toml:5: unknown genre "Novel"
```

## Directory Structure Self Review

`docs/implementation.md` のディレクトリ構成は、初回実装に入るための最小構成としては妥当です。ただし、次の点は意識して運用します。

- `internal/domain` は package を細かく分けず、Go の 1 package として保つ。ファイルは DDD の概念単位に分ける
- `Book` は value object ではなく aggregate root/entity として扱う。値の同一性だけで扱える型だけを value object にする
- `allowed_values.go` は設定ファイルの raw schema ではなく、domain が使う許可値集合を表す。TOML の形は adapter 側に閉じ込める
- `internal/adapter/localfs` は OS ファイルシステム依存を表す名前として使う。shelf という業務語だけではなく、外部境界が local filesystem であることを明示する
- `internal/adapter/queryengine` は application の query port を実装する adapter として命名する。内部で SQLite を使っても、application や domain に SQLite 名を漏らさない
- `internal/adapter/presentation` は出力整形が増えた時点で `text` や `table` などに分ける余地があるが、初回リリースでは 1 package でよい
- `fixtures/` は fixture shelf の置き場として使う。mock implementation が必要になった場合は、`fixtures/` に混ぜず、各 package の test helper か専用名に分ける
