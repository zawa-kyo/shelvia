# CLI Behavior

このドキュメントは、Shelvia の CLI から見える挙動を定義します。各コマンドが `shelf root` をどう解決し、shelf 内のファイルをどう読み込み、結果をどう表示・検索するかを扱います。

## Shelf Model

Shelvia は 1 つの `shelf root` を受け取り、その配下の `.toml` ファイルを再帰的に読み込みます。`shelf root` 直下の `config.toml` だけは例外で、書籍ファイルではなく設定ファイルとして扱います。直下以外の `config.toml` は書籍として読まず、誤配置として明示的にエラーにします。

`config.toml` には、ジャンル、出版社、判型、レーベル名などの許可値を列挙します。`init` は、`example.toml` がそのまま検証を通る最小構成を生成し、編集の起点になるコメントも書き込みます。

`example.toml` はテンプレートではなく、検証対象に含まれる実データの例として扱います。`init` 直後の `validate` が成功するように、`example.toml` と `config.toml` は互いに整合する内容で生成します。

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

`shelf root` は、全コマンドで次の順に解決します。

1. `init PATH` のような shelf root 用のコマンド引数、または `--shelf` に指定したディレクトリ
2. 環境変数 `SHELVIA_DIR`

複数の指定方法がある場合は、明示的なコマンド入力を環境変数より優先します。どれも指定されていない場合はエラー終了し、`--shelf` を渡すか `SHELVIA_DIR` を設定するように案内します。

## Commands

初回リリースでは、次のコマンドを対象にします。

- `shelvia init [PATH]`
- `shelvia add TITLE --date YYYY-MM-DD`
- `shelvia validate`
- `shelvia list`
- `shelvia search SQL_FRAGMENT`
- `shelvia show TITLE`
- `shelvia path TITLE`

`init` は、指定されたディレクトリを `shelf root` として初期化し、`config.toml` と `example.toml` を作成します。`PATH` を省略した場合は、カレントディレクトリを対象にします。対象ディレクトリが存在しない場合は新しく作成します。`example.toml` は書籍ファイルとして扱われるため、生成直後から validation に通る内容にします。生成内容は最小構成にとどめ、追加で編集してほしい箇所は TOML コメントで案内します。既存の `config.toml` や `example.toml` は上書きせずエラーにします。

`add` は、`<title>.toml` に書籍ファイルのテンプレートを作成し、ファイル内の `title` にも指定されたタイトルを設定します。CLI の日付オプションは `--date` とし、生成する TOML では `read_date` に書き込みます。

`validate` は、`shelf root` 全体を読み込み、`list` と `search` でも使える状態かを確認します。

`list` は、検証済みの書籍を既定順で表示します。既定順は `read_date desc, title asc` です。初回リリースでは、表示列を `read_date`、`rating`、`title`、`author`、`genre`、`publisher` に固定します。

`search` は、SQL 風の条件を受け取ります。内部的には独自 DSL ではなく、検証済みデータから作った一時検索ビューに対する制限付き fragment として扱います。

`search` の既定順と表示列は `list` と同じです。検索条件の末尾には `order by <column> [asc|desc]` を 1 つ指定できます。方向を省略した場合は `asc` として扱います。

`show TITLE` は、ファイル名ではなくファイル内の `title` で書籍を 1 件特定し、詳細を表示します。該当なし、または同じ `title` が複数ある場合はエラーにします。複数ある場合は、該当するファイルパスを候補として表示します。

`path TITLE` は、ファイル名ではなくファイル内の `title` で書籍を 1 件特定し、その TOML ファイルのパスを表示します。該当なし、または同じ `title` が複数ある場合の扱いは `show` と同じです。

対応するカラムは次のとおりです。

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

対応する演算子は次のとおりです。

- `=`
- `!=`
- `<`
- `<=`
- `>`
- `>=`
- `like`
- `is null`
- `is not null`
- `and`
- `or`
- `not`
- `(...)`
- `order by <column> [asc|desc]`

文字列と日付はクォートされた文字列として指定します。日付は `YYYY-MM-DD` 形式で比較します。任意項目の空文字は未指定と同じ扱いになるため、検索ビュー上では `null` として扱います。`null` を含む条件は SQL と同じ三値論理で評価し、結果が真になる行だけを返します。

初回リリースでは、次のような構文を拒否します。

- `select`、`insert`、`update`、`delete`
- `create`、`alter`、`drop`
- `join`、`union`
- サブクエリ
- 関数呼び出し
- 複数文
- コメント

## File Name Rules

`add` が生成するファイル名は macOS、Linux、Windows で安全に扱えるものに正規化します。少なくとも次のルールを満たします。

- パス区切り文字、NUL、制御文字を拒否する
- Windows で使えない文字 `< > : " / \ | ? *` を拒否する
- `.`、`..`、空文字、前後空白だけのタイトルを拒否する
- Windows の予約名 `CON`、`PRN`、`AUX`、`NUL`、`COM1` から `COM9`、`LPT1` から `LPT9` を拒否する
- 生成先に同名ファイルがある場合は上書きせずエラーにする
- ファイル名は `<title>.toml` とし、拡張子をユーザー入力から二重に付けない

## Data Flow

`validate`、`list`、`search`、`show`、`path` は同じ読み込み経路を使います。

1. `shelf root` を解決する
2. `config.toml` を設定ファイルとして読み込む
3. `config.toml` 以外の `.toml` ファイルを書籍ファイルとして読み込む
4. TOML の raw data を domain value に変換する
5. domain ルールを検証する
6. 診断または検証済み書籍を返す
7. `list` と `search` では、検証済み書籍からインメモリ SQLite ビューを作る
8. 結果を表示する

未検証のデータを `list`、`search`、`show`、`path` に渡してはいけません。
