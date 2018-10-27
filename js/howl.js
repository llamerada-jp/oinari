"use strict";

const RADIUS = 1000;  // メッセージの送信半径[m]

let v = new vein.Vein();
let pubsub2d = null;
let lat = null;
let lon = null;
let wasConnect = false;
let wasGNSS    = false;
let wasStart   = false;
let wasMap     = false;
let debugMode  = false;
let geoWatchId = null;
let map = null;
let mapHeight;
let mapWidth;
let mapDefer = $.Deferred();
let index = 0;
let markers = [];
let gmLines = [];
let gmCircle = null;
let resizeTimer = 0;
let myNid = null;

function init() {
  $('#modal-init').modal({
    backdrop: 'static',
    keyboard: false,
    show: true
  });

  $.when(
    initNet(),
    initGNSS(),
    mapDefer.promise()

  ).done(() => {
    return wait(1);

  }).done(() => {
    console.log('finish');
    wasStart = true;
    initDebug();
    $('#modal-init').modal('hide');
  });
}

function initGNSS() {
  let defer = $.Deferred();
  let isFirst = true;
  geoWatchId = navigator.geolocation.watchPosition((position) => {
    if (!debugMode &&
        (lon != position.coords.longitude || lat != position.coords.latitude)) {
      lon = position.coords.longitude;
      lat = position.coords.latitude;
      if (wasConnect === true) {
        let [x, y] = convertDeg2Rad(lon, lat);
        v.setPosition(x, y);
      }
      if (wasMap === true) {
        map.panTo({lat: lat, lng: lon});
      }
    }
    if (isFirst) {
      $('#status-gnss').html('<span class="badge badge-success">Success</span>');
      isFirst = false;
      wasGNSS = true;

      $('#map .gm-style').addClass('transform-parent');
      $($('#map .gm-style').children()[0]).addClass('transform-target');

      defer.resolve();
    }

  }, (err) => {
    // Error
    $('#status-gnss').html('<span class="badge badge-danger">Failed</span>');
    console.warn('ERROR(' + err.code + '): ' + err.message);
    defer.reject();

  }, {
	  "enableHighAccuracy": false ,
	  "timeout": 1000000 ,
	  "maximumAge": 0 ,
  });
  return defer.promise();
}

function recv(data) {
  let message = JSON.parse(data);
  // 格納場所がない場合は作る
  if (!(message.id in markers)) {
    markers[message.id] = {};
  }
  markers[message.id].latlng = {
    lat: message.lat,
    lng: message.lon
  };
  if ('image' in message) {
    markers[message.id].text = '';
    markers[message.id].image = message.image;
  }
  if ('text' in message) {
    markers[message.id].text = message.text;
    markers[message.id].image = null;
  }
  mark(message.id);
}

function send(param) {
  if (myNid == null) return;
  let message = {
    id: myNid,
    lon: lon,
    lat: lat
  };
  $.extend(message, param);
  let [x, y] = convertDeg2Rad(lon, lat);
  pubsub2d.publish('howl', x, y, RADIUS, JSON.stringify(message));
}

function markAll() {
  for (let id of Object.keys(markers)) {
    mark(id);
  }
}

function mark(id) {
  let bounds = map.getBounds();
  if (bounds.contains(markers[id].latlng)) {
    // marker要素がない場合は新規作背
    if (!('tag' in markers[id]) || markers[id].tag === false) {
      markers[id].tag = $('<div>').addClass('marker');
      markers[id].tag.append($('<div>').addClass('balloon').attr('id', 'balloon' + id));
      markers[id].tag.append($('<img>').addClass('image').attr('id', 'image' + id));
      markers[id].tag.append($('<img>').attr('src', 'img/h2.png').css('margin', '0 auto'));
      $('.transform-target').append(markers[id].tag);
    }
    // 画像の表示
    let $balloon = $('#balloon' + id);
    let $image   = $('#image' + id);
    if ('image' in markers[id] && markers[id].image !== null) {
      $image.show();
      $image.attr('src', markers[id].image);
    } else {
      $image.hide();
    }
    // 吹き出しの表示
    if ('text' in markers[id] && markers[id].text.trim() !== '') {
      $balloon.show();
      $balloon.text(markers[id].text);
    } else {
      $balloon.hide();
    }
    // markerの位置調整
    let w = $('.transform-target').width();
    let x = 0;
    if (bounds.getNorthEast().lng() > bounds.getSouthWest().lng()) {
      x = w * (markers[id].latlng.lng - bounds.getSouthWest().lng()) /
      (bounds.getNorthEast().lng() - bounds.getSouthWest().lng());
      // 1/sin(30) で補正の必要がある？
      x = (w * 0.5) + (x - w * 0.5) * 2 - (markers[id].tag.width() * 0.5);
    } else {
      console.log('todo');
    }
    let h = $('.transform-target').height();
    let y = 0;
    if (bounds.getNorthEast().lat() > bounds.getSouthWest().lat()) {
      y = h - h * (markers[id].latlng.lat - bounds.getSouthWest().lat()) /
      (bounds.getNorthEast().lat() - bounds.getSouthWest().lat());
      // 1/cos(30)で補正の必要がある？
      y = (h * 0.5) + (y - h * 0.5) * 1.15 - markers[id].tag.height();
    } else {
      console.log('todo');
    }
    markers[id].tag.css('left', x + 'px');
    markers[id].tag.css('top', y + 'px')

  } else {
    if ('tag' in markers[id] && markers[id].tag !== false) {
      // marker要素がある場合は削除
      console.log('remove')
      $(markers[id].tag).remove();
      markers[id].tag = false;
    }
  }
}

function getDistance(bLon, bLat) {
  const R = 6378137;
  let [x1, y1] = convertDeg2Rad(lon, lat);
  let [x2, y2] = convertDeg2Rad(bLon, bLat);
  let avrX = (x1 - x2) / 2;
  let avrY = (y1 - y2) / 2;
  
  return R * 2 * Math.asin(Math.sqrt(Math.pow(Math.sin(avrY), 2) + Math.cos(y1) *
    Math.cos(y2) * Math.pow(Math.sin(avrX), 2)));
}

function initNet() {
  let defer = $.Deferred();
  v.init().then(() => {
    console.log('vein connect');
    v.on('log', onGetLog);
    v.on('debug', onGetDebug);
    return v.connect('ws://veindev:8080/vein/ws', '');

  }).then(() => {
    myNid = v.getMyNid();

    if (wasGNSS === true) {
      let [x, y] = convertDeg2Rad(lon, lat);
      v.setPosition(x, y);
    }

    pubsub2d = v.accessPubsub2D('pubsub2d');
    pubsub2d.on('howl', recv);

    wasConnect = true;
    console.log('vein success');
    $('#status-net').html('<span class="badge badge-success">Success</span>');
    defer.resolve();

  }).catch((e) => {
    console.error('vein failed');
    console.log(e);
    $('#status-net').html('<span class="badge badge-danger">Failed</span>');
    defer.reject();
  });
  return defer.promise();
}

function onGetLog(log) {
  // console.log(log);
}

function onGetDebug(event) {
  // console.log(event);

  if (event.event === vein.Vein.DEBUG_EVENT_KNOWN2D) {
    for (let line of gmLines) {
      line.setMap(null);
    }
    gmLines = [];

    if (debugMode) {
      let nodes = event.content.nodes;
      for (let link of event.content.links) {
        let p1 = (link[0] == vein.Vein.NID_THIS ?
                  [lon, lat] : convertRad2Deg(nodes[link[0]][0], nodes[link[0]][1]));
        let p2 = (link[1] == vein.Vein.NID_THIS ?
                  [lon, lat] : convertRad2Deg(nodes[link[1]][0], nodes[link[1]][1]));
        let polyLine = new google.maps.Polyline({
          path: [{lng: p1[0], lat: p1[1]}, {lng: p2[0], lat: p2[1]}],
          strokeColor: '#FF0000',
          strokeOpacity: 1.0,
          strokeWeight: 2
        });
        polyLine.setMap(map);
        gmLines.push(polyLine);
      }
    }
  }
}

function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: {
      lat: Math.random() * 180.0 -  90.0,
      lng: Math.random() * 360.0 - 180.0
    },
    disableDefaultUI: true,
    draggable: false,
    mapTypeControl: false,
    zoomControl: true,
    zoom: 18 // 地図のズームを指定
  });

  map.addListener('bounds_changed', function() {
    markAll();
  });

  map.addListener('center_changed', function() {
    if (myNid != null) {
      if (!(myNid in markers)) {
        markers[myNid] = {};
      }
      markers[myNid].latlng = {
        lat: map.getCenter().lat(),
        lng: map.getCenter().lng()
      };
      mark(myNid);
      send();
    }
  });

  wasMap = true;
  $('#status-map').html('<span class="badge badge-success">Success</span>');
  mapDefer.resolve();
}

function initDebug() {
  $('#switch-debug-mode').on('click', function() {
    debugMode = !debugMode;
    // 地図の中心位置を変更できるように
    map.set('draggable', debugMode);
  });

  setInterval(function() {
    let center = map.getCenter();
    if (debugMode) {
      // 地図の中心位置をGPSの座標の代わりに利用
      if (lon != center.lng() || lat != center.lat()) {
        lon = center.lng();
        lat = center.lat();
        if (wasConnect === true) {
          let [x, y] = convertDeg2Rad(lon, lat);
          v.setPosition(x, y);
        }
      }

      // メッセージ到達範囲の円を表示
      if (gmCircle === null) {
        gmCircle = new google.maps.Circle({
          center: {lng: lon, lat: lat},
          fillOpacity: 0,
          map: map,
          radius: RADIUS,
          strokeColor: '#ff0000',
          strokeOpacity: 1,
          strokeWeight: 1
        });
      } else {
        gmCircle.setCenter({lng: lon, lat: lat});
      }

    } else {
      // 円を非表示
      if (gmCircle !== null) {
        gmCircle.setMap(null);
        gmCircle = null;
      }
    }
  }, 1000);
}

function wait(sec) {
  let defer = $.Deferred();
  setTimeout(() => {
    defer.resolve();
  }, sec * 1000);
  return defer.promise();
}

function convertDeg2Rad(lon, lat) {
  while (lat < -90.0) {
    lat += 360.0;
  }
  while (270.0 <= lat) {
    lat -= 360.0;
  }

  if (180.0 <= lat) {
    lon += 180.0;
    lat = -1.0 * (lat - 180.0);
  } else if (90.0 <= lat) {
    lon += 180.0;
    lat = 180.0 - lat;
  }

  while (lon < -180.0) {
    lon += 360.0;
  }
  while (180.0 <= lon) {
    lon -= 360.0;
  }
  return [Math.PI * lon / 180,
          Math.PI * lat / 180];
}

function convertRad2Deg(x, y) {
  return [180.0 * x / Math.PI,
          180.0 * y / Math.PI]
}

// 表示領域リサイズ時に地図の大きさなどを変更する
$(window).on('load resize', () => {
  if (resizeTimer > 0) {
    clearTimeout(resizeTimer);
  }
  
  resizeTimer = setTimeout(function () {
    let fieldHeight = $(window).height() - $('header').height() - $('footer').height();
    let fieldWidth  = $(window).width();
    let $map = $('#map');
    let $lists = $('#lists');

    if (fieldWidth > fieldHeight) {
      // 横向き
      mapHeight = fieldHeight;
      mapWidth  = fieldHeight;
      $map.height(mapHeight);
      $map.width (mapWidth);
      $lists.height(fieldHeight);
      $lists.width (fieldWidth - mapWidth);

    } else {
      // 縦向き
      mapHeight = fieldWidth;
      mapWidth  = fieldWidth;
      if (mapHeight > fieldHeight / 2) mapHeight = fieldHeight / 2;
      $map.height(mapHeight);
      $map.width (mapWidth);
      $lists.height(fieldHeight - mapHeight);
      $lists.width (fieldWidth);
    }
  }, 50);
});

// ボタンを押したらカメラ起動
$('#btn-camera').on('click', function() {
  $('[name="capture"]').click();
});

// 画像が選択されたらダンプして送る
$('[name="capture"]').on('change', function() {
  let reader = new FileReader();

  // 読み込み完了時のイベント
  reader.onload = function() {
    send({
      image: reader.result
    });
  }

  // 読み込みを実行
  reader.readAsDataURL(this.files[0]);
});

// ボタンを押したらメッセージを送信
$('#btn-howl').on('click', function() {
  let $message = $('#message');
  send({
    text: $message.val()
  });
  $message.val('');
});

// start
$(window).on('load', () => {
  init();
});
