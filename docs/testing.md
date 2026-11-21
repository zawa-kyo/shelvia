# Testing

本書は Shelvia のテスト戦略、テストコードの構造、および各テストレイヤーの設計方針を定義します。

## 全体方針

Shelvia では以下の 3 つのテストレイヤーを組み合わせて品質を担保します。

1. **Domain 単体テスト**: 外部 I/O を一切排し、ドメインのビジネスルールと不変条件を網羅的に検証します。
2. **Integration テスト**: アプリケーション層を中心に、各アダプターの境界やユースケースの連携を検証します (クエリエンジン等の重い依存は Fake/Mock に差し替えます)。
3. **E2E テスト**: ビルドされた CLI バイナリを実行し、環境変数、終了コード、標準出力を通じた主要シナリオのスモークテストを実施します。

## テスト構造の規約

### AAA パターンの徹底

すべてのテストケースは **Arrange(準備)**、**Act(実行)**、**Assert(検証)** の 3 フェーズで記述します。

- 各フェーズの区切りは空行で表し、`// Arrange` などのコメントは記載しません。
- 原則として 1 つのテストケースにおける Act は 1 回とします (複数ステップのユーザーフローを検証するシナリオテストを除く)。

### アサーションライブラリ

アサーションには `github.com/stretchr/testify/require` を統一して使用します。

- 期待値と不一致が生じた時点でテストを直ちに失敗させるため、`assert` ではなく `require` を基本とします。
- 手書きの `if ... t.Fatalf` は、テストヘルパーや Fake の実装内を除き使用しません。

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

## ドメイン単体テスト (Domain Unit Tests)

ドメイン層のテストでは、ファイルシステムやデータベースに一切触れず、メモリ上のみで純粋なビジネスロジックを検証します。

### 主な検証対象

- 必須項目の存在チェック
- `rating` の許容範囲 (0〜100)
- `read_date` の日付形式
- `AllowedValues` との照合 (ジャンル、出版社、判型、レーベル)
- 任意項目における空文字の未指定扱い
- 判型 (`edition`) とレーベル (`imprint`) の相関ルール

### ファイル構成

初期実装では概念単位で以下の対応を基本とします。

```text
book.go / book_draft.go      -> book_test.go
allowed_values.go            -> allowed_values_test.go
author.go, rating.go, etc.   -> value_objects_test.go
```

`value_objects_test.go` が肥大化して保守性が低下した段階で、`title_test.go` や `rating_test.go` などの個別ファイルへの分割を検討します。

## 結合テスト (Integration Tests)

アプリケーション層を中心に、CLI 入力、ローカルファイルシステム、出力整形のアダプター境界を結合して検証します。外部 I/O が関わるインフラ (クエリエンジンなど) は Port の背後で Fake アダプター等に差し替えて実行を高速化します。

### 主な検証対象

- `shelf root` の解決優先順位と、未指定時のエラーハンドリング
- `init` による初期ファイル生成と既存ファイルの上書き防止
- `add` によるテンプレート作成と安全なファイル名バリデーション
- `config.toml` の特殊扱い (書籍ファイルとして扱われないこと、不正配置時のエラー)
- TOML パースからドメインバリデーションまでのエラーハンドリング (ファイルパスと行番号を含む診断出力)
- `validate`, `list`, `search`, `show`, `path`, `export` の各コマンド出力
- `list` と `search` の既定ソート順 (`read_date desc, title asc`) および出力列の一致
- 許可された検索フラグメントのクエリポートへの受け渡し
- 不正・危険な SQL フラグメントのアダプター境界での拒否
- `show` および `path` における `title` による書籍特定と重複タイトルのエラーハンドリング

### Fixture の利用

実際の shelf 構造を検証する際は、`fixtures/shelves/` 配下のディレクトリ (`valid/`、`invalid/`) をそのまま `shelf root` として読み込ませます。

## E2E テスト (E2E Tests)

E2E テストは `e2e/` ディレクトリに配置し、ビルドタグ `e2e` を付与して `make e2e` から実行します。

- 実際の CLI バイナリをコンパイルし、サブプロセスとして呼び出します。
- 一時ディレクトリ上で `init`、`add`、`validate`、`list`、`search` の一連のコマンドを実行し、終了コードおよび標準出力を検証します。
- 並び替え等の複数件データを必要とするテストには `fixtures/shelves/valid/minimal` を活用します。
