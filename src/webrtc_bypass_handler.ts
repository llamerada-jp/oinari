import * as CL from "./crosslink"

const CL_PATH: string = "colonio/webrtc/";

export function NewWebrtcHandler(cl: CL.Crosslink, wi: WebrtcImplement): CL.MultiPlexer {
  let mpx = new CL.MultiPlexer();
  let cb = new CbWrap(cl);

  wi.setCb(cb);

  mpx.setObjHandlerFunc("contextInitialize", (_1: object, _2: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.contextInitialize();
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("contextAddIceServer", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.contextAddIceServer(param.iceServer);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkInitialize", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkInitialize(param.webrtcLink, param.isCreateDc);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkFinalize", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkFinalize(param.webrtcLink);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkDisconnect", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkDisconnect(param.webrtcLink);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkGetLocalSdp", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkGetLocalSdp(param.webrtcLink, param.isRemoteSdpSet);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkSend", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkSend(param.webrtcLink, new TextEncoder().encode(param.data));
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkSetRemoteSdp", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkSetRemoteSdp(param.webrtcLink, param.sdpStr, param.isOffer);
    writer.replySuccess({});
  });

  mpx.setObjHandlerFunc("linkUpdateIce", (param: any, _: Map<string, string>, writer: CL.ResponseObjWriter) => {
    wi.linkUpdateIce(param.webrtcLink, param.iceStr);
    writer.replySuccess({});
  });

  return mpx;
}

class CbWrap implements WebrtcLinkCb {
  cl: CL.Crosslink;

  constructor(cl: CL.Crosslink) {
    this.cl = cl;
  }

  onDcoError(webrtcLink: number, message: string): void {
    this.cl.call(CL_PATH + "onDcoError", {
      webrtcLink: webrtcLink,
      message: message,
    });
  }

  onDcoMessage(webrtcLink: number, message: ArrayBuffer): void {
    this.cl.call(CL_PATH + "onDcoMessage", {
      webrtcLink: webrtcLink,
      message: message,
    });
  }

  onDcoOpen(webrtcLink: number): void {
    this.cl.call(CL_PATH + "onDcoOpen", {
      webrtcLink: webrtcLink,
    });
  }

  onDcoClosing(webrtcLink: number): void {
    this.cl.call(CL_PATH + "onDcoClosing", {
      webrtcLink: webrtcLink,
    });
  }

  onDcoClose(webrtcLink: number): void {
    this.cl.call(CL_PATH + "onDcoClose", {
      webrtcLink: webrtcLink,
    });
  }

  onPcoError(webrtcLink: number, message: string): void {
    this.cl.call(CL_PATH + "onDcoMessage", {
      webrtcLink: webrtcLink,
      message: message,
    });
  }

  onPcoIceCandidate(webrtcLink: number, iceStr: string): void {
    this.cl.call(CL_PATH + "onPcoIceCandidate", {
      webrtcLink: webrtcLink,
      ice: iceStr,
    });
  }

  onPcoStateChange(webrtcLink: number, state: string): void {
    this.cl.call(CL_PATH + "onPcoStateChange", {
      webrtcLink: webrtcLink,
      state: state,
    });
  }

  onCsdSuccess(webrtcLink: number, sdpStr: string): void {
    this.cl.call(CL_PATH + "onCsdSuccess", {
      webrtcLink: webrtcLink,
      sdp: sdpStr,
    });
  }

  onCsdFailure(webrtcLink: number): void {
    this.cl.call(CL_PATH + "onCsdFailure", {
      webrtcLink: webrtcLink,
    });
  }
}