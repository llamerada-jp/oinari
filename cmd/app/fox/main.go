/*
 * Copyright 2018 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/llamerada-jp/oinari/api/core"
	api "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/oinari"
	threeLib "github.com/llamerada-jp/oinari/lib/three"
)

type fox struct {
	mtx    sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	three  threeLib.ThreeAPI
	object *api.ObjectSpec

	ObjectUUID string  `json:"objectUUID"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

var _ oinari.Application = (*fox)(nil)

func (f *fox) Setup(isInitialize bool, record []byte) error {
	f.object = &api.ObjectSpec{
		Parts: []*api.PartSpec{
			{
				Name: "sprite",
				Sprite: &api.SpriteSpec{
					Material: "matFox",
					Center: &core.Vector2{
						X: 0,
						Y: 0,
					},
				},
			},
		},
		Materials: []*api.MaterialSpec{
			{
				Name: "matFox",
				SpriteMaterial: &api.SpriteMaterialSpec{
					MapTexture: "mapFox",
				},
			},
		},
		Maps: []*api.TextureSpec{
			{
				Name: "mapFox",
				URLTexture: &api.URLTextureSpec{
					URL: "/img/fox1b.png",
				},
			},
		},
	}

	if isInitialize || record == nil {
		f.Latitude = 35.6594945
		f.Longitude = 139.6999859
		f.object.Position = core.Vector3{
			X: f.Longitude,
			Y: f.Latitude,
			Z: 0,
		}

		uuid, err := f.three.CreateObject("fox", f.object)
		if err != nil {
			return err
		}
		f.ObjectUUID = uuid

	} else {
		err := json.Unmarshal(record, f)
		if err != nil {
			return err
		}

		f.object.Position = core.Vector3{
			X: f.Longitude,
			Y: f.Latitude,
			Z: 0,
		}
	}

	f.start()

	return nil
}

func (f *fox) Marshal() ([]byte, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	return json.Marshal(f)
}

func (f *fox) Teardown(isFinalize bool) ([]byte, error) {
	f.cancel()

	if isFinalize && f.ObjectUUID != "" {
		f.three.DeleteObject(f.ObjectUUID)
	}

	return f.Marshal()
}

func (f *fox) start() {
	go func() {
		for {
			select {
			case <-f.ctx.Done():
				return

			default:
				time.Sleep(1 * time.Second)
				f.loop()
			}
		}
	}()
}

func (f *fox) loop() {
	f.Longitude += rand.Float64() * 0.00001
	f.Latitude += rand.Float64() * 0.00001

	f.object.Position.X = f.Longitude
	f.object.Position.Y = f.Latitude

	fmt.Printf("🦊 at %.3f, %.3f\n", f.Longitude, f.Latitude)

	if err := f.three.UpdateObject(f.ObjectUUID, f.object); err != nil {
		fmt.Println("🦊 update object:", err)
	}
}

func main() {
	c, cancel := context.WithCancel(context.Background())

	app := &fox{
		ctx:    c,
		cancel: cancel,
		three:  threeLib.NewThreeAPI(),
	}

	mgr := oinari.NewManager()
	mgr.Use(app.three)

	if err := mgr.Run(app); err != nil {
		fmt.Println("🦊 run:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
