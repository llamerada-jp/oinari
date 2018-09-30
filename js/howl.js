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
let mapDefer = $.Deferred();
let index = 0;
let markers = [];
let gmLines = [];
let gmCircle = null;

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

function recvHowl(data) {
  console.log('recv ' + JSON.stringify(data));
  let message = JSON.parse(data);

  let indexThis = index;
  index ++;

  let marker = new google.maps.Marker({
    icon: {
      anchor: new google.maps.Point(16, 32),
      scaledSize: new google.maps.Size(32, 32),
      url:'img/h1.png'
    },
    map: map,
    position: { lat: message.lat, lng: message.lon }
  });

  marker.addListener('click', function() {
    $('#tag' + indexThis).focus();
  });

  markers[indexThis] = marker;

  $('<tr id="' + indexThis + '">' +
    '<td><div style="display:none;">' + getDistance(message.lon, message.lat).toFixed(1) + 'm</div></td>' +
    '<td><div style="display:none;">' + message.message + '</div></td>' +
    '<td><div style="display:none;"><button type="button" class="btn btn-primary">Howl</button></div></td>' + 
    '</tr>').prependTo('#list tbody').find('td > div').slideDown(400, function() {
      let $this = $(this);
      $this.replaceWith($this.contents());
    });
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
    if (wasGNSS === true) {
      let [x, y] = convertDeg2Rad(lon, lat);
      v.setPosition(x, y);
    }

    pubsub2d = v.accessPubsub2D('pubsub2d');
    pubsub2d.on('howl', recvHowl);

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
  console.log(log);
}

function onGetDebug(event) {
  console.log(event);

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
    draggable: false,
    mapTypeControl: false,
    zoom: 18 // 地図のズームを指定
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

$(window).on('load resize', () => {
  let fieldHeight = $(window).height() - $('header').height() - $('footer').height();
  let fieldWidth  = $(window).width();
  let $map = $('#map');
  let $lists = $('#lists');

  if (fieldWidth > fieldHeight) {
    $map.height(fieldHeight);
    $map.width(fieldHeight);
    $lists.height(fieldHeight);
    $lists.width(fieldWidth - fieldHeight);

  } else {
    $map.height(fieldWidth);
    $map.width(fieldWidth);
    $lists.height(fieldHeight - fieldWidth);
    $lists.width(fieldWidth);
  }
});

$(window).on('load', () => {
  init();
});

$('#btn-howl').on('click', function() {
  let $message = $('#message');
  let message = {
    message: $message.val(),
    lon: lon,
    lat: lat
  };
  let [x, y] = convertDeg2Rad(lon, lat);
  pubsub2d.publish('howl', x, y, RADIUS, JSON.stringify(message));
  $message.val('');
});
