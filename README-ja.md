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
- `new` で書籍ファイルのひな形を作成
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

`init` は、指定した `shelf root` に `config.toml` と `example.toml` を作成します。
`example.toml` は、検証対象に含まれる書籍データのサンプルファイルです。

```text
my-shelf/
  config.toml
  example.toml
```

`shelf root` を環境変数に設定します。

```bash
export SHELVIA_DIR=./my-shelf
```

書籍ファイルのひな形を作成します。

```bash
shelvia new "Some Book" --read-date 2024-01-01
```

`new` は、`Some Book.toml` を作成します。作成されたファイルを開き、書籍情報を入力します。

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
kind = "shelvia-config"

[values]
genres = [
  "Novel",
]

publishers = [
  "Example Publisher",
]

imprints = [
  "Example Paperback",
]

[[values.editions]]
name = "Paperback"
imprint_required = true
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
shelvia query --where 'rating >= 90'
shelvia query --where 'genre = "Novel" and publisher = "Example Publisher"'
```

`query --where` は、`title`、`author`、`rating`、`read_date`、`genre`、`publisher`、`edition`、`imprint`、`series`、`translator` を対象にした絞り込み条件を受け取ります。初回リリースでは `order by` は指定できず、並び順と表示列は `list` と同じです。

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

genre も `config.toml` で管理します。使わない候補値がある場合でも、対応する配列は作成し、空の配列を置いておきます。

```toml
kind = "shelvia-config"

[values]
genres = [
  "Novel",
  "Essay",
  "Technical",
  "Business",
]

publishers = [
  "Example Publisher",
  "Another Publisher",
]

imprints = [
  "Example Paperback",
]

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

`validate` が成功した shelf は、`list` や `query` でも同じように読み込める状態です。データを追加・編集したあと、コミット前の確認として使うことを想定しています。
