---
title: ようこそ
type: docs
---

# Oinari(オイナリ)

## 概要

Oinari は XR (特に MR) や IoT をターゲットにした分散オペレーティングシステムです。
オペレーティングシステムといっても、具体的なデバイスにインストールする必要はなく、[www.oinari.io](https://www.oinari.io) へアクセスすることでシステムのノードとして動きます。

Oinari 上で動くアプリケーションは、Fog computing や Edge computing のように端末の近いところ、というよりシステムを構成する Node、つまり利用者のデバイス上で動きます。アプリケーションの実行結果は地理的に近いほかの Node にも結果が反映、具体的には XR 空間にオブジェクトがレンダリングされます。そのため、Oinari の空間はサーバリソースに制限を受けずに拡張可能です。またプログラムが地理的に近いところで動くため、アプリケーションの操作レスポンスの向上を期待しています。

Oinari 上のアプリケーションは特定の Node が停止しても他の Node で動き続けることができます。専用の API に沿って作られたアプリケーションにはマイグレーション機能が付与され、Node の生死に関わらず実行され続けます。アプリケーションは WebAssembly であり、プログラミング言語に依存しません(ただし、現在提供しているのは Go 言語向けの API だけです)。

より詳しい説明、技術的背景は[コンセプト]({{< relref "posts/concept.md" >}})のページを参照してください。

## 使い方

Web サービスとして稼働している Oinari は [www.oinari.io](https://www.oinari.io) へアクセスして試すことができます。ログインから簡単な使い方については[利用チュートリアル]({{< relref "posts/user_tutorial.md" >}})を参照してください。
Oinari の機能を利用したアプリケーションを開発、実際に動かすには[開発チュートリアル]({{< relref "posts/developer_tutorial.md" >}})を参照してください。

## ライセンス

Oinari は Apache License 2.0 で開発されています。
詳細は添付の [LICENSE](https://github.com/llamerada-jp/oinari/blob/main/LICENSE) を参照してください。