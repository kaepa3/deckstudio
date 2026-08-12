# 🎛️ DeckStudio

スマホやタブレットのブラウザを、PC用の**Webベース左手デバイス（タッチコントロールパネル）**として使えるようにするGo言語製の軽量ローカルサーバー＆Web UIシステムです。

Stream DeckやTouch Portalのように、ボタン一発でキーショートカット実行、音量調整、アプリ別プロファイルの自動切り替えが可能です。

---

## 🚀 使い方

### 1. サーバーの起動

ターミナルで本フォルダに移動し、`task` コマンドで起動します。

```bash
# ビルドしてサーバーを起動
task run
```

> ※ `task` がない場合は `go run .` でも起動できます。

起動するとターミナルにアクセス用URLが表示されます：

```text
==================================================
 🎛️  DeckStudio Server is running!
 🌐  Local Access:  http://localhost:8080
 📱  Mobile Access: http://192.168.x.x:8080
==================================================
```

### 2. スマホ・タブレットからアクセス

1. スマホをPCと同じWi-Fi（ローカルネットワーク）に接続します。
2. スマホのブラウザで **Mobile Access** のURL（例: `http://192.168.1.10:8080`）を開きます。
3. スマホ画面のボタンをタップすれば、PC側でキー入力が実行されます！
   - 💡 **ヒント**: スマホの「ホーム画面に追加」を行うと、全画面アプリとして使えます。

---

## ✨ 主な機能

- ⚡ **超低遅延操作**: Go言語 ＋ WebSocket通信による高速レスポンス。
- 🤖 **アクティブウィンドウ自動追従 (Auto Sync)**: PCで現在アクティブなアプリ（VS Code、ブラウザ、Spotifyなど）を検知し、スマホ画面のプロファイルを自動切り替え。
- 🎛️ **グローバル ＋ アプリ専用エリア**: 上部に常に使える共通キー（Undo, 音量など）、下部にアクティブアプリ専用キーを配置。
- ⚙️ **YAMLで簡単カスタマイズ**: `config.yaml` を書き換えるだけでボタンやプロファイルを自由自在に編集可能。
- 📳 **触覚フィードバック**: ボタンタップ時にスマホのバイブレーション（`navigator.vibrate`）が動作。

---

## ⚙️ 設定ファイルの検索優先順位 (`.deckstudio.yaml`)

サーバー起動時、設定ファイル **`.deckstudio.yaml`** は以下の優先順位で自動探索・読み込まれます：

1. **第1候補**: 実行時の直下 (`./.deckstudio.yaml` または `deckstudio.exe` と同じフォルダー)
2. **第2候補**: ユーザーのホームディレクトリ (`C:\Users\<ユーザー名>\.deckstudio.yaml`)
3. **第3候補**: ユーザーのホームディレクトリ専用フォルダ (`C:\Users\<ユーザー名>\.deckstudio\.deckstudio.yaml`)

※読み込まれたファイルパスがサーバー起動時のログに表示されます (`📄 Loaded configuration from: ...`)。

```yaml
global_buttons:
  - id: global_undo
    label: "Undo"
    icon: "rotate-ccw"       # Lucide Icon名 (https://lucide.dev)
    color: "#ff5f56"          # ボタンカラー
    type: "shortcut"
    keys: ["ctrl", "z"]

profiles:
  - id: vscode
    name: "VS Code"
    app_names:              # 自動切替の対象プロセス名
      - "code.exe"
    icon: "code"
    buttons:
      - id: vsc_format
        label: "Format Code"
        icon: "file-code-2"
        color: "#007acc"
        type: "shortcut"
        keys: ["shift", "alt", "f"]
```

---

## 🛠️ コマンド一覧 (`Taskfile`)

- `task`: バイナリ (`deckstudio.exe`) をビルド
- `task run`: ビルド ＋ **フォアグラウンド（対話モード）**でサーバー起動
- `task start`: ビルド ＋ **バックグラウンド（ウィンドウ非表示）**で常駐起動 🚀
- `task status`: バックグラウンドプロセスの実行状態を確認 🔍
- `task stop`: バックグラウンドプロセスを停止 🛑
- `task clean`: 生成されたバイナリを削除
- `task --list`: 利用可能なタスク一覧を表示
