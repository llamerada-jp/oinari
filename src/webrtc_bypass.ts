import * as CL from "./crosslink"

const CL_PATH: string = "colonio/webrtc/";

interface WebrtcLinkCb {
  onDcoError(webrtcLink: number, message: string): void;
  onDcoMessage(webrtcLink: number, message: ArrayBuffer): void;
  onDcoOpen(webrtcLink: number): void;
  onDcoClosing(webrtcLink: number): void;
  onDcoClose(webrtcLink: number): void;
  onPcoError(webrtcLink: number, message: string): void;
  onPcoIceCandidate(webrtcLink: number, iceStr: string): void;
  onPcoStateChange(webrtcLink: number, state: string): void;
  onCsdSuccess(webrtcLink: number, sdpStr: string): void;
  onCsdFailure(webrtcLink: number): void;
};

interface WebrtcIf {
  setCb(cb: WebrtcLinkCb): void;
  contextInitialize(): void;
  contextAddIceServer(iceServer: string): void;
  linkInitialize(webrtcLink: number, isCreateDc: boolean): void;
  linkFinalize(webrtcLink: number): void;
  linkDisconnect(webrtcLink: number): void;
  linkGetLocalSdp(webrtcLink: number, isRemoteSdpSet: boolean): void;
  linkSend(webrtcLink: number, data: Uint8Array): void;
  linkSetRemoteSdp(webrtcLink: number, sdpStr: string, isOffer: boolean): void;
  linkUpdateIce(webrtcLink: number, iceStr: string): void;
};

class DefaultWebrtcImpl implements WebrtcIf {
  private webrtcContextPcConfig: any;
  private webrtcContextDcConfig: any;
  private availableWebrtcLinks: Map<number, { peer: RTCPeerConnection, dataChannel?: RTCDataChannel }>;
  private cb?: WebrtcLinkCb;

  constructor() {
    this.webrtcContextPcConfig = null;
    this.webrtcContextDcConfig = null;
    this.availableWebrtcLinks = new Map();
  }

  setCb(cb: WebrtcLinkCb) {
    this.cb = cb;
  }

  contextInitialize() {
    this.webrtcContextPcConfig = {
      iceServers: []
    };

    this.webrtcContextDcConfig = {
      orderd: true,
      // maxRetransmitTime: 3000,
      maxPacketLifeTime: 3000
    };
  }

  contextAddIceServer(iceServer: string) {
    this.webrtcContextPcConfig.iceServers.push(
      JSON.parse(iceServer)
    );
  }

  linkInitialize(webrtcLink: number, isCreateDc: boolean) {
    let setEvent = (dataChannel: RTCDataChannel) => {
      dataChannel.addEventListener("error", (event: Event) => {
        console.log("rtc data error", webrtcLink, event);
        if (this.availableWebrtcLinks.has(webrtcLink)) {
          this.cb?.onDcoError(webrtcLink, (event as any).error.message)
        }
      });

      dataChannel.addEventListener("message", (event) => {
        if (this.availableWebrtcLinks.has(webrtcLink)) {
          if (event.data instanceof ArrayBuffer) {
            // console.log("rtc data recv", webrtcLink, dumpPacket(new TextDecoder("utf-8").decode(event.data)));
            this.cb?.onDcoMessage(webrtcLink, event.data);
          } else if (event.data instanceof Blob) {
            let reader = new FileReader();
            reader.onload = () => {
              // console.log("rtc data recv", webrtcLink, dumpPacket(new TextDecoder("utf-8").decode(reader.result)));
              this.cb?.onDcoMessage(webrtcLink, reader.result as ArrayBuffer);
            };
            reader.readAsArrayBuffer(event.data);
          } else {
            console.error("Unsupported type of message.", JSON.stringify(event.data));
          }
        }
      });

      dataChannel.addEventListener("open", () => {
        console.log("rtc data open", webrtcLink);
        if (this.availableWebrtcLinks.has(webrtcLink)) {
          this.cb?.onDcoOpen(webrtcLink);
        }
      });

      dataChannel.addEventListener("closing", () => {
        console.log("rtc data closing", webrtcLink);
        if (this.availableWebrtcLinks.has(webrtcLink)) {
          this.cb?.onDcoClosing(webrtcLink);
        }
      });

      dataChannel.addEventListener("close", () => {
        console.log("rtc data close", webrtcLink);
        if (this.availableWebrtcLinks.has(webrtcLink)) {
          this.cb?.onDcoClose(webrtcLink);
        }
      });
    };

    let peer: RTCPeerConnection;
    try {
      peer = new RTCPeerConnection(this.webrtcContextPcConfig);

    } catch (e) {
      console.error(e);
      return;
    }

    let dataChannel: RTCDataChannel | undefined;
    if (isCreateDc) {
      dataChannel = peer.createDataChannel("data_channel",
        this.webrtcContextDcConfig);
      setEvent(dataChannel);
    }

    this.availableWebrtcLinks.set(webrtcLink, {
      peer: peer,
      dataChannel: dataChannel,
    });

    peer.onicecandidate = (event) => {
      console.log("rtc on ice candidate", webrtcLink);
      if (!this.availableWebrtcLinks.has(webrtcLink)) {
        return;
      }

      let ice = (event.candidate) ? JSON.stringify(event.candidate) : "";
      this.cb?.onPcoIceCandidate(webrtcLink, ice);
    };

    peer.ondatachannel = (event) => {
      console.log("rtc peer datachannel", webrtcLink);
      let link = this.availableWebrtcLinks.get(webrtcLink);
      if (link === undefined) {
        return;
      }

      if (link?.dataChannel !== null) {
        this.cb?.onPcoError(webrtcLink, "duplicate data channel.");
      }

      link.dataChannel = event.channel;
      setEvent(event.channel);
    };

    peer.oniceconnectionstatechange = (event) => {
      console.log("rtc peer state", webrtcLink, peer.iceConnectionState);
      let link = this.availableWebrtcLinks.get(webrtcLink);
      if (link === undefined) {
        return;
      }
      this.cb?.onPcoStateChange(webrtcLink, link.peer.iceConnectionState);
    };
  }

  linkFinalize(webrtcLink: number) {
    console.assert(this.availableWebrtcLinks.has(webrtcLink));
    this.availableWebrtcLinks.delete(webrtcLink);
  }

  linkDisconnect(webrtcLink: number) {
    console.assert(this.availableWebrtcLinks.has(webrtcLink));
    let link = this.availableWebrtcLinks.get(webrtcLink);
    if (link === undefined) {
      return;
    }

    if (link.dataChannel !== undefined) {
      link.dataChannel.close();
    }

    link.peer.close();
  }

  linkGetLocalSdp(webrtcLink: number, isRemoteSdpSet: boolean) {
    console.log("rtc getLocalSdp", webrtcLink);
    console.assert(this.availableWebrtcLinks.has(webrtcLink));
    let link = this.availableWebrtcLinks.get(webrtcLink);
    if (link === undefined) {
      return;
    }

    try {
      let peer = link.peer;
      let description: RTCSessionDescriptionInit;

      if (isRemoteSdpSet) {
        peer?.createAnswer().then((sessionDescription) => {
          description = sessionDescription;
          return peer?.setLocalDescription(sessionDescription);

        }).then(() => {
          console.log("rtc createAnswer", webrtcLink);
          let sdp = description.sdp;
          if (sdp === undefined) {
            sdp = "";
          }
          this.cb?.onCsdSuccess(webrtcLink, sdp);

        }).catch((e: any) => {
          console.log("rtc createAnswer error", webrtcLink, e);
          this.cb?.onCsdFailure(webrtcLink);
        });

      } else {
        peer?.createOffer().then((sessionDescription) => {
          description = sessionDescription;
          return peer?.setLocalDescription(sessionDescription);

        }).then(() => {
          console.log("rtc createOffer", webrtcLink);
          let sdp = description.sdp;
          if (sdp === undefined) {
            sdp = "";
          }
          this.cb?.onCsdSuccess(webrtcLink, sdp);

        }).catch((e: any) => {
          console.error(e);
          this.cb?.onCsdFailure(webrtcLink);
        });
      }

    } catch (e) {
      console.error(e);
    }
  }

  linkSend(webrtcLink: number, data: Uint8Array) {
    try {
      let link = this.availableWebrtcLinks.get(webrtcLink);
      link?.dataChannel?.send(data);
    } catch (e) {
      console.error(e);
    }
  }

  linkSetRemoteSdp(webrtcLink: number, sdpStr: string, isOffer: boolean) {
    try {
      let link = this.availableWebrtcLinks.get(webrtcLink);
      let peer = link?.peer;
      let sdp = {
        type: (isOffer ? "offer" : "answer"),
        sdp: sdpStr
      };
      peer?.setRemoteDescription(new RTCSessionDescription(sdp as RTCSessionDescriptionInit));

    } catch (e) {
      console.error(e);
    }
  }

  linkUpdateIce(webrtcLink: number, iceStr: string) {
    try {
      let link = this.availableWebrtcLinks.get(webrtcLink);
      let peer = link?.peer;
      let ice = JSON.parse(iceStr);

      peer?.addIceCandidate(new RTCIceCandidate(ice));

    } catch (e) {
      console.error(e);
    }
  }
}

export class WebrtcBypass implements WebrtcIf {
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

export function NewWebrtcHandler(cl: CL.Crosslink): CL.MultiPlexer {
  let mpx = new CL.MultiPlexer();
  let cb = new CbWrap(cl);
  let wi = new DefaultWebrtcImpl();

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