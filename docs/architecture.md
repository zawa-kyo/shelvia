# Architecture

## 目的

このドキュメントは、Shelvia のアーキテクチャ全体像を示す入口です。詳細な仕様や実装方針は、粒度ごとに分けた `docs/` 配下の文書に置きます。

Shelvia は、個人の読書記録を扱う小さな CLI ツールです。管理対象の情報は `shelf root` 配下の TOML ファイルであり、SQLite は検証済みデータから起動時に作る一時的な検索ビューとしてのみ扱います。

OpenAI のハーネスエンジニアリングの考え方に従い、`AGENTS.md` は短い目次として保ち、実装に必要な永続的な知識は `docs/` 配下の Markdown に置きます。エージェントがコード、設計、制約、検証方法をリポジトリ内で発見できる状態を重視します。

参考:

- [OpenAI: ハーネスエンジニアリング](https://openai.com/ja-JP/index/harness-engineering/)

## 詳細文書

- [CLI Behavior](./cli-behavior.md): `shelf root`、`config.toml`、読み込みルール、CLI コマンド仕様
- [Implementation](./implementation.md): Clean Architecture、ディレクトリ構成、依存方向、adapter 責務
- [Testing](./testing.md): domain 単体テスト、application-level integration test、adapter test の方針

## 設計目標

- 読書データを、Shelvia の外からも編集・差分管理できるプレーンテキストとして保つ
- 検証を十分に厳しくし、検索結果を信頼できる状態にする
- 初期実装は小さく、読み返しやすい構成にする
- 場当たり的な分岐より、明示的な境界と不変条件を優先する
- エラーでは、ユーザーが直すべきファイルと理由を示す

## 非目標

- Shelvia はアプリケーション DB を永続的な情報源として所有しない
- Shelvia は汎用的なテキスト検索ツールを置き換えない
- 初回リリースでは、プラグイン機構やサービス構成を持たない

## リポジトリ知識の置き場所

`README.md` と `README-ja.md` は、ユーザー向けの説明を置く場所です。動機、インストール、使い方、データ形式、例を中心にし、実装計画やアーキテクチャ詳細は置きません。

`docs/` は、実装者とエージェント向けの設計知識を置く場所です。コマンドの責務、データフロー、検証境界、不変条件が変わる場合は、該当する設計文書も同じ変更で更新します。

`AGENTS.md` は、詳細な手順書ではなく目次として扱います。深い情報を重複して書かず、信頼できる Markdown への入口を示します。

## アーキテクチャ方針

Shelvia は Clean Architecture に則ります。依存方向は常に内側へ向け、domain は application、CLI、filesystem、SQLite、presentation に依存しません。

Onion Architecture と Hexagonal Architecture はどちらも候補になりますが、この CLI では Clean Architecture を基本方針にし、外部依存との境界表現として Hexagonal Architecture の ports/adapters を採用します。

理由は次のとおりです。

- CLI、filesystem、SQLite、presentation という外部境界が明確で、ports/adapters と相性がよい
- Domain を中心に置く点は Onion Architecture と同じだが、CLI ツールでは「どの外部入出力を adapter として差し替えるか」を明示した方が実装しやすい
- Clean Architecture の依存ルールを上位方針にすれば、Onion と Hexagonal のよい部分を過不足なく使える

## Invariants

実装が大きくなっても、次のルールは保ちます。

- 永続化は TOML ファイルのみで行い、DB は永続化しない
- SQLite は TOML から再構築され、SQLite 側の変更を TOML へ書き戻さない
- domain validation は SQLite に投入する前に完了する
- ファイル由来のエラーは、ユーザーが直すべきファイルパスを含む
- 依存方向は Clean Architecture の内側へ向ける
- `README.md` と `README-ja.md` は、ユーザー向けの挙動について同期する
- `AGENTS.md` は目次として保ち、詳細な設計は `docs/` に置く
