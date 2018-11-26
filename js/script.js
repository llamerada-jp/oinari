"use strict";

// メッセージの送信半径[m]
const MESSAGE_RADIUS = 1000;
const ROPE_RADIUS    = 100;
const STEP_RADIUS    = 50;
// 画像サイズ[px]
const PICTURE_SIZE   = 512;
//
const R = 6378137;
const O = 2 * Math.PI * R;

// 切断とみなす時間[sec]
const TIMEOUT = 60;
// 発言後、吹き出しを消すまでの時間[sec]
const EMIT_TIMEOUT = 60;
// GNSSのタイムアウト時間
const GNSS_TIMEOUT = 30 * 1000;
//
const CHANNEL_NAME = 'hoge';

// libvein関連
let v = new vein.Vein();
let pubsub2d = null;
let localNid = null;

// GNSSから取得した緯度経度
let realLocation = {};
// キャラクター移動先情報
let dstLocation = null;
let arrivalTime = null;

// ステータス
const STATUS_NONE = Symbol('none');
const STATUS_OK   = Symbol('ok');
const STATUS_NG   = Symbol('ng');
let veinStatus = STATUS_NONE;
let gnssStatus = STATUS_NONE;

// 表示名
let nickname = '';

let gmap = null;
let markers = [];
let gmLines = [];
let gmCircle = null;

// ウインドウサイズ変更時利用するタイマー
let relayoutTimer = 0;

// Entry point.
$(window).on('load', () => {
  // language
  if ((window.navigator.userLanguage ||
       window.navigator.language ||
       window.navigator.browserLanguage).substr(0,2) == 'ja') {
    $('.lang-ja').show();
  } else {
    $('.lang-en').show();
  }

  // splash
  $('#splash').addClass('d-flex').show();
  // GNSS
  resetGNSS();
});

// Start button.
$('#start-btn').click(() => {
  nickname = $('[name="nickname"]').val();
  if (nickname === '') {
    nickname = 'anonymous';
  }
  $('#splash').remove();
  $('#main').show();

  // ネットワーク接続は、Startボタンを押してもらってから。
  // GNSSなどの位置情報は取得しても送信しなければ害にならないはず。
  resetVein();

  // 地図をX軸で回転(地図の要素が完全にロードされたタイミングで回転可能)
  $('.gm-style-pbc').parent().wrap('<div class="transform-parent">');
  $('.gm-style-pbc').parent().addClass('transform-target');

  relayout();
  $('#loading').show();
  // メインループ用タイマ
  setInterval(loop, 1000);
  // 表示用タイマ
  setInterval(rendMarkers, 500);
});

function loop() {
  let NOW = Math.floor(Date.now() / 1000);
  // libveinの状態確認
  if (veinStatus == STATUS_NG) {
    resetVein();
  }

  // GNSSの状態確認
  if (gnssStatus == STATUS_NG) {
    resetGNSS();
  }

  if (veinStatus == STATUS_OK && gnssStatus == STATUS_OK) {
    $('#loading').hide();

  } else {
    $('#loading').show();

    if (veinStatus == STATUS_OK) {
      $('#loading-vein').hide();
    } else {
      $('#loading-vein').show();
    }

    if (gnssStatus == STATUS_OK) {
      $('#loading-gnss').hide();
    } else {
      $('#loading-gnss').show();
    }
    return;
  }

  // 表示位置調整
  gmap.panTo(realLocation);

  // 自キャラの位置を計算
  // 初期値 or 現在位置から遠い場合、自分の近くの場所を起点に計算し直す
  let reset = false;
  if (dstLocation == null || getDistance(realLocation, dstLocation) > ROPE_RADIUS) {
    dstLocation = getRandomPoint(realLocation, ROPE_RADIUS);
    arrivalTime = NOW;
    reset = true;
  }
  if (arrivalTime <= NOW) {
    let srcLocation = dstLocation;
    // @todo 立ち止まる可能性
    dstLocation = getRandomPoint(srcLocation, STEP_RADIUS);
    while (getDistance(realLocation, dstLocation) > ROPE_RADIUS) {
      dstLocation = getRandomPoint(srcLocation, STEP_RADIUS);
    }

    // 現在のキャラクター位置を元にlibveinのネットワーク座標を更新
    let [x, y] = convertDeg2Rad(dstLocation);
    v.setPosition(x, y);
    
    // 5m/sくらいを想定
    let time = getDistance(srcLocation, dstLocation) / 5;
    arrivalTime = NOW + time;

    
    // 更新情報を送付
    let data = {
      id: localNid,
      srcLocation: srcLocation,
      dstLocation: dstLocation,
      nickname: nickname,
      reset: reset,
      time: time
    };
    pubsub2d.publish(CHANNEL_NAME, x, y, MESSAGE_RADIUS, JSON.stringify(data));
    updateMarker(data);
  }
}

// Setup/reset GNSS.
function resetGNSS() {
  let watchId = navigator.geolocation.watchPosition((position) => {
    // Good
    gnssStatus = STATUS_OK;
    realLocation.lng = position.coords.longitude;
    realLocation.lat = position.coords.latitude;
    console.log('update GNSS');

  }, (err) => {
    // Error
    gnssStatus = STATUS_NG;
    navigator.geolocation.clearWatch(watchId);

  }, {
	  "enableHighAccuracy": false,
	  "timeout": GNSS_TIMEOUT,
	  "maximumAge": 0,
  });
}

// Setup/reset libvein.
function resetVein() {
  v.init().then(() => {
    console.log('vein connect');
    // v.on('log', onGetLog);
    v.on('debug', onGetDebug);
    return v.connect('wss://www.oinari.app/vein/ws', '');

  }).then(() => {
    // Good
    veinStatus = STATUS_OK;

    localNid = v.getMyNid();

    pubsub2d = v.accessPubsub2D('pubsub2d');
    pubsub2d.on(CHANNEL_NAME, (data) => {
      updateMarker(JSON.parse(data));
    });

    console.log('vein success');

  }).catch((e) => {
    // Error
    veinStatus = STATUS_NG;

    console.error('vein failed');
    console.error(e);
  });
}

function initGMap() {
  gmap = new google.maps.Map(document.getElementById('map'), {
    center: {
      lat: Math.random() * 180.0 -  90.0,
      lng: Math.random() * 360.0 - 180.0
    },
    disableDefaultUI: true,
    draggable: false,
    mapTypeControl: false,
    zoomControl: true,
    zoom: 17 // 地図のズームを指定
  });

  gmap.addListener('bounds_changed', function() {
    rendMarkers();
  });
}

function updateMarker(data) {
  let NOW = Math.floor(Date.now() / 1000);
  let id = data.id;

  // 格納場所がない場合は作る
  if (!(id in markers)) {
    markers[id] = {
      id: id
    };
  }
  $.extend(markers[id], data);
  if ('image' in data) {
    markers[id].text = '';
    markers[id].image = data.image;
    markers[id].emitTime = NOW;
    addImageItem(markers[id]);
  }
  if ('text' in data) {
    markers[id].text = data.text;
    markers[id].image = null;
    markers[id].emitTime = NOW;
    addTextItem(markers[id]);
  }
  if ('dstLocation' in data) {
    markers[id].srcTime = NOW;
    markers[id].dstTime = NOW + data.time;
  }
}

function rendMarkers() {
  let NOW = Math.floor(Date.now() / 1000);
  let bounds = gmap.getBounds();

  for (let id of Object.keys(markers)) {
    let marker = markers[id];
    // 一定時間更新されない場合はキャラクタを削除
    if (marker.dstTime + TIMEOUT < NOW) {
      let $tag = $('#tag' + id);
      $tag.remove();
      delete markers[id];
      continue;
    }

    // 発言後、一定時間で吹き出しを削除
    if (marker.emitTime + EMIT_TIMEOUT < NOW) {
      marker.text = '';
      marker.image = null;
    }

    if (!('dstLocation' in marker)) continue;

    let p;
    let isWalk;

    // 場所の計算
    if (marker.dstTime < NOW) {
      p = marker.dstLocation;
      isWalk = false;

    } else {
      let r = (NOW - marker.srcTime) / (marker.dstTime - marker.srcTime);
      p = {
        lng: marker.srcLocation.lng + (marker.dstLocation.lng - marker.srcLocation.lng) * r,
        lat: marker.srcLocation.lat + (marker.dstLocation.lat - marker.srcLocation.lat) * r
      };
      isWalk = true;
    }

    if (bounds.contains(p)) {
      // 要素がないので新規作背
      if (!('tag' in marker) || marker.tag === false) {
        marker.tag = $('<div class="marker"><div class="balloon"><span class="text"></span>' +
                       '<img class="image"></img></div><div class="char fox1b pixelscaled"></div>' +
                       '<div class="nickname"></div></div>');
        marker.tag.attr('id', 'tag' + id);
        $('.transform-target').append(marker.tag);
      }

      let $tag = $('#tag' + id);
      let $balloon = $tag.find('.balloon');
      let $text    = $tag.find('.text');
      let $image   = $tag.find('.image');
      let $char    = $tag.find('.char');

      // 吹き出しの表示
      if ('text' in markers[id] && markers[id].text.trim() !== '') {
        $text.text(markers[id].text);
        $text.show();
        $image.hide();
        $balloon.show();

      } else if ('image' in markers[id] && markers[id].image !== null) {
        $image.attr('src', markers[id].image);
        $text.hide();
        $image.show();
        $balloon.show();

      } else {
        $balloon.hide();
      }

      // markerの位置調整
      let w = $('.transform-target').width();
      let x = 0;
      if (bounds.getNorthEast().lng() > bounds.getSouthWest().lng()) {
        x = w * (p.lng - bounds.getSouthWest().lng()) /
        (bounds.getNorthEast().lng() - bounds.getSouthWest().lng());
      } else {
        x = w * (360.0 - p.lng + bounds.getSouthWest().lng()) /
        (360.0 - bounds.getNorthEast().lng() + bounds.getSouthWest().lng());
      }
      // 1/sin(30) で補正の必要がある？
      x = (w * 0.5) + (x - w * 0.5) * 2 - (marker.tag.width() * 0.5);

      let h = $('.transform-target').height();
      let y = 0;
      y = h - h * (p.lat - bounds.getSouthWest().lat()) /
      (bounds.getNorthEast().lat() - bounds.getSouthWest().lat());
      // 1/cos(30)で補正の必要がある？
      y = (h * 0.5) + (y - h * 0.5) * 1.15 - marker.tag.height();
      markers[id].tag.css('left', x + 'px');
      markers[id].tag.css('top',  y + 'px');

      // 左右どちらを向かせるか
      if (marker.srcLocation.lng < marker.dstLocation.lng) {
        $char.removeClass('reverse');
      } else {
        $char.addClass('reverse');
      }

    } else {
      if ('tag' in marker && marker.tag !== false) {
        // marker要素がある場合は削除
        $(marker.tag).remove();
        marker.tag = false;      
      }
    }
  }
}

function addListItem(marker, $html) {
  let now = new Date();
  let $elem = $('<li class="list-group-item">' +
                '<div class="head"><span class="font-weight-bold nickname"></span>&nbsp;&nbsp;' +
                '<span class="text-secondary">' +
                ('0' + now.getHours()).slice(-2) + ':' + ('0' + now.getMinutes()).slice(-2) +
                '</span></div></li>');
  $elem.find('.head').append($html);
  $elem.find('.nickname').text(marker.nickname);
  $elem.prependTo('#list').hide().slideDown(400);
}

function addImageItem(marker) {
  let $html = $('<div class="card bg-dark"><img class="img-fluid mx-auto d-block"></img></div>');
  $html.find('img').attr('src', marker.image);
  addListItem(marker, $html);
}

function addTextItem(marker) {
  let $html = $('<div><span class="font-weight-normal"></span></div>');
  $html.find('span').text(marker.text);
  addListItem(marker, $html);
}

function getRandomPoint(c, r) {
  let rLat = 360 * r / O;
  let rLng = 2 * rLat;
  let p = {
    lng: c.lng + (Math.random() * rLng * 2) - rLng,
    lat: c.lat + (Math.random() * rLat * 2) - rLat
  };
  while (getDistance(c, p) > r) {
    p = {
      lng: c.lng + (Math.random() * rLng * 2) - rLng,
      lat: c.lat + (Math.random() * rLat * 2) - rLat
    };
  }
  return p;
}

// ２点間の距離[m]を求める
function getDistance(p1, p2) {
  let [x1, y1] = convertDeg2Rad(p1);
  let [x2, y2] = convertDeg2Rad(p2);
  let avrX = (x1 - x2) / 2;
  let avrY = (y1 - y2) / 2;
  
  return R * 2 * Math.asin(Math.sqrt(Math.pow(Math.sin(avrY), 2) + Math.cos(y1) *
    Math.cos(y2) * Math.pow(Math.sin(avrX), 2)));
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

    let nodes = event.content.nodes;
    for (let link of event.content.links) {
      let p1 = (link[0] == vein.Vein.NID_THIS ?
                dstLocation : convertRad2Deg(nodes[link[0]][0], nodes[link[0]][1]));
      let p2 = (link[1] == vein.Vein.NID_THIS ?
                dstLocation : convertRad2Deg(nodes[link[1]][0], nodes[link[1]][1]));
      let polyLine = new google.maps.Polyline({
        path: [p1, p2],
        strokeColor: '#666',
        strokeOpacity: 0.2,
        strokeWeight: 2
      });
      polyLine.setMap(gmap);
      gmLines.push(polyLine);
    }
  }
}

function convertDeg2Rad(p) {
  let lng = p.lng;
  let lat = p.lat;

  while (lat < -90.0) {
    lat += 360.0;
  }
  while (270.0 <= lat) {
    lat -= 360.0;
  }

  if (180.0 <= lat) {
    lng += 180.0;
    lat = -1.0 * (lat - 180.0);
  } else if (90.0 <= lat) {
    lng += 180.0;
    lat = 180.0 - lat;
  }

  while (lng < -180.0) {
    lng += 360.0;
  }
  while (180.0 <= lng) {
    lng -= 360.0;
  }
  return [Math.PI * lng / 180,
          Math.PI * lat / 180];
}

function convertRad2Deg(x, y) {
  return {
    lng: 180.0 * x / Math.PI,
    lat: 180.0 * y / Math.PI
  };
}

// 表示領域リサイズ時に地図の大きさなどを変更する
function relayout() {
  if (relayoutTimer != null) {
    clearTimeout(relayoutTimer);
  }
  
  relayoutTimer = setTimeout(function () {
    let fieldHeight = $(window).height() - $('footer').height();
    let fieldWidth  = $(window).width();
    let $map = $('#map');
    let $lists = $('#lists');

    if (fieldWidth > fieldHeight) {
      // 横向き
      let mapHeight = fieldHeight;
      let mapWidth  = fieldHeight;
      $map.height(mapHeight);
      $map.width (mapWidth);
      $lists.height(fieldHeight);
      $lists.width (fieldWidth - mapWidth);

    } else {
      // 縦向き
      let mapHeight = fieldWidth;
      let mapWidth  = fieldWidth;
      if (mapHeight > fieldHeight / 2) mapHeight = fieldHeight / 2;
      $map.height(mapHeight);
      $map.width (mapWidth);
      $lists.height(fieldHeight - mapHeight);
      $lists.width (fieldWidth);
    }

    relayoutTimer = null;
  }, 50);
}
$(window).on('load resize', relayout);

// ボタンを押したらカメラ起動
$('#btn-camera').on('click', function() {
  $('[name="capture"]').click();
});

// 画像が選択されたらダンプして送る
$('[name="capture"]').on('change', function() {
  let reader = new FileReader();

  // 画像をリサイズして送る
  // https://qiita.com/komakomako/items/8efd4184f6d7cf1363f2
  reader.onload = (e) => {
    let image = new Image();
    image.onload = () => {
      let width;
      let height;
      if (image.height <= PICTURE_SIZE && image.width <= PICTURE_SIZE) {
        // 十分に小さいのでサイズ変更はなし
        width = image.width;
        height = image.height;
      } else if(image.width > image.height){
        // 横長の画像は横のサイズを指定値にあわせる
        let r = image.height / image.width;
        width = PICTURE_SIZE;
        height = PICTURE_SIZE * r;
      } else {
        // 縦長の画像は縦のサイズを指定値にあわせる
        let r = image.width / image.height;
        width = PICTURE_SIZE * r;
        height = PICTURE_SIZE;
      }
      // サムネ描画用canvasのサイズを上で算出した値に変更
      let canvas = $('<canvas>')
        .attr('width', width)
        .attr('height', height);
      let ctx = canvas[0].getContext('2d');
      // canvasに既に描画されている画像をクリア
      ctx.clearRect(0, 0, width, height);
      // canvasに画像を描画
      ctx.drawImage(image, 0, 0, image.width, image.height, 0, 0, width, height);

      // canvasからbase64画像データを取得
      let base64 = canvas.get(0).toDataURL('image/jpeg');

      // 送信
      let data = {
        id: localNid,
        nickname: nickname,
        image: base64
      };
      let [x, y] = convertDeg2Rad(dstLocation);
      pubsub2d.publish(CHANNEL_NAME, x, y, MESSAGE_RADIUS, JSON.stringify(data));
      updateMarker(data);
    }
    image.src = e.target.result;
  };

  // 画像ファイルであった場合、読み込みを実行
  let file = this.files[0];
  if (file.type == 'image/jpeg' || file.type == 'image/png') {
    reader.readAsDataURL(file);
  }
});

// ボタンを押したらメッセージを送信
$('#btn-text').on('click', function() {
  let $message = $('#message');
  let text = $message.val().trim();
  if (text != '') {
    let data = {
      id: localNid,
      nickname: nickname,
      text: $message.val()
    };
    let [x, y] = convertDeg2Rad(dstLocation);
    pubsub2d.publish(CHANNEL_NAME, x, y, MESSAGE_RADIUS, JSON.stringify(data));
    updateMarker(data);
  }
  $message.val('');
});
