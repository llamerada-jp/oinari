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

	threeAPI "github.com/llamerada-jp/oinari/api/three"
	"github.com/llamerada-jp/oinari/lib/oinari"
	threeLib "github.com/llamerada-jp/oinari/lib/three"
)

type fox struct {
	mtx    sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	three  threeLib.ThreeAPI

	ObjectUUID string `json:"objectUUID"`
	Frame      int    `json:"frame"`
}

var _ oinari.Application = (*fox)(nil)

func (f *fox) Setup(isInitialize bool, record []byte) error {
	if isInitialize || record == nil {
		f.initObject()

	} else {
		err := json.Unmarshal(record, f)
		if err != nil {
			return err
		}
	}

	f.start()

	return nil
}

// create new object
func (f *fox) initObject() error {
	var err error
	// oinari system fill position using default position when create object without specifying to position.
	f.ObjectUUID, err = f.three.CreateObject("fox", &threeAPI.ObjectSpec{
		Parts: []*threeAPI.PartSpec{
			{
				Name: "sprite",
				Sprite: &threeAPI.SpriteSpec{
					PartBaseSpec: threeAPI.PartBaseSpec{
						Scale: map[string]*threeAPI.Vector3{
							threeAPI.ScaleDefault: {
								X: 100 / 3,
								Y: 100,
								Z: 1,
							},
							threeAPI.ScaleXR: {
								X: 1.0 / 3,
								Y: 1.0,
								Z: 1,
							},
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
	})

	return err
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
	// get object to get current position
	object, err := f.three.GetObject(f.ObjectUUID)
	if err != nil {
		fmt.Println("🦊 get object error:", err)
	}
	object.Spec.Position.X += rand.Float64() * 0.00001
	object.Spec.Position.Y += rand.Float64() * 0.00001

	f.Frame++
	object.Spec.Maps[0].URLTexture.Offset.X = float64(f.Frame%3) * 0.333333

	if err := f.three.UpdateObject(f.ObjectUUID, object.Spec); err != nil {
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
