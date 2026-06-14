# Shelvia

Shelvia は、読書記録を自分の手元で管理するための CLI ツールです。1 冊ごとの記録を TOML ファイルとして保存し、コマンドラインから検証、一覧表示、検索を行えます。

読書データは、アプリ専用のデータベースではなく、エディタでそのまま読めるプレーンテキストとして残ります。Git で履歴を管理でき、`grep` や `find` でも探せます。そのうえで、評価、読了日、ジャンル、出版社、レーベルなどの項目を使って、あとから検索が可能です。

## Why Shelvia?

読書記録は、表計算ソフトや読書管理サービスでも管理できます。どちらも便利ですが、長く続けるほど「データを自分の手元に置きたい」「サービスや UI に依存したくない」「あとから自由に検索したい」という場面が出てきます。

一方で、Markdown のような自由形式のメモだけで管理すると、人間には読みやすくても、次のような検索が難しくなります。

- 評価が 90 点以上の本を探す
- 2024 年に読んだ技術書だけを見る
- 特定の出版社やレーベルの本を一覧する

Shelvia は、プレーンテキストの扱いやすさと、データベース的な検索のしやすさを両立するためのツールです。

## 特徴

- 1 冊を 1 つの TOML ファイルとして保存
- 読書データを自分のリポジトリで管理
- `init` で `config.toml` と `example.toml` を作成
- `add` で書籍ファイルのテンプレートを作成
- 必須項目、評価、日付、ジャンル、出版社などの候補値を検証
- SQL 風の条件で検索
- TOML を情報源として扱い、SQLite は検索用の一時ビューとして利用

## インストール

インストール手順は初回リリース前に追加します。

## クイックスタート

読書データを置くディレクトリを作成します。Shelvia では、このディレクトリを `shelf root` と呼びます。

```bash
shelvia init ./my-shelf
```

`init` は指定したディレクトリを `shelf root` として初期化し、その中に `config.toml` と `example.toml` を作成します。指定したディレクトリが存在しない場合は、新しく作成します。
`example.toml` は、検証対象に含まれる書籍データのサンプルファイルです。
どちらのファイルも最小構成で始まり、コメントで編集ポイントを案内します。

```text
my-shelf/
  config.toml
  example.toml
```

`shelf root` を環境変数に設定します。

```bash
export SHELVIA_DIR=./my-shelf
```

書籍ファイルのテンプレートを作成します。

```bash
shelvia add "Some Book" --date 2024-01-01
```

`add` は、`Some Book.toml` を作成し、ファイル内の `title` にも指定したタイトルを設定します。作成されたファイルを開き、書籍情報を入力します。

```toml
title = "Some Book"
author = "Some Author"
rating = 90
read_date = 2024-01-01

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
```

`config.toml` に候補値を追加します。`config.toml` は、ジャンル、出版社、判型、レーベル名などの候補値をまとめるための設定ファイルです。

```toml
# List the values you want to use in this shelf.
kind = "shelvia-config"

[values]
# Add the genres you use.
genres = [
  "Novel",
]

# Add the publishers you use.
publishers = [
  "Example Publisher",
]

# Add imprint names if you use them. Leave this as [] if not needed yet.
imprints = [
  "Example Paperback",
]

# Add the editions you use.
# Set imprint_required = true when that edition always needs an imprint.
[[values.editions]]
name = "Paperback"
imprint_required = true

[[values.editions]]
name = "Hardcover"
imprint_required = false
```

データを検証します。

```bash
shelvia validate
```

一覧を表示します。

```bash
shelvia list
```

`list` は `read_date` の新しい順、同じ日付では `title` の昇順で表示します。表示列は `read_date`、`rating`、`title`、`author`、`genre`、`publisher` です。

条件を指定して検索します。

```bash
shelvia search 'rating >= 90'
shelvia search 'genre = "Novel" and publisher = "Example Publisher"'
shelvia search 'rating >= 90 order by rating desc'
```

`search` は、`title`、`author`、`rating`、`read_date`、`genre`、`publisher`、`edition`、`imprint`、`series`、`translator` を対象にした絞り込み条件を受け取ります。`order by <column> [asc|desc]` を末尾に付けると、同じカラムで並び替えできます。`order by` を指定しない場合の並び順と表示列は `list` と同じです。

1 冊の詳細を表示するには、ファイル名ではなくファイル内の `title` を指定します。

```bash
shelvia show "Some Book"
```

生成された TOML ファイルの場所を調べる場合も、`title` を指定します。

```bash
shelvia path "Some Book"
```

同じ `title` の書籍が複数ある場合、`show` と `path` はエラーにし、該当するファイルパスを候補として表示します。

## データ形式

書籍ファイルは TOML で記述します。必須項目は次のとおりです。

- `title`
- `author`
- `rating`
- `read_date`
- `genre`
- `publisher`

任意項目は次のとおりです。

- `edition`
- `imprint`
- `series`
- `translator`
- `thoughts.summary`
- `thoughts.body`

`rating` は `0` から `100` の整数です。`read_date` は TOML の日付として書きます。`genre`、`publisher`、`edition`、`imprint` は `config.toml` と照合します。

任意項目は省略できます。任意項目に空文字を書いた場合も、未指定と同じ扱いになります。

## `config.toml`

`config.toml` は、ジャンル、出版社、判型、レーベル名などの表記揺れを防ぐための設定ファイルです。Shelvia は `shelf root` 直下の `config.toml` から設定値を読み込みます。

`init` が生成する `config.toml` は、`example.toml` がそのまま検証を通る最小構成です。そこから使う候補値だけを追加していきます。使わない候補値の配列は空のままでも構いません。

```toml
# List the values you want to use in this shelf.
kind = "shelvia-config"

[values]
# Add the genres you use.
genres = [
  "Novel",
]

# Add the publishers you use.
publishers = [
  "Example Publisher",
]

# Add imprint names if you use them. Leave this as [] if not needed yet.
imprints = [
  "Example Paperback",
]

# Add the editions you use.
# Set imprint_required = true when that edition always needs an imprint.
[[values.editions]]
name = "Paperback"
imprint_required = true

[[values.editions]]
name = "Hardcover"
imprint_required = false
```

`imprint_required = true` の edition を指定した場合、imprint の省略はエラーになります。imprint を指定する場合は edition も必要です。

## `shelf root`

Shelvia は `shelf root` を 1 つ受け取り、その配下の `.toml` ファイルを再帰的に読み込みます。ただし、`shelf root` 直下の `config.toml` は書籍ファイルではなく設定ファイルとして扱います。直下以外の `config.toml` は誤配置としてエラーになります。

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

Shelvia は `shelf root` 配下の `.toml` ファイルを再帰的に探索します。`config.toml` だけは例外として設定値の読み込みに使います。

`shelf root` は、次の順に決まります。

1. `init PATH` のような `shelf root` 用のコマンド引数、または `--shelf` に指定したディレクトリ
2. 環境変数 `SHELVIA_DIR`

明示的にディレクトリを指定した場合は、環境変数 `SHELVIA_DIR` よりもその値を優先します。どちらも指定されていない場合、Shelvia はエラー終了します。

```bash
export SHELVIA_DIR=~/reading-log
shelvia validate
shelvia --shelf ~/other-reading-log validate
```

## 検証

`validate` は、`shelf root` 全体を読み込めることだけを確認するためのコマンドです。書籍ファイルと `config.toml` を読み込み、必須項目、型、評価、日付、設定ファイル参照、edition と imprint の関係を検証します。ファイルの作成、更新、変換は行いません。

```bash
shelvia validate
```

成功した場合は、読み込んだ件数を表示します。

```text
Validated 1 book, 1 config file.
```

失敗した場合は、ファイルパス、場所、理由を表示します。

```text
Some Book.toml:5: unknown genre "Novel"
```

`validate` が成功した shelf は、`list` や `search` でも同じように読み込める状態です。データを追加・編集したあと、コミット前の確認として使うことを想定しています。
