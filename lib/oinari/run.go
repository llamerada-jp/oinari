package oinari

import (
	"fmt"
	"log"

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

	err := o.run()
	if err != nil {
		return fmt.Errorf("error on Run of oinari: %w", err)
	}
	return nil
}

func (o *oinari) run() error {
	o.initCrosslink()

	errCh := make(chan error)
	if err := o.initCoreHandler(errCh); err != nil {
		return fmt.Errorf("failed on initCoreHandler: %w", err)
	}

	if err := initWriter(o.cl, NodeCrosslinkPath); err != nil {
		return fmt.Errorf("failed on initWriter: %w", err)
	}

	if err := o.ready(); err != nil {
		return fmt.Errorf("failed on ready: %w", err)
	}

	err, ok := <-errCh
	if ok {
		close(errCh)
		return err
	}
	return nil
}

func (o *oinari) initCrosslink() {
	o.rootMpx = crosslink.NewMultiPlexer()
	o.cl = crosslink.NewCrosslink("crosslink", o.rootMpx)
}

func (o *oinari) initCoreHandler(errCh chan error) error {
	appMpx := crosslink.NewMultiPlexer()
	o.rootMpx.SetHandler("application", appMpx)
	apiMpx := crosslink.NewMultiPlexer()
	appMpx.SetHandler("api", apiMpx)
	coreAPIMpx := crosslink.NewMultiPlexer()
	apiMpx.SetHandler("core", coreAPIMpx)

	coreAPIMpx.SetHandler("setup", crosslink.NewFuncHandler(func(req *app.SetupRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		err := o.app.Setup(req.IsInitialize, req.Record)
		if err != nil {
			writer.ReplyError("setup had an error")
			errCh <- fmt.Errorf("catch an error on `Setup` method: %s", err)
		} else {
			writer.ReplySuccess(app.SetupResponse{})
		}
	}))

	coreAPIMpx.SetHandler("marshal", crosslink.NewFuncHandler(func(req *app.MarshalRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		record, err := o.app.Marshal()
		if err != nil {
			writer.ReplyError("marshal had an error")
			errCh <- fmt.Errorf("catch an error on `Marshal` method: %s", err)
		} else {
			writer.ReplySuccess(app.MarshalResponse{
				Record: record,
			})
		}
	}))

	coreAPIMpx.SetHandler("teardown", crosslink.NewFuncHandler(func(req *app.TeardownRequest, tags map[string]string, writer crosslink.ResponseWriter) {
		record, err := o.app.Teardown(req.IsFinalize)
		// ignore record if finalize
		if req.IsFinalize {
			record = nil
		}
		if err != nil {
			writer.ReplyError("teardown had an error")
			errCh <- fmt.Errorf("catch an error on `Teardown` method: %s", err)
		} else {
			writer.ReplySuccess(app.TeardownResponse{
				Record: record,
			})
			close(errCh)
		}
	}))

	return nil
}

func (o *oinari) ready() error {
	o.cl.Call(NodeCrosslinkPath+"/ready", app.ReadyRequest{}, nil, func(b []byte, err error) {
		if err != nil {
			log.Fatalf("failed to ready core api of oinari, do you have a runtime set?: %s", err.Error())
		}
	})
	return nil
}
