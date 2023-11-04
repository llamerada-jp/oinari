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
	"os"
	"sync"
	"time"

	threeAPI "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/oinari"
	threeLib "github.com/llamerada-jp/oinari/lib/three"
)

type fox struct {
	mtx    sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	three  threeLib.ThreeAPI
	object *threeAPI.ObjectSpec

	ObjectUUID string  `json:"objectUUID"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Frame      int     `json:"frame"`
}

var _ oinari.Application = (*fox)(nil)

func (f *fox) Setup(isInitialize bool, record []byte) error {
	f.object = &threeAPI.ObjectSpec{
		Parts: []*threeAPI.PartSpec{
			{
				Name: "sprite",
				Sprite: &threeAPI.SpriteSpec{
					PartBaseSpec: threeAPI.PartBaseSpec{
						Scale: &threeAPI.Vector3{
							X: 100 / 3,
							Y: 100,
							Z: 1,
						},
					},
					Material: "matFox",
					Center: &threeAPI.Vector2{
						X: 0,
						Y: 0,
					},
				},
			},
		},
		Materials: []*threeAPI.MaterialSpec{
			{
				Name: "matFox",
				SpriteMaterial: &threeAPI.SpriteMaterialSpec{
					MaterialBaseSpec: threeAPI.MaterialBaseSpec{
						AlphaTest: 0.1,
					},
					MapTexture: "mapFox",
				},
			},
		},
		Maps: []*threeAPI.TextureSpec{
			{
				Name: "mapFox",
				URLTexture: &threeAPI.URLTextureSpec{
					TextureBaseSpec: threeAPI.TextureBaseSpec{
						WrapS: threeAPI.RepeatWrapping,
						WrapT: threeAPI.RepeatWrapping,
						Repeat: &threeAPI.Vector2{
							X: 0.333333,
							Y: 1,
						},
						Offset: &threeAPI.Vector2{},
					},
					URL: "/img/fox1b.png",
				},
			},
		},
	}

	if isInitialize || record == nil {
		f.Latitude = 35.6594945
		f.Longitude = 139.6999859
		f.object.Position = &threeAPI.Vector3{
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

		f.object.Position = &threeAPI.Vector3{
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
	/*
		f.Longitude += rand.Float64() * 0.00001
		f.Latitude += rand.Float64() * 0.00001
		//*/

	f.object.Position.X = f.Longitude
	f.object.Position.Y = f.Latitude

	f.Frame++
	f.object.Maps[0].URLTexture.Offset.X = float64(f.Frame%3) * 0.333333

	if err := f.three.UpdateObject(f.ObjectUUID, f.object); err != nil {
		fmt.Println("🦊 update object error:", err)
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
