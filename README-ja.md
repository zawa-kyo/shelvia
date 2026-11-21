# Shelvia

Shelvia は、読書記録を手元で管理するための CLI ツールです。読了した本を 1 冊 1 つの TOML ファイルとして保存し、コマンドラインから検証・一覧表示・検索を行えます (未読や読書中の本は対象外です)。

読書データは特定アプリのデータベースに閉じ込められず、テキストエディタで直接閲覧・編集できるプレーンテキストとして残ります。Git によるバージョン管理や `grep`、`find` を用いた操作も容易です。そのうえで、評価、読了日、ジャンル、出版社、レーベルなどの項目に基づいた高度な検索・抽出を可能にします。

## なぜ Shelvia なのか

読書記録は表計算ソフトや読書管理サービスでも管理できます。どちらも手軽で便利ですが、使い続けるうちに「データを自分の手元で持ち続けたい」「特定のサービスや UI に依存したくない」「あとから自由に条件を指定して検索したい」といった課題が生じがちです。

一方で、Markdown などの自由形式のメモだけでは、人間には読みやすくても構造的な検索が難しくなります。

- 評価が 90 点以上の本を探す
- 2024 年に読んだ技術書だけを抽出する
- 特定の出版社やレーベルの本を一覧する

Shelvia は、プレーンテキストの手軽さ・ポータビリティと、データベースのような構造的検索のしやすさを両立するためのツールです。

## 特徴

- 1 冊を 1 つの TOML ファイルとして保存
- 読書データを手元の Git リポジトリ等でそのまま管理
- `init` で `config.toml` と `example.toml` を自動生成
- `add` で書籍ファイルのテンプレートを作成
- 必須項目、評価値、日付、ジャンルや出版社などのマスタ候補値を検証
- SQL ライクな条件式による柔軟な検索
- TOML を唯一の情報源 (Source of Truth) とし、SQLite は検索用の一時ビューとしてのみ利用

## インストール

インストール手順は初回リリース前に追加予定です。

## クイックスタート

読書データを格納するディレクトリを作成します。Shelvia ではこのディレクトリを `shelf root` と呼びます。

```bash
shelvia init ./my-shelf
```

`init` は指定したディレクトリを `shelf root` として初期化し、直下に `config.toml` と `example.toml` を作成します (ディレクトリが存在しない場合は新規作成します)。
`example.toml` は検証対象に含まれるサンプルの書籍データです。どちらのファイルも最小限の記述で生成され、コメントで編集方法を案内しています。

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

`add` は `Some Book.toml` を作成し、ファイル内の `title` に指定したタイトルを設定します。生成されたファイルを開き、書籍情報を入力します。

```toml
title = "Some Book"
author = ["Some Author"]
rating = 90
read_date = 2024-01-01

genre = ["Novel"]
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
```

`config.toml` に許容する候補値を追加します。`config.toml` は、ジャンル、出版社、判型、レーベル名などの表記揺れを防ぐための設定ファイルです。

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

登録データを検証します。

```bash
shelvia validate
```

書籍の一覧を表示します。

```bash
shelvia list
```

`list` は `read_date` の降順 (新しい順)、同日の場合は `title` の昇順で並び替えて出力します。表示列は `read_date`、`rating`、`title`、`author`、`genre`、`publisher` です。

条件を指定して検索します。

```bash
shelvia search 'rating >= 90'
shelvia search 'genre = "Novel" and publisher = "Example Publisher"'
shelvia search 'rating >= 90 order by rating desc'
```

`search` は `title`、`author`、`rating`、`read_date`、`genre`、`publisher`、`edition`、`imprint`、`series`、`translator` を対象とした絞り込み条件を受け取ります。末尾に `order by <column> [asc|desc]` を指定することで並び替えも可能です (省略時の並び順および表示列は `list` と同様)。

特定の 1 冊の詳細を表示するには、ファイル名ではなくファイル内の `title` を指定します。

```bash
shelvia show "Some Book"
```

対象の TOML ファイルのパスを調べる場合も同様に `title` を指定します。

```bash
shelvia path "Some Book"
```

同一の `title` を持つ書籍が複数ある場合、`show` と `path` はエラーとなり、候補となるファイルパスの一覧を表示します。

一覧や検索結果は CSV 形式でも出力できます。全書籍の全項目をまとめてエクスポートするには `export` を使用します。

```bash
shelvia list --format csv > books.csv
shelvia search 'rating >= 90' --format csv > selected.csv
shelvia export > all-books.csv
```

CSV にはヘッダーが含まれ、未指定の任意項目は空欄 (空配列の `translator` は `[]`) となります。`export` には感想 (`thoughts`) を含むすべての書籍データが出力されます (設定ファイルや TOML コメントは含まれません)。

## データ形式

初回リリースでは、1 つの書籍レコードにつき 1 つの読了日のみを扱います (再読ごとの履歴追跡は行いません)。

書籍データは TOML 形式で記述します。必須項目は以下のとおりです。

- `title`: タイトル
- `author`: 著者名 (1 件以上の文字列配列)
- `rating`: 評価値 (0〜100 の整数)
- `read_date`: 読了日 (TOML の日付形式)
- `genre`: ジャンル (1 件以上の文字列配列、`config.toml` と照合)
- `publisher`: 出版社 (`config.toml` と照合)

任意項目は以下のとおりです。

- `edition`: 判型・版 (`config.toml` と照合)
- `imprint`: レーベル (`config.toml` と照合)
- `series`: シリーズ名
- `translator`: 訳者名 (文字列配列)
- `thoughts.summary`: 感想の要約
- `thoughts.body`: 感想の本文

`author` と `genre` は 1 件以上の文字列を含む配列です。`translator` は任意の文字列配列で、省略可能です。各配列要素に空文字は指定できません (後方互換性のため従来の単一文字列形式も読み込み可能です)。検索時は配列の各要素に対して条件照合が行われます。CSV 出力時は各配列を JSON 配列形式で 1 つのセルに出力します。

任意項目は省略可能であり、空文字が指定された場合も未指定として扱われます。

## `config.toml`

`config.toml` は、ジャンル、出版社、判型、レーベル名などの表記揺れを防ぐためのマスタ設定ファイルです。Shelvia は `shelf root` 直下の `config.toml` から設定値を読み込みます。

`init` で生成される `config.toml` は、サンプルの `example.toml` がそのまま検証を通過する最小限の構成です。必要に応じて使用する候補値を追加してください。使わない項目の配列は空のままでも問題ありません。

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

`imprint_required = true` が設定された `edition` を指定した場合、`imprint` の指定が必須となります。また、`imprint` を指定する場合は対応する `edition` も指定する必要があります。

## `shelf root`

Shelvia は指定された 1 つの `shelf root` 配下の `.toml` ファイルを再帰的に読み込みます。ただし、`shelf root` 直下に配置された `config.toml` のみは設定ファイルとして扱われます。直下以外に配置された `config.toml` は誤配置とみなされエラーになります。

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

`shelf root` は以下の優先順位で決定されます。

1. コマンド引数 (`init PATH` など) または `--shelf` オプションで指定されたディレクトリ
2. 環境変数 `SHELVIA_DIR`

ディレクトリが明示的に指定された場合は環境変数よりも優先されます。どちらも指定されていない場合、Shelvia はエラー終了します。

```bash
export SHELVIA_DIR=~/reading-log
shelvia validate
shelvia --shelf ~/other-reading-log validate
```

## 検証

`validate` は、`shelf root` 全体のデータが正しく読み込めるかを確認するためのコマンドです。書籍ファイルと `config.toml` を読み込み、構文、必須項目、データ型、評価値の範囲、日付形式、設定ファイルの参照整合性、判型とレーベルの関係を検証します。ファイルの新規作成・更新・変換は一切行いません。

```bash
shelvia validate
```

検証に成功した場合、読み込んだ件数が表示されます。

```text
Validated 1 book, 1 config file.
```

検証に失敗した場合は、対象のファイルパス、該当箇所、エラー理由が表示されます。

```text
Some Book.toml:5: unknown genre "Novel"
```

`validate` が成功する shelf は、`list` や `search` でも正常に処理できます。新しい書籍データの追加や編集を行ったあとの確認 (Git コミット前の検証など) にご利用ください。
