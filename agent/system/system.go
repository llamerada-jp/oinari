package system

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type EventHandler interface {
	OnConnect() error
}

type System interface {
	Start(ctx context.Context) error
	TellInitComplete() error
	GetAccount() string

	connect(url, account, token string) error
	setPosition(latitude, longitude float64) error
}

type systemImpl struct {
	colonio colonio.Colonio
	evh     EventHandler
	cd      CommandDriver
	account string
}

func init() {
	rand.Seed(time.Now().UnixMicro())
}

func NewSystem(col colonio.Colonio, evh EventHandler, cd CommandDriver) System {
	return &systemImpl{
		colonio: col,
		evh:     evh,
		cd:      cd,
	}
}

func (sys *systemImpl) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (sys *systemImpl) TellInitComplete() error {
	return sys.cd.TellInitComplete()
}

func (sys *systemImpl) GetAccount() string {
	return sys.account
}

func (sys *systemImpl) connect(url, account, token string) error {
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

func (sys *systemImpl) setPosition(latitude, longitude float64) error {
	// convert L/L to radian
	_, _, err := sys.colonio.SetPosition(longitude*math.Pi/180.0, latitude*math.Pi/90.0)
	return err
}
