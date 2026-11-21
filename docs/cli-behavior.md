# CLI Behavior

本書は Shelvia の CLI における振る舞い (コマンド仕様、`shelf root` の解決順序、ファイル読み込みルール、検索・表示仕様) を定義します。

## Shelf モデル

Shelvia は指定された 1 つの `shelf root` ディレクトリを基点とし、その配下にある `.toml` ファイルを再帰的に探索して読み込みます。

- **`config.toml` の扱い**: `shelf root` の直下に配置された `config.toml` のみを設定ファイル (ジャンル・出版社・判型・レーベル等のマスタ定義) として扱います。サブディレクトリ内に `config.toml` が存在する場合は誤配置とみなし、エラーとして処理を停止します。
- **`example.toml` の扱い**: `init` 時に生成される書籍サンプルです。テンプレートではなく検証対象の実データとして扱われるため、生成直後から `config.toml` と整合し、バリデーションに合格する内容で出力されます。

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

### `shelf root` の解決優先順位

全コマンド共通で、`shelf root` は以下の優先順位で解決されます。

1. コマンド引数 (`shelvia init PATH` など) または `--shelf <dir>` オプション
2. 環境変数 `SHELVIA_DIR`

明示的な引数・オプション指定は環境変数より常に優先されます。いずれも指定されていない場合、Shelvia はエラーメッセージを出力して終了します。

## コマンド仕様

初期リリースで提供するコマンドは以下のとおりです。

- `shelvia init [PATH]`
- `shelvia add TITLE --date YYYY-MM-DD`
- `shelvia validate`
- `shelvia list`
- `shelvia search SQL_FRAGMENT`
- `shelvia show TITLE`
- `shelvia path TITLE`
- `shelvia export`

### `init [PATH]`

指定されたディレクトリを `shelf root` として初期化し、最小構成の `config.toml` と `example.toml` を作成します。

- `PATH` を省略した場合はカレントディレクトリを対象とします。
- 対象ディレクトリが存在しない場合は新規作成します。
- 既存の `config.toml` や `example.toml` が存在する場合は上書きせず、エラーにします。

### `add TITLE --date YYYY-MM-DD`

書籍ファイル `<title>.toml` を新規作成し、タイトルと読了日 (`read_date`) を反映したテンプレートを出力します。

- 安全なファイル名規則に従ってファイル名を生成します (後述)。
- 既に同名のファイルが存在する場合は上書きせずエラーにします。

### `validate`

`shelf root` 全体の書籍ファイルおよび `config.toml` を読み込み、構文、型、制約を検証します。

- ファイルの追加・変更・変換は行いません。
- 成功時は検証した書籍数と設定ファイル数を表示します。
- 失敗時は該当ファイル名、行番号、エラー詳細を表示します。

### `list`

検証済みの書籍一覧を既定順で表示します。

- **既定順**: `read_date desc, title asc`
- **表示列**: `read_date`, `rating`, `title`, `author`, `genre`, `publisher`
- **オプション**: `--format csv` を指定すると、ヘッダー付き CSV 形式で標準出力に出力します (既定値は `--format table`)。

### `search SQL_FRAGMENT`

検証済みの書籍から構築した一時的な検索ビューに対し、SQL ライクな条件式を用いて絞り込み検索を行います。

- 既定のソート順および表示列は `list` と共通です。
- クエリの末尾に `order by <column> [asc|desc]` を指定することで、単一カラムでの並び替えが可能です (昇順・降順を省略した場合は `asc`)。
- `list` と同様に `--format csv` での出力に対応します。

### `show TITLE`

ファイル名ではなく、書籍データ内の `title` に完全一致する書籍 1 件の詳細情報を表示します。

- 該当する書籍が存在しない場合はエラーにします。
- 同一の `title` を持つ書籍が複数存在する場合はエラーとし、重複しているファイルパス一覧を候補として表示します。

### `path TITLE`

書籍データ内の `title` に完全一致する書籍の TOML ファイルパスを表示します。

- 該当なし、またはタイトル重複時の挙動は `show` と同様です。

### `export`

検証済みの全書籍の全項目を CSV 形式で標準出力に出力します。

- **出力列の順序**: `title`, `author`, `rating`, `read_date`, `genre`, `publisher`, `edition`, `imprint`, `series`, `translator`, `thoughts.summary`, `thoughts.body`
- 未指定の任意項目は空文字、空の `translator` 配列は `[]` として出力されます。
- 書籍データのみを出力対象とし、`config.toml` や TOML コメントは含まれません。

## 検索条件の仕様

`search` コマンドで利用可能なカラムおよび演算子は以下のとおりです。

### 対応カラム

- `title`
- `author`
- `rating`
- `read_date`
- `genre`
- `publisher`
- `edition`
- `imprint`
- `series`
- `translator`

### 対応演算子・構文

- 比較: `=`, `!=`, `<`, `<=`, `>`, `>=`
- パターンマッチ: `like`
- NULL 判定: `is null`, `is not null`
- 論理演算: `and`, `or`, `not`, 括弧 `(...)`
- ソート句: `order by <column> [asc|desc]`

### 値の解釈と型ルール

- **文字列・日付**: クォートで囲んだ文字列として指定します。日付は `YYYY-MM-DD` 形式で比較します。
- **NULL の扱い**: 任意項目の空文字は未指定 (`null`) として扱われます。SQL 準拠の三値論理で評価され、結果が真となるレコードのみが返されます。
- **配列項目 (`author`, `genre`, `translator`)**:
  - `=` や `like` などの比較は各要素に適用され、**いずれか 1 つでも一致すれば真**となります。
  - `!=` は**すべての要素と一致しない場合にのみ真**となります。
  - 空の `translator` は `null` として扱われます。
  - 配列カラムに対するソートは、先頭要素から順に文字列比較を行います。
  - CSV 出力時は JSON 配列文字列として 1 つのセルにシリアライズされます。

### 拒否する構文

安全性を担保するため、以下の構文を含むクエリはアダプター境界で拒否されます。

- DML / DDL: `select`, `insert`, `update`, `delete`, `create`, `alter`, `drop`
- 結合・複合クエリ: `join`, `union`, サブクエリ
- 関数呼び出し、セミコロンによる複数文、SQL コメント

## ファイル名生成規則

`shelvia add` が生成するファイル名は、クロスプラットフォーム (macOS, Linux, Windows) で安全に扱えるよう正規化されます。

- パス区切り文字 (`/`, `\`)、NUL 文字、制御文字を拒否
- Windows の禁止文字 `< > : " / \ | ? *` を拒否
- `.`、`..`、空文字、前後が空白のみのタイトルを拒否
- Windows の予約名 (`CON`, `PRN`, `AUX`, `NUL`, `COM1`〜`COM9`, `LPT1`〜`LPT9`) を拒否
- 生成先に同名ファイルが既に存在する場合は上書きせずエラー
- 拡張子は一律で `.toml` を付与 (二重拡張子を防止)

## データフロー

`validate`, `list`, `search`, `show`, `path`, `export` は一貫して以下の共通パイプラインを通ります。

```text
1. shelf root の解決
2. shelf root 直下の config.toml の読み込みとマスタ定義の検証
3. 配下の書籍 TOML ファイルの探索と読み込み
4. TOML 構造体からドメインモデル (BookDraft) への変換
5. ドメインルールによるバリデーション (Book の構築)
6. 検証結果の判定 (エラーがあればファイルパス付きで即時診断を出力)
7. list / search の場合、検証済み Book からインメモリ SQLite ビューを構築
8. 目的のフォーマット (テーブル表示または CSV) で結果を出力
```

未検証のデータが検索ビューや出力レイヤーに渡ることはありません。
