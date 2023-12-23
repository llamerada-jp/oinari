---
title: Crosslink
type: docs
---

Oinari comprises several modules (browser's main worker as frontend, WebWorker, WebAssembly) and several languages (TypeScript, Go language), which call APIs implemented in those modules and work together. A simple implementation of APIs would require a function to bypass the API at the boundary of WebWorker and WebAssembly for each API that can be called from other modules. Which consumes a lot of time. Oinari's implementation provides a mechanism named Crosslink to reduce this hassle. 

## fundamental implementation

- src/crosslink.ts
- lib/crosslink/*.go

## handlers

- `application/api/core {containerID}` @lib/oinari/run.go via src/cri.ts
- `container` @src/container/manager.ts 
- `cri/container` @src/cri.ts
- `cri` @src/cri.ts
- `frontend/nodeReady` @src/index.ts
- `frontend/webrtc` @src/webrtc_bypass_handler.ts
- `node/api/core {containerID}` @node/apis/core/handler/handler.go
- `node/resource` @node/frontend/handler/resource.go
- `node/system` @node/frontend/handler/system.go
- `run` @src/controller/index.ts
- `webrtc` @src/controller/webrtc_bypass.ts