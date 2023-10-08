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
package three

import (
	"github.com/google/uuid"
	"github.com/llamerada-jp/oinari/api/core"
)

const (
	ResourceTypeThreeObject = core.ResourceType("object")
)

type Object struct {
	Meta *core.ObjectMeta `json:"meta"`
	Spec *ObjectSpec      `json:"spec"`
}

type ObjectSpec struct {
	Parts     []*PartSpec     `json:"parts"`
	Materials []*MaterialSpec `json:"materials"`
	Maps      []*TextureSpec  `json:"maps"`
	Position  core.Vector3    `json:"position"`
	// TODO: kind of the Z axis. (e.g. altitude, elevation)
}

type PartSpec struct {
	Name string `json:"name"`
	// Mesh   *MeshSpec   `json:"mesh,omitempty"`
	Sprite *SpriteSpec `json:"sprite,omitempty"`
	// Group  *GroupSpec  `json:"group,omitempty"`
}

/*
type MeshSpec struct {
}
*/

type SpriteSpec struct {
	// see: https://threejs.org/docs/?q=sprite#api/en/objects/Sprite
	Material string        `json:"material,omitempty"`
	Center   *core.Vector2 `json:"center,omitempty"`
}

/*
type GroupSpec struct {
}
*/

/*
type GeometrySpec struct {
}
*/

type MaterialSpec struct {
	Name           string              `json:"name"`
	SpriteMaterial *SpriteMaterialSpec `json:"spriteMaterial,omitempty"`
}

type MaterialBaseSpec struct {
	// see: https://threejs.org/docs/?q=material#api/en/materials/Material
}

type SpriteMaterialSpec struct {
	MaterialBaseSpec

	// see: https://threejs.org/docs/?q=material#api/en/materials/SpriteMaterial
	Color      *Color `json:"color,omitempty"`
	MapTexture string `json:"mapTexture,omitempty"`
}

type TextureSpec struct {
	Name       string          `json:"name"`
	URLTexture *URLTextureSpec `json:"urlTexture,omitempty"`
}

type TextureBaseSpec struct {
	// see: https://threejs.org/docs/?q=Texture#api/en/textures/Texture
}

type URLTextureSpec struct {
	TextureBaseSpec
	// instead of TextureLoader (https://threejs.org/docs/?q=Texture#api/en/loaders/TextureLoader)
	// load texture date on each node using URL
	URL string `json:"url"`
}

type Color struct {
	R float32 `json:"r"`
	G float32 `json:"g"`
	B float32 `json:"b"`
}

func GenerateObjectUUID() string {
	return uuid.Must(uuid.NewRandom()).String()
}

func (obj *Object) Validate() error {
	return nil
}
