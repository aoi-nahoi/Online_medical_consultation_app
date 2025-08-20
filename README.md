# オンライン診療サポートアプリ（プロトタイプMVP）

## 概要
患者が医師にオンラインで予約・チャット・ビデオ相談でき、医師が予約を管理・応答・簡易処方登録できるWebアプリケーションです。

## 技術スタック
- **フロントエンド**: Next.js 14 + TypeScript + Tailwind CSS
- **バックエンド**: Go + Gin + GORM
- **データベース**: PostgreSQL
- **リアルタイム通信**: WebSocket + WebRTC
- **認証**: JWT（HS256）
- **ファイルストレージ**: ローカルファイルシステム（開発時）

## 機能
- 患者・医師の認証・認可
- 診療枠の公開・予約
- チャット機能（画像・PDF添付対応）
- ビデオ通話（WebRTC）
- 処方記録・管理
- 監査ログ

## セットアップ

### 前提条件
- Docker & Docker Compose
- Node.js 18+
- Go 1.22+

### 起動方法
```bash
# 1. 依存関係のインストール
npm install
cd backend && go mod tidy

# 2. 環境変数の設定
cp .env.example .env

# 3. アプリケーションの起動
docker-compose up -d

# 4. フロントエンドの起動
npm run dev

# 5. バックエンドの起動
cd backend && go run cmd/api/main.go
```

## アクセス
- フロントエンド: http://localhost:3000
- バックエンドAPI: http://localhost:8080
- データベース: localhost:5432

## デモアカウント
- 医師: doctor1@example.com / pass
- 患者: patient1@example.com / pass

## 開発ガイド
詳細な開発ガイドは各ディレクトリ内のREADMEを参照してください。
