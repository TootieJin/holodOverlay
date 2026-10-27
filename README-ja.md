<p align="center">
  <strong><a href="./README.md">English</a> | 日本語</strong>
</p>

<h1 align="center">
  <img src="./icons/holodOverlay-icon.png" width="120"><br>
  holodOverlay
</h1>

<p align="center">
  ホロドリ風動画作成補助ツール<br>
  A "hololive Dreams"-style video creation tool
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Windows-Only-0078D6">
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/TootieJin/holodOverlay" alt="License"></a> 
  <a href="https://github.com/TootieJin/holodOverlay/releases/"><img src="https://img.shields.io/github/downloads/TootieJin/holodOverlay/total" alt="Releases"></a> 
  <a href="https://github.com/TootieJin/holodOverlay/stargazers"><img src="https://img.shields.io/github/stars/TootieJin/holodOverlay?style=flat&amp;color=yellow" alt="Stargazers"></a>
</p>

<p align="center">
    <a href="https://ko-fi.com/TootieJin">
      <img width="400px" src="https://cdn.prod.website-files.com/5c14e387dab576fe667689cf/64f1a9ddd0246590df69ea0b_kofi_long_button_red%402x.png"/>
    </a>
</p>

<p align="center">
  <img src="./icons/holodOverlay-thumbnail.png">
</p>

# holodOverlay / ホロドリ風動画作成補助ツール

**[TootieJin](https://tootiejin.com)によるholodOverlay**は、ホロドリの創作譜面をホロドリ風の動画にするためのオープンソースのツールです。このツールは、公式ゲーム内のカスタムチャートと全く同じ見た目を再現しつつ、動画編集ソフト（AviUtl ExEdit2）を使用してUI（HUD）を「操作」する機能を提供します。

## 主な機能
このツールは、[pjsekai-overlay-APPEND](https://github.com/TootieJin/pjsekai-overlay-APPEND) という別のツールのホロドリ版であり、以下のような追加機能を備えています：
- 本家撮影と同じように見えるように要素を追加/調整
- AviUtl ExEdit2での完全な英語対応
- **動的レーン** (実験的機能)
- **カスタムライブスキン**（詳細は[利用方法](#利用方法)をご覧ください）
- 各種UIカスタマイズ：
  - **50種類以上のライブスキン**から選べます *（無料、月額290円を支払う必要はありません笑）*
  - ランクスコアは<u>S+9</u>を超えることもあり、他のランクにも「+」が追加される場合があります
  - 交換可能なライフ値
  - 交換可能な要素（スコア／コンボ／ポーズなど）
- **サーバー一覧**
  - **ホロドリ**
    - <u>公式サーバー：</u>
      - `holodori-`: hololive Dreams 非公式データベース ([sonolus.holodori.best](https://sonolus.holodori.best))
    - <u>カスタムサーバー：</u>
      - おっと、まだないですね！
  - **プロセカ**
    - <u>公式サーバー：</u>
      - `sekai-best-`: Sekai Viewer ([sonolus.sekai.best](https://sonolus.sekai.best))
      - `sss-`: SbugaのSonolusサーバー ([sonolus.sbuga.com](https://sonolus.sbuga.com))
    - <u>カスタムサーバー：</u>
      - `chcy-`: Chart Cyanvas
        - `0`: アーカイブ ([cc.milkbun.org](https://cc.milkbun.org))
        - `1`: 分岐サーバー ([chart-cyanvas.com](https://chart-cyanvas.com))
      - `ptlv-`: Potato Leaves ([ptlv.milkbun.org](https://ptlv.milkbun.org))
      - <del>`utsk-`: Untitled Sekai ([us.pim4n-net.com](https://us.pim4n-net.com))</del>
      - `UnCh-`: UntitledCharts ([untitledcharts.com](https://untitledcharts.com))
    - <u>ローカルサーバー：</u>
      - `sync-`: Local Server ([ScoreSync](https://github.com/Piliman22/ScoreSync) + [ScoreSync Modern](https://github.com/UntitledCharts/ScoreSync-Modern))
  - **その他**
    - `custom-[譜面ID]`: カスタムサーバー（ソースURL）

### • holodOverlayのプレビュー

https://github.com/user-attachments/assets/d83ed282-73db-417d-b25c-c5a701635bcc

### • AviUtl ExEdit2のプレビュー

<img src="./icons/aviutl2-preview.png" width="70%">

## 利用規約

1. **(必須)** 動画の説明文に、こちらのテキストをコピーしてください：

**EN**
```
holodOverlay - holodori-style video creation tool:
- Developed by TootieJin (https://tootiejin.com) 
   https://github.com/TootieJin/holodOverlay
```

**JP**
```
holodOverlay - ホロドリ風動画作成補助ツール：
- 作成：TootieJin (https://tootiejin.com)
   https://github.com/TootieJin/holodOverlay
```

2. 決して**悪意のある使用**をしないでください。（SNS上でデマを流すために使う等）
3. このツールを使ったことによるトラブルや不利益などが発生しても、作者は**一切の責任を負いません。**

## 必須事項
- [AviUtl ExEdit2 **（インストーラ版）**](http://spring-fragrance.mints.ne.jp/aviutl/) *（AviUtl ExEdit2をインストールしたら、`holodOverlay.exe` を起動して、必要なプラグインを自動的にダウンロードしてください。その後、AviUtl ExEdit2を再起動する必要があります。）* 
- **[PowerShell](https://github.com/PowerShell/PowerShell)**
- システムロケールは **必ず「日本語（日本）」** に設定してください （holodOverlayで設定することもできますが、[こちら](https://github.com/TootieJin/pjsekai-overlay-APPEND/wiki/Start-Here!#changing-system-locale)にアクセスして、ご自身で調整することも可能です。）
- AviUtl ExEdit2の基本的な知識
  - 日本語版の動画ガイドプレイリストは[こちら](https://www.youtube.com/playlist?list=PLjvcc2yii9PIr27ZsMzKymY2qST3GFF_M)（推奨）。
  - 「AviUtl2これだけはやっておけ集」をまとたNote記事は[こちら](https://note.com/hikami_meruru/n/na72411631b95)（2026年版、日本語・英語対応）。
  - 英語版の動画ガイドプレイリストは[こちら](https://www.youtube.com/playlist?list=PLKIaXWZCUwrvfr3WNG9P81aVvuCHRJXUZ)。
  - AviUtl ExEdit2の簡易説明（ドキュメント）は[こちら](https://docs.aviutl2.jp/usage)。

> [!CAUTION]
> [pjsekai-overlay-APPEND](https://github.com/TootieJin/pjsekai-overlay-APPEND)とは異なり、本ツールはAviUtl（旧バージョン）には対応していません。

## 利用方法

作業中。

> [!NOTE]
> 情報を更新する場合がありますので、Wikiを頻繁に確認してください。

## ライセンス
本ソフトウェアは、GNU Affero General Public License (GNU AGPL) の下で公開されています。

```
GNU Affero General Public License v3.0

holodOverlay - A "hololive Dreams"-style video creation tool
Copyright (C) 2026 TootieJin

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
```