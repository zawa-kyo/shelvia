# Testing

## 方針

一般的なテスト戦略を取ります。Domain は単体テストで細かく確認し、それ以外は外部依存を port の裏で差し替えた結合テストで確認します。

## Domain Unit Tests

Domain の単体テストでは、外部 I/O を使わずに純粋なルールを確認します。

- 必須項目
- `rating`
- `read_date`
- 設定ファイル由来の許可値との照合
- 任意項目の空文字が未指定として扱われること
- edition と imprint の関係

## Integration Tests

Application、CLI、presentation、filesystem 境界は結合テストで確認します。インフラ層の依存先、特に SQLite などの DB は mock または fake adapter に差し替えます。

- `shelf root` の解決と、未指定時のエラー
- `init` のファイル作成方針
- `new` の作成パスと TOML テンプレート
- `new` のファイル名検証と既存ファイル衝突
- `config.toml` が書籍ファイルとして扱われないこと
- 直下以外の `config.toml` が明示エラーになること
- TOML parsing から application command までの流れ
- ファイルパスを含む検証エラー
- `validate`、`list`、`query` の command output
- `list` と `query` の既定順が `read_date desc, title asc` であること
- `list` と `query` の表示列が揃っていること
- 対応する `--where` fragment が query port に渡ること
- 未対応または危険な SQL fragment が adapter 境界で拒否されること

DB そのものの挙動は SQLite adapter の薄い adapter test に限定します。主な振る舞いは、DB に依存しない application-level integration test で検証します。
