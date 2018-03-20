var cachedPos = {
    lat: 0, lng: 0,
    kmPerLat: 111.1,
    kmPerLng: 111.1
};

var documentWasHiddenP = false;

var locs = {};
var locsByBox = {};
function locs_addBoxes(boxes) {
    for (box in boxes) {
	locsByBox[box] = boxes[box];
    }
    locs_updateFromByBox();
    try {
	localStorage.locsByBox = JSON.stringify(locsByBox);
    } finally {}
}
function locs_forget() {
    if (Object.keys(locs).length < 1000) { return }
    for (k in locsByBox) {
	if (locsByBox[k].length < 1 || Math.random() < 0.5) {
	    delete(locsByBox[k]);
	}
    }
    locs_updateFromByBox();
}
function locs_updateFromByBox() {
    locs = {};
    for (box in locsByBox) {
	for (var ix = 0; ix < locsByBox[box].length; ix++) {
	    locs[locsByBox[box][ix].id] = locsByBox[box][ix];
	}
    }
}

// We pre-compute "grid coords" -- regions' offsets from the user's position.
// Re-figure these when the user moves or when we have new region info.
function locs_updateGridCoords() {
    // update gridX, gridY, dist
    for (lid in locs) {
	locs[lid].gridXKm = (locs[lid].lng - cachedPos.lng) * cachedPos.kmPerLng;
	locs[lid].gridYKm = (locs[lid].lat - cachedPos.lat) * cachedPos.kmPerLat;
	locs[lid].dist = distanceGrid(0, 0, locs[lid].gridXKm, locs[lid].gridYKm)
    }

    // update radius
    for (lid in locs) {
	locs[lid].radiusKm = 1.0;
    }
    for (id0 in locs) {
	if (locs[id0].dist > 2.5) { continue; }
	for (id1 in locs) {
	    if (locs[id1].dist > 2.5) { continue; }
	    if (id0 == id1) { continue; }
	    d = distanceGrid(locs[id0].gridXKm, locs[id0].gridYKm, locs[id1].gridXKm, locs[id1].gridYKm);
	    if (d < locs[id0].radiusKm) { locs[id0].radiusKm = d; }
	    if (d < locs[id1].radiusKm) { locs[id1].radiusKm = d; }
	}
    }
}

var knownRts = {};
function knownRts_forget() {
    if (Object.keys(knownRts).length < 1000) { return }
    for (k in knownRts) {
	if (Math.random() < 0.5) {
	    delete(knownRts[k]);
	    continue
	}
	if (!locs[knownRts[k].ends[0]]) {
	    delete(knownRts[k]);
	    continue
	}
	if (!locs[knownRts[k].ends[1]]) {
	    delete(knownRts[k]);
	    continue
	}
    }
}

var prisms = {};
var bounties = {};
var npcs = {};
var coins = 0;
var trophies = 0;
var cred = 0;

var particles = {}
var particlesP = false;
var particlesPrevTick = 0
function particles_add(ps) {
    particlesP = true;
    while (true) {
	var pix = '' + Math.floor(1000000 * Math.random());
	if (!particles[pix]) {
	    particles[pix] = ps;
	    return
	}
    }
}
function particles_tick() {
    var now = Date.now();
    if (now <= particlesPrevTick) {
	return
    }
    var deltaT = (now - particlesPrevTick) / 1000;
    particlesPrevTick = now;
    drawMap();
    for (pix in particles) {
	p = particles[pix];
	if (p.doneP()) {
	    delete(particles[pix]);
	    if (!Object.keys(particles).length) {
		particlesP = false;
		return
	    }
	    continue
	}
	p.tick(deltaT);
    }
}

var textures = {};
var i = new Image();
i.src = '/client/fadey_circle.png';
i.onload = function() {
  textures.fadey_circle = i;
}
var icons = new Image();
icons.src = '/client/icons.png';
icons.onload = function() {
    textures.icons = icons;
}
iconCenters = {
    'prism-have': { x: 0.125, y: 0.125 },
    'prism-want': { x: 0.375, y: 0.125 },
    'chekn-done': { x: 0.125, y: 0.375 },
    'chekn-nyet': { x: 0.375, y: 0.375 },
    'trophy':     { x: 0.125, y: 0.625 },
    'npc-have':   { x: 0.125, y: 0.875 },
    'npc-theirs': { x: 0.375, y: 0.875 },
}

var nearRegionID = '';
var tappedRegionID = '';

var shaderPrograms = {};

var messages = ['<i>Intializing&hellip;</i>'];
function messages_add(htmlStringOrListOfHtmlStrings) {
    if (typeof(htmlStringOrListOfHtmlStrings) == "string") {
	htmlStringOrListOfHtmlStrings = [htmlStringOrListOfHtmlStrings];
    }
    messages = htmlStringOrListOfHtmlStrings.concat(messages);
}
function messages_forget() {
    if (messages.length > 20) {
	messages.length = Math.floor( messages.length / 2);
    }
}

var diary = [];
function diary_add(i) {
    diary.unshift(i);
    try {
	localStorage.diary = JSON.stringify(diary);
    } finally{}
}
function diary_recent(c) {
    for (ix = 0; ix < diary.length; ix++) {
	if (diary[ix].category && diary[ix].category == c) {
	    return diary[ix];
	}
    }
    return null
}
function diary_forget() {
    if (diary.length > 500) {
	diary.length = Math.floor(diary.length / 2);
    }
}

var probes = [];
function probes_add(i) {
    probes.unshift(i);
}
function probes_forget() {
    if (len(probes) > 30) {
	probes.length = Math.floor(probes.length / 2);
    }
}
function probes_tooCloseP(lat, lng) {
    for (var ix = 0; ix < probes.length; ix++) {
	if (distanceLL(lat, lng, probes[ix].lat, probes[ix].lng) < 0.5) {
	    return true;
	}
    }
    return false;
}

var checkins = {};
checkins_add = function(id) {
    checkins[id] = Date.now()
}
checkins_forget = function() {
    var hourAgo = Date.now() - (60 * 60 * 1000);
    for (k in checkins) {
	if (checkins[k] < hourAgo) {
	    delete(checkins[k]);
	}
    }
}

function forget() {
    locs_forget();
    knownRts_forget();
    messages_forget();
    diary_forget();
    checkins_forget();
}

// km betwen two lat/lng locs
function distanceLL(lat1, lng1, lat2, lng2) {
    var dXKm = (lng1 - lng2) * cachedPos.kmPerLng;
    var dYKm = (lat1 - lat2) * cachedPos.kmPerLat;
    return Math.sqrt((dXKm * dXKm) + (dYKm * dYKm));
}

function distanceGrid(x1, y1, x2, y2) {
    var dX = x1 - x2;
    var dY = y1 - y2;
    return Math.sqrt((dX * dX) + (dY * dY));
}

function ingestCheckins(cs) {
    for (var ix = 0; ix < cs.length; ix++) {
	checkins_add(cs[ix]);
	var loc = locs[cs[ix]];
	if (!loc) {
	    // We checked into a location we didn't know about.
	    // It was probably created just seconds ago.
	    // This would be a good time to fetch the latest loc data:
	    setTimeout(pace, 1);
	}
	if (loc && loc.lat && loc.lng) {
	    diary_add({
		category: 'checkin',
		lat: loc.lat,
		lng: loc.lng,
	    });
	    particles_add(new CheckinParticleSystem(loc));
	}
    }
    try {
	localStorage.checkins = JSON.stringify(checkins);
    } finally{}
}

function ingest(j) { // crunch the data we got from the server
    if (j.chkn) {
	ingestCheckins(j.chkn);
    }
    if (j.regs) {
	ingestRegs(j.regs);
    }
    if (j.orts) {
	ingestRoutes(j.orts);
    }
    if (j.nrts) {
	ingestRoutes(j.nrts);
	for (var ix = 0; ix < j.nrts.length; ix++) {
	    particles_add(new NewRouteParticleSystem(j.nrts[ix]));
	}
    }
    if (j.msgs) {
	ingestMessages(j.msgs);
    }
    if (j.inv) {
	ingestInventory(j.inv);
    }
    if (j.npcs) {
	ingestNPCs(j.npcs);
    }
}

function ingestMessages(msgs) {
    messages_add(msgs);
    var msghtml = '';
    for (var ix = 0; ix < messages.length; ix++) {
	msghtml = msghtml + '<div class="msg">MSG</div>'.replace(/MSG/, messages[ix]);
    }
    $('#messages').html(msghtml)
}

function ingestRegs(newregs) {
    locs_addBoxes(newregs);
    locs_updateGridCoords();
}

function ingestInventory(i) {
    if (!i) {
	return
    }
    if (i.prisms) {
	prisms = {};
	for (var ix = 0; ix < i.prisms.length; ix++) {
	    if (prisms[i.prisms[ix]]) {
		prisms[i.prisms[ix]] += 1;
	    } else {
		prisms[i.prisms[ix]] = 1;
	    }
	}
    }
    if (i.bounties) {
	bounties = {}
	for (var ix = 0; ix < i.bounties.length; ix++) {
	    bounties[i.bounties[ix]] = true;
	}
	localStorage.bounties = JSON.stringify(bounties);
    }
    coins = i.coins;
    cred = i.cred;
    trophies = i.trophies;
}

function ingestNPCs(l) {
    npcs = {};
    for (nix = 0; nix < l.length; nix++) {
	npc = l[nix];
	npcs[npc.reg] = npc;
    }
}

function ingestRoutes(routes) {
    for (var ix = 0; ix < routes.length; ix++) {
	knownRts[ JSON.stringify(routes[ix]) ] = routes[ix]
    }
}

function maybeProbe() {
    var predict = predictDirection();

    var a = Math.random();
    var lat = cachedPos.lat + a * predict.dLat;
    var lng = cachedPos.lng + a * predict.dLng;
    while (probes_tooCloseP(lat, lng)) {
	a += Math.random();
	lat = cachedPos.lat + a * predict.dLat;
	lng = cachedPos.lng + a * predict.dLng;
    }
    if (a > 3) {
	return
    }
    probes_add({
	lat: lat,
	lng: lng,
    });

    for (lid in locs) {
	var loc = locs[lid];
	if (distanceLL(lat, lng, loc.lat, loc.lng) < 0.5) {
	    return
	}
    }

    $.ajax({url: '/a/probe', data: {
	lat: lat,
	lng: lng,
    }, type: 'GET', dataType: 'json'}).done(function(j) {
	ingest(j);
    })
}

function pace() {
    diary_add({
	category: 'pace',
	lat: cachedPos.lat,
	lng: cachedPos.lng,
	timestamp: Date.now(),
    });
    probes_add({
	lat: cachedPos.lat,
	lng: cachedPos.lng,
    });
    var retries = 3;
    helper = function() {
	$.ajax({url: '/a/pace', data: {
	    lat: cachedPos.lat,
	    lng: cachedPos.lng
	}, type: 'GET', dataType: 'json'}).done(function(j) {
	    forget();
	    ingest(j);
	    drawMap();
	    maybeProbe();
	}).fail(function(j, status, errorThrown) {
	    drawMap();
	    if (retries > 0) {
		retries--;
		helper();
	    } else {
		if (errorThrown) {
		    newMsg('↯ Trouble communicating with server, got error "ERR". Tried three times, giving up.'.replace(/ERR/, errorThrown));
		} else {
		    newMsg('↯ Trouble communicating with server. Tried three times, giving up.');
		}
	    }
	});
    }
    helper();
}

var gl = {}; // webGL context
var glv = {}; // GL variables holder

function getPolyShaderProgram(gl) {
  var fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);
  gl.shaderSource(fragmentShader, 'precision mediump float; uniform vec4 poly_color; void main() { gl_FragColor = poly_color; }');
  gl.compileShader(fragmentShader);
  if (!gl.getShaderParameter(fragmentShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling frag shader: ' + gl.getShaderInfoLog(fragmentShader));
  }
  var vertexShader = gl.createShader(gl.VERTEX_SHADER);
  gl.shaderSource(vertexShader, 'attribute vec2 poly_vert_pos; void main() { gl_Position = vec4(poly_vert_pos, 0, 1); }');
  gl.compileShader(vertexShader);
  if (!gl.getShaderParameter(vertexShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling vert shader: ' + gl.getShaderInfoLog(vertexShader));  
  }
  
  // Create the shader program
  
  polyShaderProgram = gl.createProgram();
  gl.attachShader(polyShaderProgram, vertexShader);
  gl.attachShader(polyShaderProgram, fragmentShader);
  gl.linkProgram(polyShaderProgram);
  
  // If creating the shader program failed, alert
  if (!gl.getProgramParameter(polyShaderProgram, gl.LINK_STATUS)) {
    alert('Unable to initialize the shader program: ' + gl.getProgramInfoLog(shader));
  }
  
  // variables "exported" by program
  glv.poly_vert_pos = gl.getAttribLocation(polyShaderProgram, 'poly_vert_pos');
  glv.poly_color = gl.getUniformLocation(polyShaderProgram, 'poly_color');

  return polyShaderProgram;
}

function getFadeyCircleShaderProgram(gl) {
  var fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);
  gl.shaderSource(fragmentShader, 'precision mediump float; uniform vec3 circle_color; uniform sampler2D tex; varying vec2 circle_tex_pos; void main() { gl_FragColor = vec4(circle_color, texture2D(tex, abs(circle_tex_pos)).r ); }');
  gl.compileShader(fragmentShader);
  if (!gl.getShaderParameter(fragmentShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling frag shader: ' + gl.getShaderInfoLog(fragmentShader));
  }
  var vertexShader = gl.createShader(gl.VERTEX_SHADER);
  gl.shaderSource(vertexShader, 'attribute vec2 circle_pos; uniform vec2 circle_center_pos; uniform float circle_radius; varying vec2 circle_tex_pos; void main() { gl_Position = vec4(circle_center_pos + circle_radius * circle_pos, 0, 1); circle_tex_pos = circle_pos; }');
  gl.compileShader(vertexShader);
  if (!gl.getShaderParameter(vertexShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling vert shader: ' + gl.getShaderInfoLog(vertexShader));  
  }
  
  // Create the shader program
  
  fadeyCircleShaderProgram = gl.createProgram();
  gl.attachShader(fadeyCircleShaderProgram, vertexShader);
  gl.attachShader(fadeyCircleShaderProgram, fragmentShader);
  gl.linkProgram(fadeyCircleShaderProgram);
  
  // If creating the shader program failed, alert
  if (!gl.getProgramParameter(fadeyCircleShaderProgram, gl.LINK_STATUS)) {
    alert('Unable to initialize the shader program: ' + gl.getProgramInfoLog(fadeyCircleShaderProgram));
  }

  // variables "exported" by program
  glv.circle_center_pos = gl.getUniformLocation(fadeyCircleShaderProgram, 'circle_center_pos');
  glv.circle_color = gl.getUniformLocation(fadeyCircleShaderProgram, 'circle_color');
  glv.circle_radius = gl.getUniformLocation(fadeyCircleShaderProgram, 'circle_radius');
  glv.circle_pos = gl.getAttribLocation(fadeyCircleShaderProgram, 'circle_pos');

    return fadeyCircleShaderProgram;
}

// draw one icon from a grid of icons.
function getAlphaIconShaderProgram(gl) {
  // varying vec2 tex_pos : position within the icon grid
  var fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);
  gl.shaderSource(fragmentShader, 'precision mediump float; uniform sampler2D tex; varying vec2 tex_pos; void main() { gl_FragColor = texture2D(tex, tex_pos); }');
  gl.compileShader(fragmentShader);
  if (!gl.getShaderParameter(fragmentShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling frag shader: ' + gl.getShaderInfoLog(fragmentShader));
  }
  var vertexShader = gl.createShader(gl.VERTEX_SHADER);
  gl.shaderSource(vertexShader, 'attribute vec2 pos; uniform vec2 grid_center_pos; uniform float icon_radius; uniform vec2 tex_center; varying vec2 tex_pos; void main() { gl_Position = vec4(grid_center_pos + icon_radius * pos, 0, 1); tex_pos.x = tex_center.x + 0.125 * pos.x; tex_pos.y = tex_center.y - 0.125 * pos.y; }');
  gl.compileShader(vertexShader);
  if (!gl.getShaderParameter(vertexShader, gl.COMPILE_STATUS)) {  
      alert('An error occurred compiling vert shader: ' + gl.getShaderInfoLog(vertexShader));  
  }
  
  // Create the shader program
  alphaIconShaderProgram = gl.createProgram();
  gl.attachShader(alphaIconShaderProgram, vertexShader);
  gl.attachShader(alphaIconShaderProgram, fragmentShader);
  gl.linkProgram(alphaIconShaderProgram);
  
  // If creating the shader program failed, alert
  if (!gl.getProgramParameter(alphaIconShaderProgram, gl.LINK_STATUS)) {
    alert('Unable to initialize the shader program: ' + gl.getProgramInfoLog(alphaIconShaderProgram));
  }

  // variables "exported" by program
    glv.grid_center_pos = gl.getUniformLocation(alphaIconShaderProgram, 'grid_center_pos');
    glv.icon_radius = gl.getUniformLocation(alphaIconShaderProgram, 'icon_radius');
    glv.tex_center = gl.getUniformLocation(alphaIconShaderProgram, 'tex_center');
    glv.pos = gl.getUniformLocation(alphaIconShaderProgram, 'pos');

    return alphaIconShaderProgram;
}
  

function initGL() {
  var canvas = document.getElementById('glcanvas');

  // Initialize the GL context
  gl = canvas.getContext('webgl');
  
  shaderPrograms.poly = getPolyShaderProgram(gl);
  shaderPrograms.fadeyCircle = getFadeyCircleShaderProgram(gl);
  shaderPrograms.icon = getAlphaIconShaderProgram(gl);

  gl.enable(gl.BLEND);
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
}

function drawMapRoutes(gl) {
    gl.useProgram(shaderPrograms.poly);

    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.poly_vert_pos);
    gl.vertexAttribPointer(glv.poly_vert_pos, 2, gl.FLOAT, false, 0, 0);

    gl.uniform4f(glv.poly_color, 0.4, 0.4, 0.9, 1);
    gl.lineWidth(0.5);

    for (k in knownRts) {
	ends = knownRts[k].ends;
	if (! locs[ends[0]]) { continue; }
	if (! locs[ends[1]]) { continue; }
	
	
	gl.bufferData(
            gl.ARRAY_BUFFER,
            new Float32Array([
		locs[ends[0]].gridXKm, locs[ends[0]].gridYKm, 
		locs[ends[1]].gridXKm, locs[ends[1]].gridYKm
            ]),
            gl.STATIC_DRAW);
	gl.drawArrays(gl.LINES, 0, 2);
    }    
    
}

/* If we have bounty/bounties, highlight them with golden lines */
function drawMapBountyLines(gl) {
    if (!bounties) {
	return;
    }
    gl.useProgram(shaderPrograms.poly);

    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.poly_vert_pos);
    gl.vertexAttribPointer(glv.poly_vert_pos, 2, gl.FLOAT, false, 0, 0);

    gl.uniform4f(glv.poly_color, 0.9, 0.9, 0.1, 1);
    gl.lineWidth(1.5)

    for (rid in bounties) {
	var r = locs[rid]
	if (!r) { continue }

	gl.bufferData(
            gl.ARRAY_BUFFER,
            new Float32Array([
		0, 0,
		r.gridXKm, r.gridYKm
            ]),
            gl.STATIC_DRAW);
	gl.drawArrays(gl.LINES, 0, 2);
    }
  
}

function drawMapGrid(gl) {
    gl.useProgram(shaderPrograms.poly);    

    // Create a buffer to hold our lovely triangles
    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.poly_vert_pos);
    gl.vertexAttribPointer(glv.poly_vert_pos, 2, gl.FLOAT, false, 0, 0);

    gl.bufferData(
        gl.ARRAY_BUFFER,
        new Float32Array([
                -0.05,  0.00,
            0.00,  0.05,
            0.00, -0.05,

            0.05,  0.00,
            0.00, -0.05,
            0.00,  0.05,
        ]),
        gl.STATIC_DRAW);
    gl.uniform4f(glv.poly_color, 0.0, 0.5, 0.0, 1);
    gl.drawArrays(gl.TRIANGLES, 0, 6);

    gl.bufferData(
        gl.ARRAY_BUFFER,
        new Float32Array([
                -2.0,  0.0,
            2.0,  0.0,

            0.0,  2.0,
            0.0, -2.0,
        ]),
        gl.STATIC_DRAW);
    gl.uniform4f(glv.poly_color, 0.0, 0.5, 0.0, 1);
    gl.lineWidth(1.0)
    gl.drawArrays(gl.LINES, 0, 4);

    gl.bufferData(
        gl.ARRAY_BUFFER,
        new Float32Array([
                -2.0,  0.5,
            2.0,  0.5,

                -2.0, -0.5,
            2.0, -0.5,

            0.5,  2.0,
            0.5, -2.0,

                -0.5,  2.0,
                -0.5, -2.0,
        ]),
        gl.STATIC_DRAW);
    gl.uniform4f(glv.poly_color, 0.0, 0.5, 0.0, 1);
    gl.lineWidth(0.5)
    gl.drawArrays(gl.LINES, 0, 8);
}

function drawMapIcons(gl) {
    if (!textures.icons) { return; }
    gl.useProgram(shaderPrograms.icon);

    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.circle_pos);
    gl.vertexAttribPointer(glv.circle_pos, 2, gl.FLOAT, false, 0, 0);
    
    var texture = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, textures.icons);

    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

    gl.bufferData(
        gl.ARRAY_BUFFER,
	new Float32Array([
		-1.0, -1.0,
		-1.0,  1.0,
	        1.0, -1.0,

                1.0,  1.0,
                1.0, -1.0,
                -1.0,  1.0
	]),
        gl.STATIC_DRAW);
    gl.uniform1f(glv.icon_radius, 0.05);

    for (lid in locs) {
	loc = locs[lid];

	if (Math.abs(loc.gridXKm) > 1.1) { continue; }
	if (Math.abs(loc.gridYKm) > 1.1) { continue; }

	var ic = iconCenters['prism-want'];
	if (checkins[lid]) {
	    ic = iconCenters['chekn-done'];
	} else if (bounties[lid]) {
	    ic = iconCenters['trophy'];
	} else if (npcs[lid] && npcs[lid].yrs) {
	    ic = iconCenters['npc-have'];
	} else if (npcs[lid] && !npcs[lid].yrs) {
	    ic = iconCenters['npc-theirs'];
	} else if (prisms[lid]) {
	    ic = iconCenters['prism-have'];
	}
	gl.uniform2f(glv.grid_center_pos, loc.gridXKm, loc.gridYKm);
	gl.uniform2f(glv.tex_center, ic.x, ic.y);
	gl.drawArrays(gl.TRIANGLES, 0, 6);
    }
}

function locDom(loc) {
      var r = Math.floor(loc.color[0] * 255.0);
      var g = Math.floor(loc.color[1] * 255.0);
      var b = Math.floor(loc.color[2] * 255.0);
      var rgb = '' + r + ',' + g + ',' + b;
      var display_d = Math.round(loc.dist * 100) / 100
      var dom = $('<div><span class="regcolor">■</span><span class="regname">NOM</span>&nbsp;<span class="regdist">D?</span>km</div>');
      dom.find('span.regcolor').css({'color': 'rgb(RGB)'.replace(/RGB/, rgb)});
      dom.find('span.regname').text(loc.name);
    dom.find('span.regdist').text(display_d);

    return dom
}

function drawMapRegions(gl) {
    
  if (!textures.fadey_circle) { return; }
  if (!Object.keys(locs).length) { return; }

  gl.useProgram(shaderPrograms.fadeyCircle);

  // Create a buffer to hold our lovely triangles
  var buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.enableVertexAttribArray(glv.circle_pos);
  gl.vertexAttribPointer(glv.circle_pos, 2, gl.FLOAT, false, 0, 0);

  var texture = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, texture);
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, textures.fadey_circle);

  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);

  var distanceSortedList = []

    for (loc in locs) {
	pos = locs[loc];
	if (pos.dist - pos.radiusKm > 1.5) { continue; }
	distanceSortedList.push(pos);
  }
  distanceSortedList.sort(function(p1, p2) {return p2.dist - p1.dist});
  for (var ix = 0; ix < distanceSortedList.length; ix++) {
      pos = distanceSortedList[ix];
	
      gl.bufferData(
          gl.ARRAY_BUFFER,
	  new Float32Array([
	      -1.0, -1.0,
	      -1.0,  1.0,
	       1.0, -1.0,

               1.0,  1.0,
               1.0, -1.0,
	      -1.0,  1.0
	  ]),
          gl.STATIC_DRAW);
      gl.uniform3f(glv.circle_color, pos.color[0], pos.color[1], pos.color[2]);
      gl.uniform2f(glv.circle_center_pos, pos.gridXKm, pos.gridYKm);
      
      gl.uniform1f(glv.circle_radius, pos.radiusKm);
      gl.drawArrays(gl.TRIANGLES, 0, 6);
  }
    if (distanceSortedList.length > 0 && distanceSortedList[distanceSortedList.length-1].dist < 1.0) {
	closestLoc = distanceSortedList[distanceSortedList.length-1]
	nearRegionID = closestLoc.id;
	var fsqURL = 'https://foursquare.com/explore?mode=url&ll=' + closestLoc.lat + ',' + closestLoc.lng;
	if (closestLoc.fsq) {
	    fsqURL = closestLoc.fsq;
	}
	var pingdom = $('<div><a id="fsq" class="ui-corner-all">4sq</a><span class="reg"></span></div>');
	if (fsqURL.includes('wikipedia')) {
	    pingdom = $('<div><a id="fsq" class="ui-corner-all">Wi</a><span class="reg"></span></div>');
	}
	pingdom.find('a#fsq').attr('href', fsqURL);
	pingdom.find('span.reg').append(locDom(closestLoc));
	$('#ping').empty().append(pingdom);
    } else {
	nearRegionID = '';
	var fsqURL = 'https://foursquare.com/explore?mode=url&ll=' + cachedPos.lat + ',' + cachedPos.lng;
	$('#ping').html('<div><a id="fsq" class="ui-corner-all" href="' + fsqURL + '">4sq</a>Ping</div>');
    }
    if ((!locs[tappedRegionID]) ||
	locs[tappedRegionID].dist > 5.0 ||
	tappedRegionID == nearRegionID) {
	tappedRegionID = '';
    }
    
    if (tappedRegionID == '') {
	if (nearRegionID) {
	    $('#gazetteer').html('☛ Tap the map to show info about tapped region here');
	} else {
	    $('#gazetteer').html('⏳Hang out in some Wikipedia-known neighborhood several minutes for a more interesting map⏳');
	}
    } else {
	$('#gazetteer').html(locDom(locs[tappedRegionID]));
    }
}

function drawMap() {
  // Clear the color
  gl.clearColor(0.5, 0.6, 0.4, 1.0);
  gl.clear(gl.COLOR_BUFFER_BIT);

  drawMapRegions(gl);
  drawMapRoutes(gl);
  drawMapIcons(gl);
  drawMapBountyLines(gl);
  drawMapGrid(gl);
}

function newMsg(html) {
    ingest({msgs: [html]});
}

function pingClick(e) {
    // gray out button a couple of seconds, thus discouraging race conditions.
    $('#ping').prop("disabled", true);
    setTimeout(function() {
	$('#ping').prop("disabled", false);
    }, 2 * 1000);
    
    var retries = 3;
    var token = '' + Math.floor(10000000 * Math.random());
    helper = function() {
	$.ajax({
	    url: '/a/checkin?' + $.param({
		lat: cachedPos.lat,
		lng: cachedPos.lng,
		token: token
	    }),
	    method: 'POST',
	    dataType: 'json',
	    cache: false}).done(function(j) {
		checkins_forget();
		ingest(j);
		drawMap();
		window.setTimeout(fetchCurPos, 10 * 1000)
	    }).fail(function(j, status, errorThrown) {
		if (retries > 0) {
		    window.setTimeout(helper, 1 * 1000)
		} else {
		    var dbgHelp = '' + status;
		    if (errorThrown && errorThrown.message) {
			dbgHelp = ' .message=' + errorThrown.message;
		    } else if (errorThrown && typeof(errorThrown)=='string') {
			dbgHelp = ' errorThrown(string)=' + errorThrown;
		    } else if (errorThrown) {
			dbgHelp = ' typeof(errorThrown)=' + typeof(errorThrown);
			for (f in errorThrown) {
			    dbgHelp += ' ' + f;
			}
		    };
		    if (dbgHelp) {
			newMsg('↯ Trouble communicating with server, got error "ERR". Tried three times, giving up.'.replace(/ERR/, dbgHelp));
		    } else {
			newMsg('↯ Trouble communicating with server. Tried three times, giving up.');
		    }
		}
	    });
    }
    helper();
}

function receivedCurPos(pos) {
    cachedPos = {
	lat: pos.coords.latitude,
	lng: pos.coords.longitude,
	kmPerLat: 111.1,
	kmPerLng: 111.1 * Math.cos(pos.coords.latitude * 3.14159 / 180.0),
    }
    locs_updateGridCoords();
    diary_add({
	category: 'curpos-recv',
	lat: pos.coords.latitude,
	lng: pos.coords.longitude,
    });

    // We just got new location. Maybe we should ask the server for
    // info about our surroundings?
    // If we've moved a ways. Or if it's been a while.
    recentPace = diary_recent('pace');
    if ((!recentPace) ||
	distanceLL(pos.coords.latitude,
		   pos.coords.longitude,
		   recentPace.lat, recentPace.lng) > 0.1 ||
	Date.now() - recentPace.timestamp > 5 * 60 * 1000) {
	pace();
    } else {
	drawMap();
    }
}

var botGoalLat = 34.072;
var botGoalLng = -118.292
var botGoalStrategy = 'orbit';
var botGoalTooFarKm = 5.0;
var botPrevReg = '';
function setBotGoalLL(lat, lng) {
    botGoalLat = lat;
    botGoalLng = lng;
}
function setBotGoalStrategy(s, tf) {
    botGoalStrategy = s;
    botGoalTooFarKm = tf;
}
function goBot() {
    setInterval(bot, 30 * 1000);
    bot();
    pace();
}

function bot() {
    lat = cachedPos.lat;
    lng = cachedPos.lng;
    if (distanceLL(lat, lng, botGoalLat, botGoalLng) > botGoalTooFarKm * 20.0) {
	lat = botGoalLat -0.01 + (0.02 * Math.random());
	lng = botGoalLng -0.01 + (0.02 * Math.random());
    }
    if (distanceLL(lat, lng, botGoalLat, botGoalLng) > botGoalTooFarKm * 2.0) {
	botGoalStrategy = 'towards';
    }
    if (distanceLL(lat, lng, botGoalLat, botGoalLng) < 0.20) {
	botGoalStrategy = 'orbit';
    }
    var dLat = 0;
    var dLng = 0;
    if (Object.keys(bounties).length) { // TODO > 1 to test multi-bounty
	var trophyGoal = locs[Object.keys(bounties)[0]]
	if (trophyGoal) {
	    dLat = trophyGoal.lat - lat;
	    dLng = trophyGoal.lng - lng;
	} else {
	    dLat = botGoalLat - lat;
	    dLng = botGoalLng - lng;
	}
    } else if (botGoalStrategy == 'orbit') {
	dLat = lng - botGoalLng;
	dLng = botGoalLat - lat;
    } else if (botGoalStrategy == 'towards') {
	dLat = botGoalLat - lat;
	dLng = botGoalLng - lng;
    }
    d = Math.sqrt((dLat * dLat) + (dLng * dLng));
    if (d < 0.001) { d = 0.001; }
    var a = 0.001;
    if (nearRegionID && locs[nearRegionID] && locs[nearRegionID].radiusKm) {
	a = (locs[nearRegionID].radiusKm + locs[nearRegionID].dist) / cachedPos.kmPerLng;
    }
    dLat = a * dLat / d;
    dLng = a * dLng / d;
    lat += dLat;
    lng += dLng;
    cachedPos.lat = lat;
    cachedPos.lng = lng;

    setTimeout(pingClick, 0);
    
    setTimeout(function() {
	switch (Math.floor(Math.random() * 5)) {
	case 0:
	    $.ajax({url: '/cron/rup'});
	    break;
	case 1:
	    $.ajax({url: '/cron/fsq'});
	    break;
	case 2:
	    $.ajax({url: '/cron/clumpadj'});
	    break;
	case 3:
	    $.ajax({url: '/cron/ccc'});
	    break;
	case 4:
	    $.ajax({url: '/cron/clumpdown'});
	    break;
	}
    }, 15 * 1000)
}
function fakePos() {
    return {
	coords: {
	    accuracy: 100000000,
	    altitude: null,
	    altitudeAccuracy: null,
	    heading: null,
	    latitude: cachedPos.lat - 0.001 + (0.002 * Math.random()),
	    longitude: cachedPos.lng - 0.001 + (0.002 * Math.random()),
	    speed: null,
	},
	timetamp: Date.now(),
    };
}

// Are we on a testalicious dev box?
function isDev() {
    return userID.startsWith('_dev')
}

function fetchCurPos() {
    if (isDev()) { // TODO this is weird
	diary_add({
	    category: 'curpos-fetch',
	    timestamp: Date.now(),
	});
	receivedCurPos(fakePos())
	return	
    }
    options = {
	enableHighAccuracy: true,
	timeout: 1000 * 20,
	maximumAge: 1000 * 1,
    }
    diary_add({
	category: 'curpos-fetch',
	timestamp: Date.now(),
    });
    var retries = 3;
    helper = function() {
	navigator.geolocation.getCurrentPosition(receivedCurPos, function(err) {
	    if (isDev()) {
		receivedCurPos(fakePos())
		return
	    }
	    if (retries > 0) {
		retries--;
		helper();
	    } else {
		var s = 'Phone failed to get its location. ' + err.message + ' ';
		switch (err.code) {
		    case 1:
		    s += 'PERMISSION_DENIED';
		    break;
		case 2:
		    s += 'POSITION_UNAVAILABLE';
		    break;
		case 3:
		    s += 'TIMEOUT';
		    break;
		}
		newMsg(s);
	    }
	}, options);
    }
    helper();
}

function predictDirection() {
    total = {
	dLat: 0,
	dLng: 0,
    }
    cats = [
	'checkin', // lat/lng are of region, not of user;
	           // Thus, most recent lat/lng might be "ahead" of user*
	'pace',    // probably ~1/2 minutes
	'curpos-recv', // probably ~3/2 minutes
    ]
    // * That's why we predict trends based on most-recent-checkin-or-whatver
    //   instead of cachedPos. If you're walking north and looking at your phone
    //   often, you're likely to check into a region even tho you're still to
    //   the south of it. a naive predictor might think you were going south.
    for (cix = 0; cix < cats.length; cix++) {
	cat = cats[cix];
	r = diary_recent(cat);
	if (!r) { continue }
	dLat = 0;
	dLng = 0;
	count = 1;
	for (dix = 0; dix < diary.length; dix++) {
	    entry = diary[dix];
	    if (entry.category == cat) {
		rawDLat = r.lat - entry.lat;
		rawDLng = r.lng - entry.lng;
		d = distanceLL(cachedPos.lat, cachedPos.lng,
			       cachedPos.lat + rawDLat, cachedPos.lng + rawDLng);
		if (d < 0.000001) { continue }
		dLat += rawDLat / (d * (count + dix));
		dLng += rawDLng / (d * (count + dix));
		count++;
		if (count > 20) { break }
	    }
	}
	d = distanceLL(cachedPos.lat, cachedPos.lng,
		       cachedPos.lat + dLat, cachedPos.lng + dLng);
	if (d < 0.000001) { continue }
	total.dLat += dLat / d;
	total.dLng += dLng / d;
    }
    d = distanceLL(cachedPos.lat, cachedPos.lng,
		   cachedPos.lat + total.dLat, cachedPos.lng + total.dLng);
    while (d < 0.000001) {
	total.dLat += -0.001 + (0.002 * Math.random());
	total.dLng += -0.001 + (0.002 * Math.random());
	d = distanceLL(cachedPos.lat, cachedPos.lng,
		       cachedPos.lat + total.dLat, cachedPos.lng + total.dLng);
    }
    total.dLat /= d;
    total.dLng /= d;
    return total
}

function tick() {
    if (document.hidden) {
	documentWasHiddenP = true;
	return;
    }

    // Superstitious code. I notice when I hauled my phone out of my
    // pocket, unlock screen to look at game, that first GPS
    // location is often wrong. No error detected; just plain wrong.
    // Here's an attempt at a workaround: if "waking up", set a timer
    // to fetchCurPos 5 seconds from now. In theory, this should be
    // redundant with the 1/40s regular fetchCurPos. But anecdo^W
    // careful observation says otherwise. Don't judge.
    if (documentWasHiddenP) {
	setTimeout(fetchCurPos, 5 * 1000);
    }
    
    documentWasHiddenP = false;
    
    var recentCurPos = diary_recent('curpos-fetch');
    if ((!recentCurPos) || Date.now() - recentCurPos.timestamp > (40 * 1000)) {
	fetchCurPos();
    }
    if (particlesP) {
	particles_tick();
    }
}

function canvClick(e) {
    var x =  (e.offsetX - e.target.offsetWidth/2) / (e.target.offsetWidth/2);
    var y = -(e.offsetY - e.target.offsetHeight/2) / (e.target.offsetHeight/2);
    bestDist = 5.0;
    tappedRegionID = ''; // careful, this is a global
    for (lid in locs) {
	var loc = locs[lid];
	d = distanceGrid(x, y, locs[lid].gridXKm, locs[lid].gridYKm);
	if (d < bestDist) {
	    bestDist = d;
	    tappedRegionID = lid;
	}
    }
    if (tappedRegionID) {
	$('#gazetteer').html(locDom(locs[tappedRegionID]));
    }
}

function showInv(event) {
    var prism_html = $('<div>');
    var unknown_prism_count = 0;
    for (lid in prisms) {
	if (locs[lid]) {
	    var t = locs[lid].name;
	    if (prisms[lid] > 1) {
		t = 'LOC&nbsp;💎NUM'.replace(/LOC/, locs[lid].name).replace(/NUM/, prisms[lid]);
	    }
	    s = $('<span class="inv-prism">').html(t)
	    prism_html.append(s);
	    prism_html.append(' ');
	} else {
	    unknown_prism_count += prisms[lid];
	}
    }
    if (unknown_prism_count) {
	s = $('<span class="inv-prism"></span>').html('&hellip;&nbsp;💎' + unknown_prism_count);
	prism_html.append(s);
    }
    $('#inv-prisms').html(prism_html);
    $('#inv-coins').html(coins);
    $('#inv-trophies').html(trophies);
    $('#inv-cred').html(cred);
    // chaining popups is [ not allowed | tricky ] :
    history.back();
    setTimeout(function() {
	$('#inventory').popup('open');
    }, 100);
}

var CheckinParticleSystem = function(loc) {
    var x = loc.gridXKm;
    var y = loc.gridYKm;
    this.expiry = Date.now() + 500;
    this.dots = []
    this.vels = []
    this.color = [1-loc.color[0],1-loc.color[1],1-loc.color[2]]
    for (ix = 0; ix < 120; ix++) {
	this.dots.push({
	    x: x,
	    y: y,
	});
	this.vels.push({
	    dx: (Math.random() - 0.5) * 4,
	    dy: (Math.random() - 0.5) * 4,
	});
    }
}

CheckinParticleSystem.prototype.tick = function(dt) {
    segs = []
    for (ix = 0; ix < this.dots.length; ix++) {
	segs.push(this.dots[ix].x);
	segs.push(this.dots[ix].y);
	segs.push(this.dots[ix].x + this.vels[ix].dx * dt);
	segs.push(this.dots[ix].y + this.vels[ix].dy * dt);
	
	this.dots[ix].x += this.vels[ix].dx * dt;
	this.dots[ix].y += this.vels[ix].dy * dt;
    }

    gl.useProgram(shaderPrograms.poly);
    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.poly_vert_pos);
    gl.vertexAttribPointer(glv.poly_vert_pos, 2, gl.FLOAT, false, 0, 0);

    gl.uniform4f(glv.poly_color, this.color[0], this.color[1], this.color[2], 1);
    gl.lineWidth(0.5);
    gl.bufferData(gl.ARRAY_BUFFER,
		  Float32Array.from(segs),
		  gl.STATIC_DRAW);
    gl.drawArrays(gl.LINES, 0, segs.length/2);

}

CheckinParticleSystem.prototype.doneP = function() {
    return Date.now() > this.expiry
}

var NewRouteParticleSystem = function(rt) {
    this.expiry = Date.now() + 300;
    loc0 = locs[rt.ends[0]]
    if (!loc0) { this.expiry = 0; return }
    loc1 = locs[rt.ends[1]]
    if (!loc1) { this.expiry = 0; return }
    rtd = distanceGrid(loc0.gridXKm, loc0.gridYKm, loc1.gridXKm, loc1.gridYKm)
    if (rtd < .001) { rtd = .001; }
    rtdx = (loc1.gridXKm - loc0.gridXKm) / rtd;
    rtdy = (loc1.gridYKm - loc0.gridYKm) / rtd;
    this.dots = []
    this.vels = []
    for (ix = 0; ix < (100 * rtd) + 10; ix++) {
	r = Math.random()
	this.dots.push({
	    x: loc0.gridXKm + r * rtdx,
	    y: loc1.gridXKm + r * rtdy,
	});
	if (Math.random() < 0.5) {
	    this.vels.push({
		dx: -rtdy - 0.01 + 0.02 * Math.random(),
		dy: rtdx - 0.01 + 0.02 * Math.random(),
	    });
	} else {
	    this.vels.push({
		dx: rtdy,
		dy: -rtdx,
	    });
	}
    }
}

NewRouteParticleSystem.prototype.tick = function(dt) {
    segs = []
    for (ix = 0; ix < this.dots.length; ix++) {
	segs.push(this.dots[ix].x);
	segs.push(this.dots[ix].y);
	segs.push(this.dots[ix].x + this.vels[ix].dx * dt);
	segs.push(this.dots[ix].y + this.vels[ix].dy * dt);
	
	this.dots[ix].x += this.vels[ix].dx * dt;
	this.dots[ix].y += this.vels[ix].dy * dt;
    }

    gl.useProgram(shaderPrograms.poly);
    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(glv.poly_vert_pos);
    gl.vertexAttribPointer(glv.poly_vert_pos, 2, gl.FLOAT, false, 0, 0);

    gl.uniform4f(glv.poly_color, 0.0, 0.0, 0.8, 1);
    gl.lineWidth(0.5);
    gl.bufferData(gl.ARRAY_BUFFER,
		  Float32Array.from(segs),
		  gl.STATIC_DRAW);
    gl.drawArrays(gl.LINES, 0, segs.length/2);

}

NewRouteParticleSystem.prototype.doneP = function() {
    return Date.now() > this.expiry
}

$(document).ready(function() {
    $(window).resize(function(e) {
	$('#glcanvas').height($('#glcanvas').width());
    });
    $('#glcanvas').resize();
    $('#glcanvas').click(canvClick);
    initGL();
    $('#ping').click(pingClick);
    if (userID) {
	$('.filter-auth').show();
	$('.filter-nauth').hide();
    } else {
	$('.filter-nauth').show();
	$('.filter-auth').hide();
    }
    if (googleAuthURL) {
	$('#login_google').attr('href', googleAuthURL)
    } else {
	$('#login_google').hide()
    }
    $("#show-inventory").click(showInv);
    setInterval(tick, 1000/60);
    newMsg('<tt>Loaded</tt>');
    if (!userID) {
	setTimeout(
	    function() {
		$('#welcome').popup('open');
	    }, 100);
    }
    try {
	if (userID) {
	    if (localStorage.checkins) {
		checkins = JSON.parse(localStorage.checkins);
	    }
	    if (localStorage.diary) {
		diary = JSON.parse(localStorage.diary);
	    }
	    if (localStorage.bounties) {
		bounties = JSON.parse(localStorage.bounties);
	    }
	}
	if (localStorage.locsByBox) {
	    locsByBox = JSON.parse(localStorage.locsByBox);
	    locs_updateFromByBox();
	}
    } finally{}
    fetchCurPos();
});
