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
- 必須項目、評価、日付、ジャンル、出版社などの候補値を検証
- SQL 風の条件で検索
- TOML を情報源として扱い、SQLite は検索用の一時ビューとして利用

## インストール

インストール手順は初回リリース前に追加します。

## クイックスタート

読書データを置くディレクトリを用意します。Shelvia では、このディレクトリを `shelf root` と呼びます。

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

書籍ファイルを `books/` 配下に作成します。

```toml
title = "Some Book"
author = "Some Author"
rating = 90
read_date = 2024-01-01

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"

series = ""
translator = ""

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
```

vocab ファイルを `vocab/` 配下に作成します。vocab ファイルは、ジャンル、出版社、判型、レーベル名などの候補値をまとめるためのファイルです。

```toml
# vocab/genres.toml
values = [
  "Novel",
]
```

```toml
# vocab/publishers.toml
values = [
  "Example Publisher",
]
```

```toml
# vocab/editions.toml
[[values]]
name = "Paperback"
imprint_required = true
```

```toml
# vocab/imprints.toml
values = [
  "Example Paperback",
]
```

`shelf root` を環境変数に設定します。

```bash
export SHELVIA_DIR=./my-shelf
```

データを検証します。

```bash
shelvia validate
```

一覧を表示します。

```bash
shelvia list
```

条件を指定して検索します。

```bash
shelvia query --where 'rating >= 90'
shelvia query --where 'genre = "Novel" and publisher = "Example Publisher"'
```

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

`rating` は `0` から `100` の整数です。`read_date` は TOML の日付として書きます。`genre`、`publisher`、`edition`、`imprint` は vocab ファイルと照合します。

## vocab

vocab は、ジャンル、出版社、判型、レーベル名などの表記揺れを防ぐためのファイルです。Shelvia は `shelf root` の `vocab/` 配下から、次のファイルを読み込みます。

- `genres.toml`
- `publishers.toml`
- `editions.toml`
- `imprints.toml`

genre もユーザーが vocab として管理します。使わない vocab がある場合でも、対応するファイルは作成し、空の `values` を置いておきます。

```toml
# vocab/genres.toml
values = [
  "Novel",
  "Essay",
  "Technical",
  "Business",
]
```

```toml
# vocab/publishers.toml
values = [
  "Example Publisher",
  "Another Publisher",
]
```

```toml
# vocab/editions.toml
[[values]]
name = "Paperback"
imprint_required = true

[[values]]
name = "Hardcover"
imprint_required = false
```

```toml
# vocab/imprints.toml
values = [
  "Example Paperback",
]
```

`imprint_required = true` の edition を指定した場合、imprint の省略はエラーになります。imprint を指定する場合は edition も必要です。

## `shelf root`

Shelvia は `shelf root` を 1 つ受け取り、その配下の `books/` と `vocab/` を読み込みます。

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

`books/` 配下では、`.toml` ファイルを再帰的に読み込みます。`vocab/` 配下では、`genres.toml`、`publishers.toml`、`editions.toml`、`imprints.toml` を読み込みます。

`shelf root` は、次の順に決まります。

1. `--shelf` に指定したディレクトリ
2. 環境変数 `SHELVIA_DIR`

`--shelf` を指定した場合は、環境変数 `SHELVIA_DIR` よりもその値を優先します。

```bash
export SHELVIA_DIR=~/reading-log
shelvia validate
shelvia --shelf ~/other-reading-log validate
```

## 検証

`validate` は、`shelf root` 全体を読み込めることだけを確認するためのコマンドです。書籍ファイルと vocab を読み込み、必須項目、型、評価、日付、vocab 参照、edition と imprint の関係を検証します。ファイルの作成、更新、変換は行いません。

```bash
shelvia validate
```

`validate` が成功した shelf は、`list` や `query` でも同じように読み込める状態です。データを追加・編集したあと、コミット前の確認として使うことを想定しています。
