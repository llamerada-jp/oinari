package oinari

import (
	app "github.com/llamerada-jp/oinari/api/app/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

type oinari struct {
	// application
	app Application

	// crosslink
	cl      crosslink.Crosslink
	rootMpx crosslink.MultiPlexer
}

func Run(app Application) error {
	o := &oinari{
		app: app,
	}

	return o.run()
}

func (o *oinari) run() error {
	if err := o.initCrosslink(); err != nil {
		return err
	}

	errCh := make(chan error)
	if err := o.initCoreHandler(errCh); err != nil {
		return err
	}

	if err := initWriter(o.cl, CrosslinkPath); err != nil {
		return err
	}

	if err := o.ready(); err != nil {
		return err
	}

	err, ok := <-errCh
	if ok {
		return err
	}
	return nil
}

func (o *oinari) initCrosslink() error {
	o.rootMpx = crosslink.NewMultiPlexer()
	o.cl = crosslink.NewCrosslink("crosslink", o.rootMpx)
	return nil
}

func (o *oinari) initCoreHandler(errCh chan error) error {
	appMpx := crosslink.NewMultiPlexer()
	o.rootMpx.SetHandler(CrosslinkPath, appMpx)

	appMpx.SetHandler("setup", crosslink.NewFuncHandler(func(req *app.SetupRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := o.app.Setup(req.FirstInPod)
		if err != nil {
			writer.ReplyError("setup had an error")
			errCh <- err
		} else {
			writer.ReplySuccess(app.SetupResponse{})
		}
	}))

	appMpx.SetHandler("dump", crosslink.NewFuncHandler(func(req *app.DumpRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		data, err := o.app.Dump()
		if err != nil {
			writer.ReplyError("dump had an error")
			errCh <- err
		} else {
			writer.ReplySuccess(app.DumpResponse{
				DumpData: data,
			})
		}
	}))

	appMpx.SetHandler("restore", crosslink.NewFuncHandler(func(req *app.RestoreRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := o.app.Restore(req.DumpData)
		if err != nil {
			writer.ReplyError("restore had an error")
			errCh <- err
		} else {
			writer.ReplySuccess(app.RestoreResponse{})
		}
	}))

	appMpx.SetHandler("teardown", crosslink.NewFuncHandler(func(req *app.TeardownRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := o.app.Teardown(req.LastInPod)
		if err != nil {
			writer.ReplyError("teardown had an error")
			errCh <- err
		} else {
			writer.ReplySuccess(app.TeardownResponse{})
			close(errCh)
		}
	}))

	return nil
}

func (o *oinari) ready() error {
	o.cl.Call(CrosslinkPath+"/ready", app.ReadyRequest{}, nil, func(b []byte, err error) {
		// `ready` request only tell the status to the manager of this node, do nothing
	})
	return nil
}
