---
title: Crosslink
---

Oinari comprises several modules (browser's main worker as frontend, WebWorker, WebAssembly) and several languages (TypeScript, Go language), which call APIs implemented in those modules and work together. A simple implementation of APIs would require a function to bypass the API at the boundary of WebWorker and WebAssembly for each API that can be called from other modules. Which consumes a lot of time. Oinari's implementation provides a mechanism named Crosslink to reduce this hassle. 

## fundamental implementation

- src/crosslink.ts
- lib/crosslink/*.go

## handler implementations

- `/application {container}` @src/cri.ts, lib/oinari/run.go
- `/colonio/webrtc` @src/webrtc_bypass_handler.ts
- `/container/term` @src/container/manager.ts 
- `/container` @src/cri.ts
- `/cri` @src/cri.ts
- `/frontend/onInitComplete` @src/index.ts
- `/node {sandbox, container}` @src/cri.ts 
- `/resource` @node/frontend/handler/resource.go
- `/run` @src/controller/index.ts
- `/system` @node/frontend/handler/system.go