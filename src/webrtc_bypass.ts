import * as CL from "./crosslink"

const CL_PATH: string = "colonio/webrtc/";

export class WebrtcBypass implements WebrtcImplement {
  cl: CL.Crosslink;
  cb: WebrtcLinkCb | undefined;

  constructor(cl: CL.Crosslink, mpxWorker: CL.MultiPlexer) {
    this.cl = cl;
    mpxWorker.setHandler("webrtc", this.makeWebrtcHandler());
  }

  setCb(cb: WebrtcLinkCb) {
    this.cb = cb;
  }

  contextInitialize() {
    this.cl.call(CL_PATH + "contextInitialize", {});
  }

  contextAddIceServer(iceServer: string) {
    this.cl.call(CL_PATH + "contextAddIceServer", {
      iceServer: iceServer,
    });
  }

  linkInitialize(webrtcLink: number, isCreateDc: boolean) {
    this.cl.call(CL_PATH + "linkInitialize", {
      webrtcLink: webrtcLink,
      isCreateDc: isCreateDc,
    });
  }

  linkFinalize(webrtcLink: number) {
    this.cl.call(CL_PATH + "linkFinalize", {
      webrtcLink: webrtcLink,
    });
  }

  linkDisconnect(webrtcLink: number) {
    this.cl.call(CL_PATH + "linkDisconnect", {
      webrtcLink: webrtcLink,
    });
  }

  linkGetLocalSdp(webrtcLink: number, isRemoteSdpSet: boolean) {
    this.cl.call(CL_PATH + "linkGetLocalSdp", {
      webrtcLink: webrtcLink,
      isRemoteSdpSet: isRemoteSdpSet,
    });
  }

  linkSend(webrtcLink: number, data: Uint8Array) {
    this.cl.call(CL_PATH + "linkSend", {
      webrtcLink: webrtcLink,
      data: new TextDecoder().decode(data),
    });
  }

  linkSetRemoteSdp(webrtcLink: number, sdpStr: string, isOffer: boolean) {
    this.cl.call(CL_PATH + "linkSetRemoteSdp", {
      webrtcLink: webrtcLink,
      sdpStr: sdpStr,
      isOffer: isOffer,
    });
  }

  linkUpdateIce(webrtcLink: number, iceStr: string) {
    this.cl.call(CL_PATH + "linkUpdateIce", {
      webrtcLink: webrtcLink,
      iceStr: iceStr,
    });
  }

  private makeWebrtcHandler(): CL.MultiPlexer {
    let mpx: CL.MultiPlexer = new CL.MultiPlexer();

    mpx.setObjHandlerFunc("onDcoError", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onDcoError(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onDcoMessage", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onDcoMessage(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onDcoOpen", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onDcoOpen(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onDcoClosing", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onDcoClosing(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onDcoClose", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onDcoClose(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onPcoError", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onPcoError(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onPcoIceCandidate", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onPcoIceCandidate(data.webrtcLink, data.ice);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onPcoStateChange", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onPcoStateChange(data.webrtcLink, data.state);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onCsdSuccess", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onCsdSuccess(data.webrtcLink, data.sdp);
      writer.replySuccess({});
    });

    mpx.setObjHandlerFunc("onCsdFailure", (data: any, _: Map<string, string>, writer: CL.ResponseObjWriter): void => {
      this.cb?.onCsdFailure(data.webrtcLink);
      writer.replySuccess({});
    });

    return mpx;
  }
};

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