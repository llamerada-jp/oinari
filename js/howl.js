"use strict";

let v = new vein.Vein();
let pubsub2d = null;
let lat = null;
let lon = null;
let wasConnect = false;
let wasGNSS    = false;
let wasStart   = false;
let wasMap     = false;
let geoWatchId = null;
let map = null;
let mapDefer = $.Deferred();

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
    $('#modal-init').modal('hide');
  });
}

function initGNSS() {
  let defer = $.Deferred();
  let isFirst = true;
  geoWatchId = navigator.geolocation.watchPosition((position) => {
    lon = position.coords.longitude;
    lat = position.coords.latitude;
    if (wasConnect === true) {
      let [x, y] = convertLonLat2Rad(lon, lat);
      v.setPosition(x, y);
    }
    if (wasMap === true) {
      map.panTo({lat: lat, lng: lon});
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
  $('<tr>' +
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
  let [x1, y1] = convertLonLat2Rad(lon, lat);
  let [x2, y2] = convertLonLat2Rad(bLon, bLat);
  let avrX = (x1 - x2) / 2;
  let avrY = (y1 - y2) / 2;
  
  return R * 2 * Math.asin(Math.sqrt(Math.pow(Math.sin(avrY), 2) + Math.cos(y1) *
    Math.cos(y2) * Math.pow(Math.sin(avrX), 2)));
}

function initNet() {
  let defer = $.Deferred();
  v.init().then(() => {
    console.log('vein connect');
    return v.connect('http://veindev:8080/vein/core.json', '');

  }).then(() => {
    if (wasGNSS === true) {
      let [x, y] = convertLonLat2Rad(lon, lat);
      v.setPosition(x, y);
    }

    pubsub2d = v.accessPubsub2D('pubsub2d');
    pubsub2d.on('howl', recvHowl);

    wasConnect = true;
    console.log('vein success');
    $('#status-net').html('<span class="badge badge-success">Success</span>');
    defer.resolve();

  }).catch(() => {
    console.log('vein failed');
    $('#status-net').html('<span class="badge badge-danger">Failed</span>');
    defer.reject();
  });
  return defer.promise();
}

function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    center: {
      lat: Math.random() * 180.0 -  90.0,
      lng: Math.random() * 360.0 - 180.0
    },
    zoom: 19 // 地図のズームを指定
  });

  wasMap = true;
  $('#status-map').html('<span class="badge badge-success">Success</span>');
  mapDefer.resolve();
}

function wait(sec) {
  let defer = $.Deferred();
  setTimeout(() => {
    defer.resolve();
  }, sec * 1000);
  return defer.promise();
}

function convertLonLat2Rad(lon, lat) {
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

$(window).on('load resize', () => {
  let fieldHeight = $(window).height() - $('#footer').height();
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
  let [x, y] = convertLonLat2Rad(lon, lat);
  pubsub2d.publish('howl', x, y, 100, JSON.stringify(message));
  $message.val('');
});
