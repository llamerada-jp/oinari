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

import * as LS from "../local_settings";
import * as UTIL from "./util";

const DummyNames = [
  "Aardvark",
  "Alligator",
  "Alpaca",
  "Angelfish",
  "Ant",
  "Anteater",
  "Archerfish",
  "Armadillo",
  "Axolotl",
  "Barracuda",
  "Bat",
  "Bear",
  "Beaver",
  "Bee",
  "Bird",
  "Bison",
  "Blobfish",
  "Bluejay",
  "Butterfly",
  "Camel",
  "Capybara",
  "Caracal",
  "Cassowary",
  "Cat",
  "Chameleon",
  "Cheetah",
  "Chimpanzee",
  "Cougar",
  "Crab",
  "Crocodile",
  "Cuttlefish",
  "Deer",
  "Dhole",
  "Dingo",
  "Dog",
  "Dolphin",
  "Eagle",
  "Echidna",
  "Elephant",
  "Fennec",
  "Fish",
  "Flyingfish",
  "Fox",
  "Frog",
  "Gar",
  "Gaur",
  "Gila",
  "Giraffe",
  "Goblin",
  "Gorilla",
  "Hammerhead",
  "Hedgehog",
  "Hippopotamus",
  "Horse",
  "Hummingbird",
  "Jellyfish",
  "Kangaroo",
  "Kingfisher",
  "Koala",
  "Kookaburra",
  "Ladybug",
  "Lemming",
  "Lemur",
  "Leopard",
  "Lion",
  "Llama",
  "Lynx",
  "Manatee",
  "Mantis",
  "Mantis",
  "Markhor",
  "Marmoset",
  "Meerkat",
  "Mole",
  "Monkey",
  "Muskox",
  "Narwhal",
  "Nudibranch",
  "Numbat",
  "Ocelot",
  "Octopus",
  "Okapi",
  "Ostrich",
  "Owl",
  "Panda",
  "Panda",
  "Pangolin",
  "Parrot",
  "Peacock",
  "Peafowl",
  "Pelican",
  "Penguin",
  "Pika",
  "Platypus",
  "Polarbear",
  "Pufferfish",
  "Quokka",
  "Quoll",
  "Raccoon",
  "Rattlesnake",
  "Rhino",
  "Saola",
  "Shark",
  "Shrimp",
  "Snail",
  "Snake",
  "Spider",
  "Squirrel",
  "Starfish",
  "Stork",
  "Sunfish",
  "Tapir",
  "Tasmanian",
  "Tasmanian",
  "Tiger",
  "Tortoise",
  "Turtle",
  "Vulture",
  "Wallaby",
  "Walrus",
  "Warthog",
  "Wobbegong",
  "Wolf",
  "Wolverine",
  "Wombat",
  "Zebra"
];

const initSettingsDivID = "initSettings";
const deviceNameGenButtonID = "initSettingsDeviceNameRandom";
const deviceNameInputID = "initSettingsDeviceName";
const viewTypeInputID = "initSettingsViewType";
const localStoreInputID = "initSettingsLocalStore";
const syncGNSSInputID = "initSettingsSyncGNSS";
const spawnPositionInputID = "initSettingsSpawnPosition";
const allowApplicationsInputID = "initSettingsAllowApplications";
const submitButtonID = "initSettingsSubmit";

let localSettings: LS.LocalSettings;
let onSubmit: () => void;

export function init(ls: LS.LocalSettings, submit: () => void): void {
  localSettings = ls;
  onSubmit = submit;

  initElements();
  loadSettings();
  updateInputs();
}

function initElements(): void {
  let deviceNameGenButton = document.getElementById(deviceNameGenButtonID);
  let deviceNameInput = document.getElementById(deviceNameInputID) as HTMLInputElement;
  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  let spawnPositionInput = document.getElementById(spawnPositionInputID) as HTMLInputElement;
  let submitButton = document.getElementById(submitButtonID) as HTMLButtonElement;

  deviceNameInput?.addEventListener("change", () => {
    updateInputs();
  });

  deviceNameGenButton?.addEventListener("click", () => {
    deviceNameInput!.value = DummyNames[Math.floor(Math.random() * DummyNames.length)];
    deviceNameInput!.dispatchEvent(new Event("change"));
  });

  if (!!navigator.geolocation) {
    syncGNSSInput?.addEventListener("click", () => {
      if (syncGNSSInput!.checked) {
        spawnPositionInput!.disabled = true;
      } else {
        spawnPositionInput!.disabled = false;
      }
    });

  } else {
    syncGNSSInput!.disabled = true;
    syncGNSSInput!.checked = false;
    spawnPositionInput!.disabled = false;
  }

  submitButton?.addEventListener("click", () => {
    if (!validateInputs()) {
      throw new Error("could submit with invalid inputs.");
    }
    storeSettings();
    onSubmit();
  });
}

export function show(): void {
  loadSettings();
  let initSettingsEl = document.getElementById(initSettingsDivID);
  initSettingsEl!.classList.remove("d-none");
}

export function hide(): void {
  let initSettingsEl = document.getElementById(initSettingsDivID);
  initSettingsEl!.classList.add("d-none");
}

function updateInputs(): void {
  let submitButton = document.getElementById(submitButtonID) as HTMLButtonElement;

  if (validateInputs()) {
    submitButton!.disabled = false;
  } else {
    submitButton!.disabled = true;
  }
}

function validateInputs(): boolean {
  let deviceNameInput = document.getElementById(deviceNameInputID) as HTMLInputElement;

  if (deviceNameInput!.value.length === 0) {
    return false;
  }

  return true;
}

function loadSettings(): void {
  let deviceNameInput = document.getElementById(deviceNameInputID) as HTMLInputElement;
  deviceNameInput!.value = localSettings.deviceName;

  let viewTypeInput = document.getElementById(viewTypeInputID) as HTMLSelectElement;
  UTIL.setSelectValue(viewTypeInput!, localSettings.viewType)

  let localStoreInput = document.getElementById(localStoreInputID) as HTMLInputElement;
  localStoreInput!.checked = localSettings.enableLocalStore;

  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  syncGNSSInput!.checked = localSettings.enableGNSS;

  if (localSettings.position) {
    let spawnPositionSelect = document.getElementById(spawnPositionInputID) as HTMLSelectElement;
    UTIL.setSelectValue(spawnPositionSelect, localSettings.position);
  }

  let allowApplicationsInput = document.getElementById(allowApplicationsInputID) as HTMLSelectElement;
  UTIL.setSelectValue(allowApplicationsInput!, localSettings.allowApplications);
}

function storeSettings(): void {
  let deviceNameInput = document.getElementById(deviceNameInputID) as HTMLInputElement;
  localSettings.deviceName = deviceNameInput!.value;

  let viewTypeInput = document.getElementById(viewTypeInputID) as HTMLInputElement;
  localSettings.viewType = viewTypeInput!.value;

  let localStoreInput = document.getElementById(localStoreInputID) as HTMLInputElement;
  localSettings.enableLocalStore = localStoreInput!.checked;

  let syncGNSSInput = document.getElementById(syncGNSSInputID) as HTMLInputElement;
  localSettings.enableGNSS = syncGNSSInput!.checked;

  if (!localSettings.enableGNSS) {
    let spawnPositionSelect = document.getElementById(spawnPositionInputID) as HTMLSelectElement;
    let positionStr = spawnPositionSelect.options[spawnPositionSelect.selectedIndex].value;
    localSettings.position = positionStr;
  }

  let allowApplicationsSelect = document.getElementById(allowApplicationsInputID) as HTMLSelectElement;
  localSettings.allowApplications = allowApplicationsSelect.options[allowApplicationsSelect.selectedIndex].value;
}
