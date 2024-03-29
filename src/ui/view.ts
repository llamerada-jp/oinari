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

import * as THREE from "three";

export interface ApplyObjectsRequest {
  objects: Object[];
}

export interface DeleteObjectsRequest {
  uuids: string[];
}

export interface Object {
  meta: ObjectMeta;
  spec: ObjectSpec;
}

interface ObjectMeta {
  type: string;
  name: string;
  owner: string;
  creatorNode: string;
  uuid: string;
  deletionTimestamp: string;
}

interface ObjectSpec {
  parts: PartSpec[];
  materials: MaterialSpec[];
  maps: TextureSpec[];
  position: Vec3;
}

interface PartSpec {
  name: string;
  sprite: SpriteSpec;
}

interface PartBaseSpec {
  scale: Scale;
}

interface Scale {
  default: Vec3;
  xr: Vec3;
  landscape: Vec3;
}

interface SpriteSpec extends PartBaseSpec {
  material: string;
  center: Vec2;
}

interface MaterialSpec {
  name: string;
  spriteMaterial: SpriteMaterialSpec;
}

interface MaterialBaseSpec {
  alphaTest: number;
}

interface SpriteMaterialSpec extends MaterialBaseSpec {
  color: Color;
  mapTexture: string;
}

interface TextureSpec {
  name: string;
  urlTexture: URLTextureSpec;
}

interface TextureBaseSpec {
  wrapS: number;
  wrapT: number;
  offset: Vec2;
  repeat: Vec2;
}

interface URLTextureSpec extends TextureBaseSpec {
  url: string;
}

interface Color {
  r: number;
  g: number;
  b: number;
}

interface Vec2 {
  x: number;
  y: number;
}

interface Vec3 {
  x: number;
  y: number;
  z: number;
}

interface TextureEntry {
  texture: THREE.Texture
  url: string
}

export const ScaleModeLandScape = "landscape";
export const ScaleModeXR = "xr";
export type ScaleMode = typeof ScaleModeLandScape | typeof ScaleModeXR;

export class ObjectWrapper extends THREE.Group {
  sprites: Map<string, THREE.Sprite>;
  materials: Map<string, THREE.Material>;
  textures: Map<string, TextureEntry>;
  objPosition: Vec3;
  position!: THREE.Vector3;
  scaleMode: ScaleMode;

  constructor(scaleMode: ScaleMode) {
    super();

    this.sprites = new Map<string, THREE.Sprite>();
    this.materials = new Map<string, THREE.Material>();
    this.textures = new Map<string, TextureEntry>();
    this.objPosition = { x: 0, y: 0, z: 0 } as Vec3;
    this.scaleMode = scaleMode;
  }

  applyObject(obj: Object): void {
    this.applyTextures(obj.spec.maps);
    this.applyMaterials(obj.spec.materials);
    this.applyParts(obj.spec.parts);
    this.objPosition = obj.spec.position;
  }

  transformPosition(transformer: google.maps.CoordinateTransformer): void {
    // TODO: this is wrong code. i couldn't find the correct way.
    let pos = transformer.fromLatLngAltitude({
      lat: this.objPosition.y,
      lng: this.objPosition.x,
      altitude: this.objPosition.z
    });
    this.position.set(pos[0], pos[1], pos[2]);
  }

  applyParts(parts: PartSpec[]): void {
    for (let part of parts) {
      this.applySprite(part);
    }

    for (let [name, obj] of this.sprites) {
      let using = false;
      for (let part of parts) {
        if (name === part.name) {
          using = true;
          break;
        }
      }
      if (!using) {
        this.remove(obj);
        this.sprites.delete(name);
      }
    }
  }

  applySprite(part: PartSpec): void {
    let sprite = this.sprites.get(part.name);
    let material = this.materials.get(part.sprite.material);
    if (material === undefined) {
      throw new Error("material not found");
    }
    if (material.type !== "SpriteMaterial") {
      throw new Error("material is not SpriteMaterial");
    }

    if (sprite !== undefined && sprite.material !== material) {
      this.remove(sprite);
      sprite = undefined;
    }

    if (sprite === undefined) {
      sprite = new THREE.Sprite(material as THREE.SpriteMaterial);
      this.add(sprite);
      this.sprites.set(part.name, sprite);
    }

    if (part.sprite.scale !== undefined) {
      let scale = this.selectScale(part.sprite.scale);
      if (scale && scale.x !== 0 && scale.y !== 0 && scale.z !== 0) {
        sprite.scale.set(scale.x, scale.y, scale.z);
      }
    }
  }

  applyMaterials(materials: MaterialSpec[]): void {
    for (let material of materials) {
      if (material.spriteMaterial !== undefined) {
        this.applySpriteMaterial(material);
      }
    }

    for (let [name, _] of this.materials) {
      let using = false;
      for (let material of materials) {
        if (name === material.name) {
          using = true;
          break;
        }
      }
      if (!using) {
        this.materials.delete(name);
      }
    }
  }

  applySpriteMaterial(material: MaterialSpec): void {
    let entry = this.materials.get(material.name);
    if (entry === undefined || entry.type !== "SpriteMaterial") {
      let color: THREE.Color;
      if (material.spriteMaterial === undefined || material.spriteMaterial.color === undefined) {
        color = new THREE.Color(0xffffff);
      } else {
        color = new THREE.Color(material.spriteMaterial.color.r, material.spriteMaterial.color.g, material.spriteMaterial.color.b);
      }
      entry = new THREE.SpriteMaterial({
        alphaTest: material.spriteMaterial.alphaTest,
        color: color,
        map: this.textures.get(material.spriteMaterial.mapTexture)?.texture,
      });
      this.materials.set(material.name, entry);
      return;
    }

    let spriteMaterial = entry as THREE.SpriteMaterial;
    let texture = this.textures.get(material.spriteMaterial.mapTexture)?.texture;
    if (spriteMaterial.map !== texture) {
      let color = material.spriteMaterial.color !== undefined ? { r: 1, g: 1, b: 1 } as THREE.Color : undefined;
      entry = new THREE.SpriteMaterial({
        color: color,
        map: this.textures.get(material.spriteMaterial.mapTexture)?.texture,
      });
      this.materials.set(material.name, entry);
      return;
    }

    if (material.spriteMaterial.color === undefined) {
      if (spriteMaterial.color !== undefined) {
        spriteMaterial.color = new THREE.Color(0xffffff);
      }

    } else if (spriteMaterial.color === undefined) {
      spriteMaterial.color = { r: 1, g: 1, b: 1 } as THREE.Color;

    } else if (spriteMaterial.color.r !== material.spriteMaterial.color.r ||
      spriteMaterial.color.g !== material.spriteMaterial.color.g ||
      spriteMaterial.color.b !== material.spriteMaterial.color.b) {
      spriteMaterial.color.setRGB(material.spriteMaterial.color.r, material.spriteMaterial.color.g, material.spriteMaterial.color.b);
    }
  }

  applyTextures(textures: TextureSpec[]): void {
    for (let texture of textures) {
      let entry = this.textures.get(texture.name);
      if (entry === undefined || entry.url !== texture.urlTexture.url) {
        entry = {
          texture: new THREE.TextureLoader().load(texture.urlTexture.url),
          url: texture.urlTexture.url,
        };
        this.textures.set(texture.name, entry);
      }

      if (texture.urlTexture.wrapS !== 0 && texture.urlTexture.wrapS !== entry.texture.wrapS) {
        entry.texture.wrapS = texture.urlTexture.wrapS as THREE.Wrapping;
      }
      if (texture.urlTexture.wrapT !== 0 && texture.urlTexture.wrapT !== entry.texture.wrapT) {
        entry.texture.wrapT = texture.urlTexture.wrapT as THREE.Wrapping;
      }
      if (texture.urlTexture.offset !== undefined) {
        entry.texture.offset.set(texture.urlTexture.offset.x, texture.urlTexture.offset.y);
      }
      if (texture.urlTexture.repeat !== undefined) {
        entry.texture.repeat.set(texture.urlTexture.repeat.x, texture.urlTexture.repeat.y);
      }
    }

    for (let [name, _] of this.textures) {
      let using = false;
      for (let texture of textures) {
        if (name === texture.name) {
          using = true;
          break;
        }
      }
      if (!using) {
        this.textures.delete(name);
      }
    }
  }

  selectScale(scale: Scale): Vec3 {
    if (this.scaleMode === ScaleModeLandScape && scale.landscape !== undefined) {
      return scale.landscape;
    } else if (this.scaleMode === ScaleModeXR && scale.xr !== undefined) {
      return scale.xr;
    } else {
      return scale.default;
    }
  }
}