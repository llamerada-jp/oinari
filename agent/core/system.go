package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/llamerada-jp/colonio/go/colonio"
)

type SystemEventHandler interface {
	OnConnect(sys *System) error
}

type System struct {
	online  bool
	colonio colonio.Colonio
	evh     SystemEventHandler
	gcd     GlobalCommandDriver
	lcd     LocalCommandDriver
	pods    map[string]*PodImpl
}

func init() {
	rand.Seed(time.Now().UnixMicro())
}

func NewSystem(col colonio.Colonio, evh SystemEventHandler, gcd GlobalCommandDriver, lcd LocalCommandDriver) *System {
	return &System{
		online:  false,
		colonio: col,
		evh:     evh,
		gcd:     gcd,
		lcd:     lcd,
		pods:    make(map[string]*PodImpl),
	}
}

func (sys *System) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if !sys.online {
				continue
			}
			err := sys.loop(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (sys *System) loop(ctx context.Context) error {
	err := sys.dealResources()
	if err != nil {
		return err
	}

	for uuid, pod := range sys.pods {
		effective, err := pod.Update(ctx, sys.colonio)
		if err != nil {
			return err
		}
		if !effective {
			delete(sys.pods, uuid)
		}
	}

	return nil
}

func (sys *System) dealResources() error {
	resources := make([]struct {
		resourceType ResourceType
		js           string
	}, 0)

	// to avoid dead-lock of ForeachLocalValue, don't call colonio's method in the callback func
	localData := sys.colonio.KvsGetLocalData()
	defer localData.Free()
	for _, key := range localData.GetKeys() {
		v, err := localData.GetValue(key)
		if err != nil {
			return err
		}
		js, err := v.GetString()
		if err != nil {
			return err
		}
		resourceEntry := strings.Split(key, "/")
		if len(resourceEntry) != 2 {
			return fmt.Errorf("local value key is not supported format", key)
		}
		resources = append(resources, struct {
			resourceType ResourceType
			js           string
		}{
			resourceType: ResourceType(resourceEntry[0]),
			js:           js,
		})
	}

	for _, resource := range resources {
		err := sys.dealResource(resource.resourceType, resource.js)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (sys *System) dealResource(t ResourceType, js string) error {
	switch t {
	case ResourceTypePod:
		var pod Pod
		err := json.Unmarshal([]byte(js), &pod)
		if err != nil {
			return err
		}
		err = sys.schedulePod(&pod)
		if err != nil {
			return err
		}

	default:
		log.Printf("debug unhandled resource %s", string(t))
	}
	return nil
}

func (sys *System) schedulePod(pod *Pod) error {
	if pod.Status.RunningNode != "" {
		return nil
	}

	switch pod.Spec.Scheduler.Type {
	case "creator":
		err := sys.gcd.EncouragePod(pod.Spec.CreatorNode, pod.Meta.Uuid)
		return err

	default:
		return fmt.Errorf("Unsupported scheduling policy:%s", pod.Spec.Scheduler.Type)
	}
}

func (sys *System) Connect(url, token string) error {
	err := sys.colonio.Connect(url, token)
	if err != nil {
		return err
	}

	err = sys.evh.OnConnect(sys)
	if err != nil {
		return err
	}

	sys.online = true
	return nil
}

func (sys *System) ApplyPod(name, image string) (string, error) {
	pod := Pod{
		Meta: ObjectMeta{
			Type: ResourceTypePod,
			Name: name,
		},
		Spec: PodSpec{
			CreatorNode: sys.colonio.GetLocalNid(),
			// Support single container pod temporary.
			Containers: []ContainerSpec{
				{
					Name:  "main",
					Image: image,
				},
			},
			Scheduler: SchedulerSpec{
				Type: "creator",
			},
		},
	}
	// Retry with new uuid if the pod having the same uuid is exist.
	for {
		pod.Meta.Uuid = uuid.Must(uuid.NewRandom()).String()
		key := string(ResourceTypePod) + "/" + pod.Meta.Uuid
		js, err := json.Marshal(pod)
		if err != nil {
			return "", err
		}
		err = sys.colonio.KvsSet(key, string(js), colonio.KvsProhibitOverwrite)
		// TODO: retry only if the same uuid id exists
		if err == nil {
			return pod.Meta.Uuid, nil
		}
	}
}

func (sys *System) Terminate(uuid string) error {
	// TODO
	log.Println("fix it")
	return nil
}

func (sys *System) SetPosition(latitude, longitude float64) error {
	// convert L/L to radian
	_, _, err := sys.colonio.SetPosition(longitude*math.Pi/180.0, latitude*math.Pi/90.0)
	return err
}

func (sys *System) EncouragePod(ctx context.Context, uuid string) error {
	// skip if pod exists
	if _, ok := sys.pods[uuid]; ok {
		return nil
	}

	// create and update pod
	pod := NewPod(uuid)
	sys.pods[uuid] = pod
	effective, err := pod.Update(ctx, sys.colonio)
	if err != nil {
		return err
	}
	if !effective {
		delete(sys.pods, uuid)
	}

	return nil
}
