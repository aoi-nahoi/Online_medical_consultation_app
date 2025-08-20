# オンライン診療サポートアプリ セットアップガイド

## 概要
このプロジェクトは、患者と医師をつなぐオンライン診療プラットフォームのMVPプロトタイプです。

## 技術スタック
- **フロントエンド**: Next.js 14 + TypeScript + Tailwind CSS
- **バックエンド**: Go + Gin + GORM
- **データベース**: PostgreSQL
- **リアルタイム通信**: WebSocket + WebRTC
- **認証**: JWT（HS256）

## 前提条件
以下のソフトウェアがインストールされている必要があります：

- Docker & Docker Compose
- Node.js 18+
- Go 1.22+

## セットアップ手順

### 1. リポジトリのクローン
```bash
git clone <repository-url>
cd Online_medical_consultation_app
```

### 2. 環境変数の設定
```bash
# .envファイルを作成
cp .env.example .env

# 必要に応じて値を編集
DATABASE_URL=postgres://telemed:telemed123@localhost:5432/telemed?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

### 3. データベースの起動
```bash
# PostgreSQLとRedisを起動
docker-compose up -d postgres redis

# データベースの準備が完了するまで待機（約30秒）
```

### 4. バックエンドのセットアップ
```bash
cd backend

# Goの依存関係をインストール
go mod tidy

# アプリケーションを起動
go run cmd/api/main.go
```

バックエンドが正常に起動すると、以下のメッセージが表示されます：
```
Database connected successfully
Running database migrations...
Database migrations completed successfully
Server starting on port 8080
```

### 5. フロントエンドのセットアップ
```bash
# 新しいターミナルで
cd ..  # プロジェクトルートに戻る

# Node.jsの依存関係をインストール
npm install

# 開発サーバーを起動
npm run dev
```

フロントエンドが正常に起動すると、以下のメッセージが表示されます：
```
- Local:        http://localhost:3000
- Network:      http://192.168.x.x:3000
```

## アクセス方法

### アプリケーション
- **フロントエンド**: http://localhost:3000
- **バックエンドAPI**: http://localhost:8080
- **データベース**: localhost:5432

### デモアカウント
以下のアカウントでログインできます：

**医師アカウント**
- メール: doctor1@example.com
- パスワード: pass

**患者アカウント**
- メール: patient1@example.com
- パスワード: pass

## 主要機能

### 患者機能
- 予約の作成・管理
- チャット機能
- ビデオ通話参加
- 処方内容の確認

### 医師機能
- 診療枠の管理
- 予約の承認・却下
- チャット機能
- ビデオ通話開始
- 処方の記録

## 開発ガイド

### バックエンド開発
```bash
cd backend

# コードの実行
go run cmd/api/main.go

# テストの実行
go test ./...

# コードの整形
go fmt ./...
```

### フロントエンド開発
```bash
# 開発サーバーの起動
npm run dev

# ビルド
npm run build

# 型チェック
npm run type-check

# リント
npm run lint
```

### データベース操作
```bash
# PostgreSQLに接続
docker exec -it telemed_postgres psql -U telemed -d telemed

# テーブルの確認
\dt

# データの確認
SELECT * FROM users;
```

## トラブルシューティング

### よくある問題

#### 1. データベース接続エラー
```bash
# コンテナの状態確認
docker-compose ps

# コンテナの再起動
docker-compose restart postgres

# ログの確認
docker-compose logs postgres
```

#### 2. ポートの競合
```bash
# 使用中のポートを確認
netstat -tulpn | grep :8080
netstat -tulpn | grep :3000

# プロセスの終了
kill -9 <PID>
```

#### 3. Goモジュールの問題
```bash
cd backend
go clean -modcache
go mod tidy
```

#### 4. Node.jsの依存関係の問題
```bash
rm -rf node_modules package-lock.json
npm install
```

## アーキテクチャ

### ディレクトリ構造
```
Online_medical_consultation_app/
├── backend/                 # Goバックエンド
│   ├── cmd/api/            # メインアプリケーション
│   ├── internal/           # 内部パッケージ
│   │   ├── config/         # 設定管理
│   │   ├── database/       # データベース接続・マイグレーション
│   │   ├── handlers/       # HTTPハンドラー
│   │   ├── middleware/     # ミドルウェア
│   │   ├── models/         # データモデル
│   │   ├── repositories/   # データアクセス層
│   │   ├── services/       # ビジネスロジック
│   │   └── websocket/      # WebSocket処理
│   └── go.mod              # Goモジュール定義
├── src/                    # Next.jsフロントエンド
│   ├── app/               # App Router
│   │   ├── (auth)/        # 認証関連ページ
│   │   ├── patient/       # 患者用ページ
│   │   ├── doctor/        # 医師用ページ
│   │   └── globals.css    # グローバルスタイル
│   └── components/        # 再利用可能なコンポーネント
├── docker-compose.yml      # Docker Compose設定
├── package.json            # Node.js依存関係
└── README.md              # プロジェクト概要
```

### API設計
- **ベースURL**: `/api/v1`
- **認証**: JWT Bearer Token
- **レスポンス形式**: JSON
- **エラーハンドリング**: HTTPステータスコード + エラーメッセージ

### データベース設計
- **RDBMS**: PostgreSQL
- **ORM**: GORM
- **マイグレーション**: 自動マイグレーション
- **シードデータ**: 初期データの自動生成

## デプロイメント

### 本番環境での注意点
1. **環境変数**: 本番用の値に変更
2. **JWT_SECRET**: 強力な秘密鍵を使用
3. **データベース**: 本番用PostgreSQLインスタンス
4. **セキュリティ**: HTTPS、ファイアウォール設定
5. **ログ**: 本番用ログ設定

### Docker本番ビルド
```bash
# フロントエンドのビルド
npm run build

# Dockerイメージのビルド
docker build -t telemed-app .

# 本番環境での実行
docker run -p 80:3000 telemed-app
```

## 貢献方法

### 開発フロー
1. フィーチャーブランチの作成
2. 機能の実装
3. テストの作成・実行
4. プルリクエストの作成
5. コードレビュー
6. マージ

### コーディング規約
- **Go**: `gofmt`、`golint`に準拠
- **TypeScript**: ESLint、Prettierに準拠
- **コミットメッセージ**: 日本語で分かりやすく

## ライセンス
このプロジェクトはMITライセンスの下で公開されています。

## サポート
問題が発生した場合は、以下の手順で対応してください：

1. このドキュメントの確認
2. GitHub Issuesでの報告
3. 開発チームへの連絡

---

**注意**: このアプリケーションは医療目的での使用は想定されておらず、デモンストレーション目的でのみ使用してください。
