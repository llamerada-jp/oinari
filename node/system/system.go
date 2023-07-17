package system

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/llamerada-jp/colonio/go/colonio"
	"github.com/llamerada-jp/oinari/node/frontend"
)

type EventHandler interface {
	OnConnect() error
}

type System interface {
	Start(ctx context.Context) error
	Connect(url, account, token string) error

	GetAccount() string
	GetNode() string
	SetPosition(latitude, longitude float64) error
}

type systemImpl struct {
	colonio colonio.Colonio
	evh     EventHandler
	fd      frontend.Driver
	account string
}

func init() {
	rand.Seed(time.Now().UnixMicro())
}

func NewSystem(col colonio.Colonio, evh EventHandler, fd frontend.Driver) System {
	return &systemImpl{
		colonio: col,
		evh:     evh,
		fd:      fd,
	}
}

func (sys *systemImpl) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (sys *systemImpl) GetAccount() string {
	return sys.account
}

func (sys *systemImpl) GetNode() string {
	return sys.colonio.GetLocalNid()
}

func (sys *systemImpl) Connect(url, account, token string) error {
	err := sys.colonio.Connect(url, token)
	if err != nil {
		return err
	}

	sys.account = account

	err = sys.evh.OnConnect()
	if err != nil {
		return err
	}

	return nil
}

func (sys *systemImpl) SetPosition(latitude, longitude float64) error {
	// convert L/L to radian
	_, _, err := sys.colonio.SetPosition(longitude*math.Pi/180.0, latitude*math.Pi/180.0)
	return err
}
