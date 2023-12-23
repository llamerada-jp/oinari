import * as CL from "../crosslink"

export class WebrtcBypass implements WebrtcImplement {
  cl: CL.Crosslink;
  cb: WebrtcLinkCb | undefined;
  clPath: string;

  constructor(cl: CL.Crosslink, mpxWorker: CL.MultiPlexer, clPath:string) {
    this.cl = cl;
    this.clPath = clPath + "/";
    mpxWorker.setHandler("webrtc", this.makeWebrtcHandler());
  }

  setCb(cb: WebrtcLinkCb) {
    this.cb = cb;
  }

  contextInitialize() {
    this.cl.call(this.clPath + "contextInitialize", {});
  }

  contextAddIceServer(iceServer: string) {
    this.cl.call(this.clPath + "contextAddIceServer", {
      iceServer: iceServer,
    });
  }

  linkInitialize(webrtcLink: number, isCreateDc: boolean) {
    this.cl.call(this.clPath + "linkInitialize", {
      webrtcLink: webrtcLink,
      isCreateDc: isCreateDc,
    });
  }

  linkFinalize(webrtcLink: number) {
    this.cl.call(this.clPath + "linkFinalize", {
      webrtcLink: webrtcLink,
    });
  }

  linkDisconnect(webrtcLink: number) {
    this.cl.call(this.clPath + "linkDisconnect", {
      webrtcLink: webrtcLink,
    });
  }

  linkGetLocalSdp(webrtcLink: number, isRemoteSdpSet: boolean) {
    this.cl.call(this.clPath + "linkGetLocalSdp", {
      webrtcLink: webrtcLink,
      isRemoteSdpSet: isRemoteSdpSet,
    });
  }

  linkSend(webrtcLink: number, data: Uint8Array) {
    this.cl.call(this.clPath + "linkSend", {
      webrtcLink: webrtcLink,
      data: data.slice(),
    });
  }

  linkSetRemoteSdp(webrtcLink: number, sdpStr: string, isOffer: boolean) {
    this.cl.call(this.clPath + "linkSetRemoteSdp", {
      webrtcLink: webrtcLink,
      sdpStr: sdpStr,
      isOffer: isOffer,
    });
  }

  linkUpdateIce(webrtcLink: number, iceStr: string) {
    this.cl.call(this.clPath + "linkUpdateIce", {
      webrtcLink: webrtcLink,
      iceStr: iceStr,
    });
  }

  private makeWebrtcHandler(): CL.MultiPlexer {
    let mpx: CL.MultiPlexer = new CL.MultiPlexer();

    mpx.setHandlerFunc("onDcoError", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onDcoError(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onDcoMessage", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onDcoMessage(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onDcoOpen", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onDcoOpen(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onDcoClosing", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onDcoClosing(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onDcoClose", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onDcoClose(data.webrtcLink);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onPcoError", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onPcoError(data.webrtcLink, data.message);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onPcoIceCandidate", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onPcoIceCandidate(data.webrtcLink, data.ice);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onPcoStateChange", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onPcoStateChange(data.webrtcLink, data.state);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onCsdSuccess", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onCsdSuccess(data.webrtcLink, data.sdp);
      writer.replySuccess({});
    });

    mpx.setHandlerFunc("onCsdFailure", (data: any, _: Map<string, string>, writer: CL.ResponseWriter): void => {
      this.cb?.onCsdFailure(data.webrtcLink);
      writer.replySuccess({});
    });

    return mpx;
  }
};
