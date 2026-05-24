# Architecture

## 目的

このドキュメントは、Shelvia の実装に関するアーキテクチャ判断を記録するための設計書です。人間とエージェントの両方が、リポジトリ内だけで設計意図を追えることを目的にします。

Shelvia は、個人の読書記録を扱う小さな CLI ツールです。管理対象の情報は `shelf root` 配下の TOML ファイルであり、SQLite は検証済みデータから起動時に作る一時的な検索ビューとしてのみ扱います。

OpenAI のハーネスエンジニアリングの考え方に従い、`AGENTS.md` は短い目次として保ち、実装に必要な永続的な知識は `docs/` 配下の Markdown に置きます。エージェントがコード、設計、制約、検証方法をリポジトリ内で発見できる状態を重視します。

参考:

- [OpenAI: ハーネスエンジニアリング](https://openai.com/ja-JP/index/harness-engineering/)

## 設計目標

- 読書データを、Shelvia の外からも編集・差分管理できるプレーンテキストとして保つ
- 検証を十分に厳しくし、検索結果を信頼できる状態にする
- 初期実装は小さく、読み返しやすい構成にする
- 場当たり的な分岐より、明示的な境界と不変条件を優先する
- エラーでは、ユーザーが直すべきファイルと理由を示す

## 非目標

- Shelvia はアプリケーション DB を永続的な情報源として所有しない
- Shelvia は汎用的なテキスト検索ツールを置き換えない
- 初回リリースでは、プラグイン機構やサービス構成を持たない

## リポジトリ知識の置き場所

`README.md` と `README-ja.md` は、ユーザー向けの説明を置く場所です。動機、インストール、使い方、データ形式、例を中心にし、実装計画やアーキテクチャ詳細は置きません。

`docs/architecture.md` は、実装者とエージェント向けのアーキテクチャ設計を置く場所です。コマンドの責務、データフロー、検証境界、不変条件が変わる場合は、このファイルも同じ変更で更新します。

`AGENTS.md` は、詳細な手順書ではなく目次として扱います。深い情報を重複して書かず、信頼できる Markdown への入口を示します。

## アーキテクチャ方針

Shelvia は Clean Architecture に則ります。依存方向は常に内側へ向け、domain は application、CLI、filesystem、SQLite、presentation に依存しません。

Onion Architecture と Hexagonal Architecture はどちらも候補になりますが、この CLI では Clean Architecture を基本方針にし、外部依存との境界表現として Hexagonal Architecture の ports/adapters を採用します。

理由は次のとおりです。

- CLI、filesystem、SQLite、presentation という外部境界が明確で、ports/adapters と相性がよい
- Domain を中心に置く点は Onion Architecture と同じだが、CLI ツールでは「どの外部入出力を adapter として差し替えるか」を明示した方が実装しやすい
- Clean Architecture の依存ルールを上位方針にすれば、Onion と Hexagonal のよい部分を過不足なく使える

## Shelf Model

Shelvia は 1 つの `shelf root` を受け取り、その配下の `.toml` ファイルを再帰的に読み込みます。`shelf root` 直下の `config.toml` だけは例外で、書籍ファイルではなく設定ファイルとして扱います。

`config.toml` には、ジャンル、出版社、判型、レーベル名などの許可値を列挙します。

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

`shelf root` は次の順で決まります。

1. `--shelf` に指定したディレクトリ
2. 環境変数 `SHELVIA_DIR`

どちらも指定されていない場合はエラー終了し、`--shelf` を渡すか `SHELVIA_DIR` を設定するように案内します。

## Commands

初回リリースでは、次のコマンドを対象にします。

- `shelvia init PATH`
- `shelvia new TITLE --read-date YYYY-MM-DD`
- `shelvia validate`
- `shelvia list`
- `shelvia query --where SQL_FRAGMENT`

`init` は、指定された `shelf root` に `config.toml` と `example.toml` を作成します。

`new` は、`<title>.toml` に書籍ファイルのテンプレートを作成します。

`validate` は、`shelf root` 全体を読み込み、`list` と `query` でも使える状態かを確認します。

`list` は、検証済みの書籍を予測可能な既定順で表示します。

`query --where` は、初回リリースでは SQL 風の条件を受け取ります。内部的には独自 DSL ではなく、検証済みデータから作った一時 SQLite ビューに対する制限付き SQL fragment として扱います。

## Layering

実装は小さなレイヤーに分けます。

```text
outer
  CLI / filesystem / sqlite / presentation
    -> application
      -> domain
inner
```

依存方向は外側から内側へ向けます。内側の domain は外側の実装詳細を知りません。Application は ports を定義し、外側の adapters がそれを実装します。

## Directory Structure

Go 実装では、Clean Architecture の境界を `internal/` 配下に置きます。初回リリースでは過度に細かく分けず、外側の adapter と内側の domain/application が見分けられる粒度に留めます。

```text
.
├── cmd/
│   └── shelvia/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── book.go
│   │   ├── config.go
│   │   ├── validation.go
│   │   └── value_objects.go
│   ├── application/
│   │   ├── commands.go
│   │   ├── ports.go
│   │   ├── shelf_service.go
│   │   └── query_service.go
│   └── adapter/
│       ├── cli/
│       │   ├── parser.go
│       │   └── runner.go
│       ├── filesystem/
│       │   ├── shelf_repository.go
│       │   └── templates.go
│       ├── sqlite/
│       │   └── query_store.go
│       └── presentation/
│           └── output.go
├── docs/
│   └── architecture.md
├── mocks/
├── README-ja.md
├── README.md
└── AGENTS.md
```

各ディレクトリの責務は次のとおりです。

- `cmd/shelvia`: 実行可能ファイルの入口。依存の組み立てだけを行い、業務ロジックを持たない
- `internal/domain`: 書籍、設定値、値オブジェクト、純粋な検証ルール
- `internal/application`: use case と port 定義。domain と port interface にだけ依存する
- `internal/adapter/cli`: コマンドライン引数を application command に変換する
- `internal/adapter/filesystem`: `shelf root` の `.toml` 探索、`config.toml` 読み込み、テンプレート作成
- `internal/adapter/sqlite`: 検証済み書籍から一時 SQLite ビューを作り、query port を実装する
- `internal/adapter/presentation`: 成功・失敗・検索結果の表示整形
- `docs`: エージェントと実装者向けの設計文書
- `mocks`: README や手動確認で使うサンプル shelf

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
- `internal/adapter/sqlite` の型を domain や application の公開型に混ぜない
- CLI parser の都合を domain の型や validation に持ち込まない

小さいうちは package を増やしすぎません。新しい adapter や service は、重複を減らすか依存境界を守る必要が出た時点で追加します。

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

Application は filesystem や SQLite の具象実装には依存しません。必要な操作は port として定義し、外側の adapter が実装します。

### Domain

Domain レイヤーは、書籍と設定ファイル由来の許可値に関するルールを所有します。

- 必須項目が存在する
- `rating` は `0` から `100` の整数である
- `read_date` は TOML の日付である
- `genre`、`publisher`、`edition`、`imprint` は設定ファイルの許可値と照合できる
- `imprint` は `edition` なしでは指定できない
- `edition.imprint_required = true` の場合は `imprint` が必須になる
- 任意テキスト項目の空文字は未指定として扱う

Domain は、ファイルシステム、SQLite、端末表示、コマンドライン引数に依存しません。

### Filesystem Adapter

Filesystem adapter は、shelf の読み書きを担当します。

- `shelf root` 配下の `.toml` ファイルを再帰的に探索する
- `shelf root` 直下の `config.toml` を設定ファイルとして読み込む
- `config.toml` 以外の `.toml` ファイルを書籍ファイルとして読み込む
- `init` のために `config.toml` と `example.toml` を作る
- `new` のために 1 冊分の TOML テンプレートを作る

ファイル由来の診断では、`Some Book.toml` のようなユーザーに見えるパスを保持します。

### SQLite Adapter

SQLite adapter は、検証済みの書籍からインメモリの検索ビューを作ります。SQLite は検索を助けるための一時ビューであり、永続化は行いません。

`query --where` は、最終的な SQL を組み立てる前に、明らかに危険または未対応の SQL fragment を拒否します。初回リリースでは保守的にし、単純な比較、真偽演算、明示的に対応した並び替えだけを許可します。

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

## Data Flow

`validate`、`list`、`query` は同じ読み込み経路を使います。

1. `shelf root` を解決する
2. `config.toml` を設定ファイルとして読み込む
3. `config.toml` 以外の `.toml` ファイルを書籍ファイルとして読み込む
4. TOML の raw data を domain value に変換する
5. domain ルールを検証する
6. 診断または検証済み書籍を返す
7. `list` と `query` では、検証済み書籍からインメモリ SQLite ビューを作る
8. 結果を表示する

未検証のデータを `list` や `query` に渡してはいけません。

## Invariants

実装が大きくなっても、次のルールは保ちます。

- 永続化は TOML ファイルのみで行い、DB は永続化しない
- SQLite は TOML から再構築され、SQLite 側の変更を TOML へ書き戻さない
- domain validation は SQLite に投入する前に完了する
- ファイル由来のエラーは、ユーザーが直すべきファイルパスを含む
- 依存方向は Clean Architecture の内側へ向ける
- `README.md` と `README-ja.md` は、ユーザー向けの挙動について同期する
- `AGENTS.md` は目次として保ち、詳細な設計は `docs/` に置く

## Test Strategy

一般的なテスト戦略を取ります。Domain は単体テストで細かく確認し、それ以外は外部依存を port の裏で差し替えた結合テストで確認します。

### Domain Unit Tests

Domain の単体テストでは、外部 I/O を使わずに純粋なルールを確認します。

- 必須項目
- `rating`
- `read_date`
- 設定ファイル由来の許可値との照合
- 任意項目の空文字が未指定として扱われること
- edition と imprint の関係

### Integration Tests

Application、CLI、presentation、filesystem 境界は結合テストで確認します。インフラ層の依存先、特に SQLite などの DB は mock または fake adapter に差し替えます。

- `shelf root` の解決と、未指定時のエラー
- `init` のファイル作成方針
- `new` の作成パスと TOML テンプレート
- `config.toml` が書籍ファイルとして扱われないこと
- TOML parsing から application command までの流れ
- ファイルパスを含む検証エラー
- `validate`、`list`、`query` の command output
- 対応する `--where` fragment が query port に渡ること
- 未対応または危険な SQL fragment が adapter 境界で拒否されること

DB そのものの挙動は SQLite adapter の薄い adapter test に限定します。
主な振る舞いは、DB に依存しない application-level integration test で検証します。
