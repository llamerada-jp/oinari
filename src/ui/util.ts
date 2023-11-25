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

export type clickEventCB = () => void;

export function addListItem(list: HTMLElement, temp: HTMLTemplateElement, contents: Map<string, string | clickEventCB>): void {
  let item = temp.content.cloneNode(true) as HTMLElement;
  for (const [key, value] of contents) {
    if (typeof value === "string") {
      (item.querySelector(key) as HTMLElement).innerText = value;
    } else {
      item.querySelector(key)?.addEventListener("click", value);
    }
  }
  list.append(item);
}

export function closeModal(elName: string): void {
  // I want to use bootstrap.Modal.getInstance causes an extra backdrop and doesn't fade when calling the hide method.
  // This code is ugly workaround.
  let closer = document.getElementById(elName);
  if (closer == null) {
    throw new Error("HTMLElement of modal closer is not found");
  }
  closer.dispatchEvent(new Event("click"));
}

export function makeListEmptyHide(list: HTMLElement | null): void {
  if (list == null) return;
  list.classList.add("d-none");
  while (list.firstChild) {
    list.removeChild(list.firstChild);
  }
}

export function makeListShow(list: HTMLElement | null): void {
  if (list == null) return;
  list.classList.remove("d-none");
}

export function setSelectValue(select: HTMLSelectElement, value: string): void {
  for (let i = 0; i < select.options.length; i++) {
    if (select.options[i].value === value) {
      select.selectedIndex = i;
      break;
    }
  }
}