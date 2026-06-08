# Testing

## 方針

一般的なテスト戦略を取ります。Domain は単体テストで細かく確認し、それ以外は外部依存を port の裏で差し替えた結合テストで確認します。

## Test Structure

各テストケースは AAA パターンで構成します。Arrange で入力、fixture、fake、期待値を準備し、Act で対象の処理を実行し、Assert で結果を検証します。

AAA の区切りはコメントではなく空行で表します。`// Arrange`、`// Act`、`// Assert` のようなコメントは書きません。テスト名と変数名だけで意図が読めるようにします。

Assertion は `github.com/stretchr/testify/require` に揃えます。期待と違う時点でテストを止めたいケースが多いため、基本は `assert` ではなく `require` を使います。手書きの `if` と `t.Fatal` は、テスト用 fake の振る舞いを表す場合を除き避けます。

```go
func TestExample(t *testing.T) {
	t.Run("必要な項目がそろっていれば本を記録できる", func(t *testing.T) {
		draft := validBookDraft()
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err)
		require.Equal(t, "Some Book", book.Title().String())
	})
}
```

Act は原則として 1 つにします。CLI の一連のユーザーフローを確認する結合テストのように、複数の操作がシナリオそのものを表す場合は、操作ごとに小さな AAA を繰り返します。

## Domain Unit Tests

Domain の単体テストでは、外部 I/O を使わずに純粋なルールを確認します。

- 必須項目
- `rating`
- `read_date`
- 設定ファイル由来の許可値との照合
- 任意項目の空文字が未指定として扱われること
- edition と imprint の関係

Go のテストは package 単位で実行されるため、プロダクションコードとテストコードを必ず 1 ファイルずつ対応させる必要はありません。ただし、ドメイン層では tactical DDD の役割が読み取れるように、概念単位で対応が分かる粒度にします。

初回実装では次の対応を基本にします。

```text
book.go / book_draft.go      -> book_test.go
allowed_values.go            -> allowed_values_test.go
author.go, rating.go, etc.   -> value_objects_test.go
```

`value_objects_test.go` が肥大化し、対象の value object を探しにくくなった場合は、`title_test.go`、`rating_test.go` のように value object ごとの test file へ分割します。最初から薄い test file を大量に作るより、読みやすさが落ちた時点で分割します。

## Integration Tests

Application、CLI、presentation、local filesystem 境界は結合テストで確認します。インフラ層の依存先、特に query engine の DB は mock または fake adapter に差し替えます。

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

DB そのものの挙動は query engine adapter の薄い adapter test に限定します。主な振る舞いは、DB に依存しない application-level integration test で検証します。

Fixture shelf は `fixtures/shelves/` 配下に置きます。`valid/` は正常系、`invalid/` は失敗系に分け、各 fixture directory をそのまま `shelf root` として読み込ませます。

## E2E Tests

E2E test は `e2e/` に置き、`make e2e` から build tag `e2e` 付きで実行します。これは Go の integration test を置き換えるものではなく、実際の CLI binary、環境変数、終了コード、標準出力を含む主要導線の smoke test として扱います。

E2E test では、一時ディレクトリに shelf を作成して `init`、`validate`、`list`、`query` を確認します。並び替えのように複数データが必要な確認では、`fixtures/shelves/valid/minimal` を使い、出力内のタイトル順を機械的に検証します。
