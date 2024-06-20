---
title: Welcome
type: docs
---

# Oinari(オイナリ)

## Overview
Oinari is a distributed operating system targeting XR (particularly MR) and IoT. Despite being referred to as an operating system, it does not require installation on specific devices; instead, it operates as a node in the system accessible through [www.oinari.io](https://www.oinari.io).

Applications running on Oinari run on the Nodes that are part of the system, i.e., the user's device, rather than in close geographic proximity to the terminal, as in Fog computing or Edge computing. The output of the application is also reflected in other geographically nearby Nodes, specifically, objects are rendered in XR space. Therefore, Oinari's space is not limited by server resources and is scalable. Also, because programs run in geographic proximity, it is expected to improve the operational responsiveness of the application.

Applications on Oinari can continue to run on other Nodes even if some Nodes are down. Applications developed using the specialized API will have migration capabilities and will continue to run regardless of whether the Node is alive or dead. Applications are WebAssembly and language-independent (although currently only APIs for the Go language is only available).

For a more detailed description and technical background, please refer to the [concept]({{< relref "posts/concept.md" >}}) page.

## Getting Started

You can try Oinari working as a web service by accessing [www.oinari.io](https://www.oinari.io). Please refer to the [User tutorial]({{< relref "posts/user_tutorial.md" >}}) for login and simple usage. To develop and run an application using Oinari's features, please refer to the [Developer tutorial]({{< relref "posts/developer_tutorial.md" >}}).

## License

Oinari is licensed under the Apache License 2.0. Check the [LICENSE](https://github.com/llamerada-jp/oinari/blob/main/LICENSE) for details.