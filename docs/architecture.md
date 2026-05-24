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

- はアプリケーション DB を永続的な情報源として所有しない
- Shelvia は汎用的なテキスト検索ツールを置き換えない
- 初回リリースでは、プラグイン機構やサービス構成を持たない

## リポジトリ知識の置き場所

`README.md` と `README-ja.md` は、ユーザー向けの説明を置く場所です。動機、インストール、使い方、データ形式、例を中心にし、実装計画やアーキテクチャ詳細は置きません。

`docs/architecture.md` は、実装者とエージェント向けのアーキテクチャ設計を置く場所です。コマンドの責務、データフロー、検証境界、不変条件が変わる場合は、このファイルも同じ変更で更新します。

`AGENTS.md` は、詳細な手順書ではなく目次として扱います。深い情報を重複して書かず、信頼できる Markdown への入口を示します。

## Shelf Model

Shelvia は 1 つの `shelf root` を受け取り、その配下の `books/` と `vocab/` を読み込みます。

```text
my-shelf/
  books/
    2024/
      Some Book.toml
  vocab/
    genres.toml
    publishers.toml
    editions.toml
    imprints.toml
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

`init` は、`books/`、`vocab/`、空の vocab ファイルを作成します。

`new` は、`--read-date` の年を使って `books/<year>/<title>.toml` に書籍ファイルのひな形を作成します。

`validate` は、`shelf root` 全体を読み込み、`list` と `query` でも使える状態かを確認します。

`list` は、検証済みの書籍を予測可能な既定順で表示します。

`query --where` は、初回リリースでは SQL 風の条件を受け取ります。内部的には独自 DSL ではなく、検証済みデータから作った一時 SQLite ビューに対する制限付き SQL fragment として扱います。

## Layering

実装は小さなレイヤーに分けます。

```text
CLI
  -> application
    -> domain
    -> ports
      -> filesystem
      -> sqlite
      -> presentation
```

### CLI

CLI レイヤーは、引数の解析、`shelf root` の解決、application へのコマンド委譲だけを担当します。書籍の意味的な検証や、TOML・SQLite の構造には踏み込みません。

### Application

Application レイヤーは、コマンドの流れを調停します。

- vocab を読み込む
- 書籍ファイルを読み込む
- domain ルールを検証する
- 一時的な検索ビューを作る
- コマンド出力を返す

`init` が既存ファイルを上書きしない、といったコマンド単位の判断は application が持ちます。将来 `--force` のようなオプションを追加する場合も、この層で扱います。

### Domain

Domain レイヤーは、書籍と vocab のルールを所有します。

- 必須項目が存在する
- `rating` は `0` から `100` の整数である
- `read_date` は TOML の日付である
- `genre`、`publisher`、`edition`、`imprint` は vocab と照合できる
- `imprint` は `edition` なしでは指定できない
- `edition.imprint_required = true` の場合は `imprint` が必須になる
- 任意テキスト項目の空文字は未指定として扱う

Domain は、ファイルシステム、SQLite、端末表示、コマンドライン引数に依存しません。

### Filesystem

Filesystem adapter は、shelf の読み書きを担当します。

- `books/**/*.toml` を再帰的に読み込む
- `vocab/` から必要な vocab ファイルを読み込む
- `init` のために初期ディレクトリと vocab ファイルを作る
- `new` のために 1 冊分の TOML ひな形を作る

ファイル由来の診断では、`books/2024/Some Book.toml` のようなユーザーに見えるパスを保持します。

### SQLite

SQLite adapter は、検証済みの書籍からインメモリの検索ビューを作ります。SQLite は検索を助けるための一時ビューであり、永続化は行いません。

`query --where` は、最終的な SQL を組み立てる前に、明らかに危険または未対応の SQL fragment を拒否します。初回リリースでは保守的にし、単純な比較、真偽演算、明示的に対応した並び替えだけを許可します。

### Presentation

Presentation レイヤーは、ユーザー向けの表示を担当します。`validate` の成功・失敗は、短く読みやすくします。

成功例:

```text
Validated 1 book, 4 vocab files.
```

失敗例:

```text
books/2024/Some Book.toml:5: unknown genre "Novel"
```

## Data Flow

`validate`、`list`、`query` は同じ読み込み経路を使います。

1. `shelf root` を解決する
2. vocab ファイルを読み込む
3. 書籍 TOML ファイルを読み込む
4. TOML の raw data を domain value に変換する
5. domain ルールを検証する
6. 診断または検証済み書籍を返す
7. `list` と `query` では、検証済み書籍からインメモリ SQLite ビューを作る
8. 結果を表示する

未検証のデータを `list` や `query` に渡してはいけません。

## Invariants

実装が大きくなっても、次のルールは保ちます。

- 永続化は TOML ファイルのみで行い、DB は永続化しないにである
- SQLite は TOML から再構築され、SQliteの更新は TOML には反映しない
- domain validation は SQLite に投入する前に完了する
- ファイル由来のエラーは、ユーザーが直すべきファイルパスを含む
- `README.md` と `README-ja.md` は、ユーザー向けの挙動について同期する
- `AGENTS.md` は目次として保ち、詳細な設計は `docs/` に置く

## Test Strategy

初期のテストでは、次を重点的に確認します。

- `shelf root` の解決と、未指定時のエラー
- `init` のファイル作成
- `new` の作成パスと TOML ひな形
- TOML parsing と必須項目検証
- vocab 検証
- 任意項目の空文字が未指定として扱われること
- edition と imprint の関係
- ファイルパスを含む検証エラー
- 対応する `--where` fragment の SQLite query
- 未対応または危険な SQL fragment の拒否

テストは、守るレイヤーの近くに置きます。CLI が実装されたら、主要なユーザーフローは end-to-end のコマンドテストでも確認します。
