package oinari

import (
	"encoding/json"
	"io"
	"log"

	app "github.com/llamerada-jp/oinari/api/app/core"
	"github.com/llamerada-jp/oinari/lib/crosslink"
)

var Writer io.Writer

type writer struct {
	cl   crosslink.Crosslink
	path string
}

func initWriter(cl crosslink.Crosslink, path string) error {
	Writer = &writer{
		cl:   cl,
		path: path,
	}

	return nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	type writeResponse struct {
		len int
		err error
	}
	resCh := make(chan writeResponse)

	w.cl.Call(w.path+"/output", app.OutputRequest{
		Payload: p,
	}, nil, func(b []byte, err error) {
		if err != nil {
			resCh <- writeResponse{
				len: 0,
				err: err,
			}
		}

		var res app.OutputResponse
		err = json.Unmarshal(b, &res)
		if err != nil {
			log.Fatalf("unmarshal response of output failed on oinari api: %s", err.Error())
		}

		resCh <- writeResponse{
			len: res.Length,
			err: nil,
		}
	})

	res := <-resCh
	return res.len, res.err
}
