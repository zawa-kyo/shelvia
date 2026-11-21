# Implementation

本書は Shelvia のレイヤー構造、ディレクトリ配置、各層の責務、および依存関係のルールを定義します。

## レイヤー構成と依存規則

Shelvia は Clean Architecture を採用し、外部入出力の境界には Ports and Adapters (Hexagonal Architecture) パターンを適用します。

```text
[外側]
  CLI / localfs / queryengine / presentation (Adapters)
    ↓
  Application (Use Cases & Ports)
    ↓
  Domain (Entities, Value Objects & Domain Rules)
[内側]
```

### 依存規則

- **単方向の内側依存**: 依存関係は外側から内側へ向かい、内側のドメイン層およびアプリケーション層は外側の具象実装 (ファイルシステム、SQLite、CLI 引数パーサーなど) を一切参照しません。
- **Ports & Adapters**: アプリケーション層が必要な外部操作をポート (インターフェース) として定義し、外側のアダプター層がそれを実装します。
- **禁止事項**:
  - `internal/domain` から `internal/application` や `internal/adapter` への依存
  - `internal/application` から `internal/adapter` への依存
  - `internal/adapter/queryengine` の内部型 (SQLite 固有の型など) をドメインやアプリケーション層へ漏洩させること
  - CLI パーサーの都合 (フラグ解析用構造体など) をドメインモデルに持ち込むこと

## ディレクトリ構成

初期リリースでは過度なパッケージ細分化を避け、Clean Architecture の境界が明確に保たれる最小限の粒度で構成します。

```text
.
├── cmd/
│   └── shelvia/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── book.go
│   │   ├── book_draft.go
│   │   ├── allowed_values.go
│   │   ├── author.go
│   │   ├── edition.go
│   │   ├── file_path.go
│   │   ├── genre.go
│   │   ├── imprint.go
│   │   ├── optional_text.go
│   │   ├── publisher.go
│   │   ├── rating.go
│   │   ├── read_date.go
│   │   └── title.go
│   ├── application/
│   │   ├── commands.go
│   │   ├── ports.go
│   │   ├── shelf_service.go
│   │   └── query_service.go
│   └── adapter/
│       ├── cli/
│       │   ├── parser.go
│       │   └── runner.go
│       ├── localfs/
│       │   ├── shelf_repository.go
│       │   └── templates.go
│       ├── queryengine/
│       │   └── query_store.go
│       └── presentation/
│           └── output.go
├── docs/
│   ├── architecture.md
│   ├── cli-behavior.md
│   ├── implementation.md
│   └── testing.md
├── fixtures/
│   └── shelves/
│       ├── valid/
│       │   └── minimal/
│       │       ├── config.toml
│       │       ├── example.toml
│       │       └── Some Book.toml
│       └── invalid/
│           ├── nested-config/
│           ├── missing-required/
│           ├── unknown-config-value/
│           └── unsafe-filename/
├── README-ja.md
├── README.md
└── AGENTS.md
```

## 各レイヤーの責務

### `cmd/shelvia`

アプリケーションのエントリーポイントです。CLI 引数の受け取り、各アダプターの生成、依存関係の注入 (DI)、および実行のみを担当し、ビジネスロジックは一切持ちません。

### `internal/domain`

書籍および設定ファイルに由来するコアビジネスルールと不変条件を保持します。

- **`Book`**: 書籍記録を表す Entity / Aggregate Root。一意な識別とライフサイクルを持ちます。
- **`BookDraft`**: TOML から抽出された未検証の入力データを表す構造体。`Book` を構築するための中間データであり、永続的な同一性は持ちません。
- **`AllowedValues`**: `config.toml` から読み込まれたジャンル、出版社、判型、レーベルなどの許可値セット。ドメイン層の照合に利用します (TOML のスキーマ定義自体はアダプター層に隠蔽します)。
- **Value Objects (`Title`, `Author`, `Rating`, `ReadDate`, `Genre`, etc.)**: 各属性の制約 (値の範囲、空文字禁止など) を自己検証する不変オブジェクト。
- **パッケージ設計方針**: `internal/domain` は技術的な分類 (`value_objects/` など) で細分化せず、1 つの Go パッケージ内に保ちます。将来分割が必要になった場合は、ドメイン概念単位 (`internal/domain/book` など) での分割を検討します。

### `internal/application`

ユースケースの調停と実行フローの制御を行います。

- コマンド (`validate`, `list`, `search`, `add`, `init` など) の実行オーケストレーション
- ポートインターフェース (`ShelfRepository`, `QueryStore` など) の定義
- ファイル上書きの禁止判定 (`init` や `add` での既存ファイル衝突防止)
- 具体的なストレージやクエリエンジンには依存せず、ドメインモデルとポートのみを操作します。

### `internal/adapter/cli`

OS のコマンドライン引数を解析し、アプリケーション層のコマンドオブジェクトに変換して実行を委譲します。

### `internal/adapter/localfs`

ローカルファイルシステム上の `shelf root` の探索、ファイル入出力を担当します。

- `shelf root` 配下の `.toml` ファイルの再帰的探索
- `config.toml` の読み込みと構文解析
- `init` による初期ファイル (`config.toml`, `example.toml`) の書き出し
- `add` によるテンプレート作成と、OS 非依存の安全なファイル名バリデーション
- エラー診断時にユーザーが直感的に修正できるよう、相対パスや行番号を保持します。

### `internal/adapter/queryengine`

検証済みの書籍コレクションからインメモリの一時的な検索ビューを構築し、アプリケーション層の検索ポートを実装します。

- 内部実装として SQLite を利用しますが、パッケージ名やインターフェースに「SQLite」という技術名を漏らしません。
- `search` コマンドで受け取った SQL ライクなクエリフラグメントを、[CLI Behavior](./cli-behavior.md) で定義された許可範囲内に収まっているか事前に検証し、不正な構文を拒否します。

### `internal/adapter/presentation`

コマンドの実行結果を端末向けに整形して出力します。

- `validate` の成否 (件数またはエラー箇所) のメッセージ整形
- `list` および `search` のテーブル表示、CSV 出力
- `show` による詳細情報表示、`path` によるファイルパス表示

## Fixture Shelves

`fixtures/shelves/` は、結合テストや手動検証で利用する実際の shelf データを配置するディレクトリです。

- **`valid/`**: 正常に検証を通過する shelf。`minimal/` は `README` で示される基本構成を再現しています。
- **`invalid/`**: 特定のバリデーション違反を検証するための shelf。1 ディレクトリにつき 1 つのエラー原因 (設定ファイル重複、必須項目欠落、未登録のジャンル、不正なファイル名など) を再現し、テスト名と対応させます。
- **モックとの分離**: `fixtures/` は本物のファイル構造としてそのまま読み込ませるものであり、テストダブル (Mock/Fake) のコードとは明確に区別します。
