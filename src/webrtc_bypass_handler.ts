import * as CL from "./crosslink"

const CL_PATH: string = "colonio/webrtc/";

export function NewWebrtcHandler(cl: CL.Crosslink, wi: WebrtcImplement): CL.MultiPlexer {
  let mpx = new CL.MultiPlexer();
  let cb = new CbWrap(cl);

  wi.setCb(cb);

  mpx.setHandlerFunc("contextInitialize", (_1: object, _2: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.contextInitialize();
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("contextAddIceServer", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.contextAddIceServer(param.iceServer);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkInitialize", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkInitialize(param.webrtcLink, param.isCreateDc);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkFinalize", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkFinalize(param.webrtcLink);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkDisconnect", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkDisconnect(param.webrtcLink);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkGetLocalSdp", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkGetLocalSdp(param.webrtcLink, param.isRemoteSdpSet);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkSend", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkSend(param.webrtcLink, param.data);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkSetRemoteSdp", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
    wi.linkSetRemoteSdp(param.webrtcLink, param.sdpStr, param.isOffer);
    writer.replySuccess({});
  });

  mpx.setHandlerFunc("linkUpdateIce", (param: any, _: Map<string, string>, writer: CL.ResponseWriter) => {
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