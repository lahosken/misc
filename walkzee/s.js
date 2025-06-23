var locs = JSON.parse(localStorage.getItem("wz-locs") || "{}");
var locTimes = JSON.parse(localStorage.getItem("wz-locTimes") || "{}");
var hopper = JSON.parse(localStorage.getItem("wz-hopper") || "[]" );
var hands = JSON.parse(localStorage.getItem("wz-hands") || "[]" );
var score = JSON.parse(localStorage.getItem("wz-score") || "10" );


// transient state
var claw = { area: "void" }; // which dice did I pick up (intending to move)?

var cachedPos = { // where am I (as of last time fetched position)?
    lat: 0, lng: 0,
    kmPerLat: 111.1,
    kmPerLng: 87.83,
};

const HANDRULE = {
    "5kind": { p: "Five of a kind", l: "5×❓", j: is5Kind, pt: 5 },
    "4kind": { p: "Four of a kind", l: "4×❓", j: is4Kind, pt: 3 },
    "Straight": { p: "Straight", l: "Straight", j: isStraight, pt: 2 },
    "House": { p: "Full House", l: "🏠", j: isFullHouse, pt: 2 },
    "3Kind": { p: "Three of a kind", l: "3×❓", j: is3Kind, pt: 1.9 },
    "2Pair": { p: "Two Pair", l: "2×❓,&nbsp;2×❓", j: is2Pair, pt: 1.9 },
    "Aces": { p: "Pair of ⚀(1)s", l: '2×<img src="ace.png">', j: isPair1s, pt: 1.7 },
    "Deuces": { p: "Pair of ⚁(2)s", l: '2×<img src="deuce.png">', j: isPair2s, pt: 1.2  },
    "Treys": { p: "Pair of ⚂(3)s", l: '2×<img src="trey.png">', j: isPair3s, pt: 1.3  },
    "Caters": { p: "Pair of ⚃(4)s", l: '2×<img src="cater.png">',  j: isPair4s, pt: 1.4  },
    "Cinques": { p: "Pair of ⚄(5)s", l: '2×<img src="cinque.png">', j: isPair5s, pt: 1.5  },
    "Boxcars": { p: "Pair of ⚅(6)s", l: '2×<img src="boxcar.png">', j: isPair6s, pt: 1.6  },
};

const UNICS = "○⚀⚁⚂⚃⚄⚅";

function is5Kind(hand) {
    if (hand[0] <= 0) { return false }
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] != hand[0]) { return false }
    }
    return true
}
function is4Kind(hand) {
    var counts = [0, 0, 0, 0, 0, 0, 0];
    hand.forEach((die) => {
	counts[die]++;
    });
    if (counts[0] > 0) return false
    for (var ix = 1; ix < counts.length; ix++) {
	if (counts[ix] >= 4) return true
    }
    return false
}
function is3Kind(hand) {
    var counts = [0, 0, 0, 0, 0, 0, 0];
    hand.forEach((die) => {
	counts[die]++;
    });
    if (counts[0] > 0) return false
    for (var ix = 1; ix < counts.length; ix++) {
	if (counts[ix] >= 3) return true
    }
    return false
}
function isFullHouse(hand) {
    if (is5Kind(hand)) return true
    var counts = [0, 0, 0, 0, 0, 0, 0];
    hand.forEach((die) => {
	counts[die]++;
    });
    var got3 = false;
    var got2 = false;
    for (var ix = 1; ix < counts.length; ix++) {
	if (counts[ix] == 5) { return true } // 5 of a kind is a full house i guess
	if (counts[ix] == 3) { got3 = true; }
	if (counts[ix] == 2) { got2 = true; }
    }
    return (got3 && got2)
}
function is2Pair(hand) {
    if (is4Kind(hand)) return true
    var counts = [0, 0, 0, 0, 0, 0, 0];
    hand.forEach((die) => {
	counts[die]++;
	if (die <= 0) { return false }
    });
    var pairs = 0
    for (var ix = 1; ix < counts.length; ix++) {
	if (counts[ix] >= 2) { pairs++; }
    }
    return (pairs >= 2)
}
function isStraight(hand) {
    var counts = [0, 0, 0, 0, 0, 0, 0];
    hand.forEach((die) => {
	counts[die]++;
    });
    if (counts[0] > 0) return false
    for (var ix = 1; ix < counts.length; ix++) {
	if (counts[ix] > 1) return false
    }
    if (counts[1] && counts[6]) return false
    return true
}
function isPair1s(hand) {
    const t = 1;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}
function isPair2s(hand) {
    const t = 2;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}
function isPair3s(hand) {
    const t = 3;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}
function isPair4s(hand) {
    const t = 4;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}
function isPair5s(hand) {
    const t = 5;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}
function isPair6s(hand) {
    const t = 6;
    count = 0;
    for (var ix = 0; ix < hand.length; ix++) {
	if (hand[ix] <= 0) return false
	if (hand[ix] == t) { count++ }
    }
    return (count > 1)
}

window.addEventListener("resize", (event) => {
    const c = document.getElementById("rvcanvas");
    c.height = c.width;

    redraw();
});
window.addEventListener("load", (event) => {
    const rf = document.getElementById("refresh");
    rf.addEventListener("click", rfBtnClick);
    const c = document.getElementById("rvcanvas");
    c.height = c.width;
    c.addEventListener("click", canvClick);

    fetchCurPos();

    setInterval(tickMinute, 60 * 1000);

    rightsizeHopper();
    while (hands.length < 3) {
	addHand();
    }
    redraw();
}, false);

function rfBtnClick(ev) {
    ev.target.disabled = true;
    fetchCurPos();
    setTimeout(() => {
	ev.target.disabled = false;
    }, 10 * 1000);
}

function rightsizeHopper() {
    var zeroCount = 0;
    hopper.forEach((v) => {
	if (v == 0) { zeroCount++ }
    });
    if ((zeroCount < 3) || (zeroCount / hopper.length < 0.33)) {
	hopper.push(0, 0, 0)
	return
    }
    if (zeroCount < 6) return
    if (zeroCount / hopper.length < 0.4) return
    if (hopper[hopper.length-1] == 0) {
	hopper.length--
	return
    }
}

function newMsg(s) {
    if (s) {
	document.getElementById("status").innerText = s;
    } else {
	document.getElementById("status").innerHTML = "&nbsp;";
    }
}

function dist(lat1, lng1, lat2, lng2) {
    const nsKm = (lat2 - lat1) * cachedPos.kmPerLat;
    const ewKm = (lng2 - lng1) * cachedPos.kmPerLng;
    return Math.sqrt((nsKm ** 2) + (ewKm ** 2))
}

function canvClick(e) {
    const eKm = (e.offsetX - e.target.offsetWidth/2) / (e.target.offsetWidth/2);
    const nKm = -(e.offsetY - e.target.offsetHeight/2) / (e.target.offsetHeight/2);
    const clickLat = cachedPos.lat + (nKm / cachedPos.kmPerLat);
    const clickLng = cachedPos.lng + (eKm / cachedPos.kmPerLng);

    var blocksToCheck = {};
    for (var northOffset = -1; northOffset < 2; northOffset++) {
	for (var eastOffset = -1; eastOffset < 2; eastOffset++) {
	    const neighborLat = cachedPos.lat + (northOffset / cachedPos.kmPerLat);
	    const neighborLng = cachedPos.lng + (eastOffset / cachedPos.kmPerLng);
	    const neighborLatX10 = Math.floor(neighborLat * 10.0)
	    const neighborLngX10 = Math.floor(neighborLng * 10.0)
	    const blockKey = "" + neighborLatX10 + "," + neighborLngX10;
	    blocksToCheck[blockKey] = true;
	}
    }
    var closest_dist = 9999;
    var closest_bix = {
	block: "",
	ix: -1
    }
    Object.keys(blocksToCheck).forEach((block) => {
	for (var ix = 0; ix < locs[block].length; ix++) {
	    const loc = locs[block][ix];
	    const d = dist(loc.lat, loc.lng, clickLat, clickLng);
	    if (d < closest_dist) {
		closest_dist = d;
		closest_bix = {
		    block: block,
		    ix: ix,
		}
	    }
	}
    });
    if (closest_bix.block.length == 0) {
	console.log("clicked nothing, I guess.")
	return
    }
    const clickedLoc = locs[closest_bix.block][closest_bix.ix];
    if (clickedLoc.roll <= 0) {
	claw = { area: "void" }
	newMsg();
	redraw();
	return
    }
    if (dist(clickedLoc.lat, clickedLoc.lng, cachedPos.lat, cachedPos.lng) > 0.5) {
	newMsg('Too far. Get closer to "pick up" this die.')
	return
    }
    claw = {
	area: "map",
	block: closest_bix.block,
	ix: closest_bix.ix,
    }
    redraw();
    newMsg("Where do you want it?");
}

function drawMap() {
    const cv = document.getElementById("rvcanvas");
    const ctx = cv.getContext("2d");
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.font = "" + cv.height/16 + "px serif";
    ctx.fillStyle = "green";
    ctx.fillRect(0, 0, cv.width, cv.height);
    ctx.fillStyle = "black";
    pxPerKm = cv.width/2;
    for (const [blockKey, blockLocs] of Object.entries(locs)) {
	for (var ix = 0; ix < blockLocs.length; ix++) {
	    const loc = blockLocs[ix];
	    const northOffsetKm = (loc.lat - cachedPos.lat) * cachedPos.kmPerLat;
	    if (Math.abs(northOffsetKm) > 1.5) continue
	    const y = (cv.height/2) - (pxPerKm * northOffsetKm);
	    const eastOffsetKm = (loc.lng - cachedPos.lng) * cachedPos.kmPerLng;
	    if (Math.abs(eastOffsetKm) > 1.5) continue
            const x = (cv.width/2) + (pxPerKm * eastOffsetKm);
            if (claw.area == "map" && claw.block == blockKey && claw.ix == ix) {
		ctx.fillStyle = "yellow";
	    } else if (loc.roll == 0) {
		ctx.fillStyle = "green";		
	    }else if (dist(loc.lat, loc.lng, cachedPos.lat, cachedPos.lng) < 0.5) {
		ctx.fillStyle = "white";
	    } else {
		ctx.fillStyle = "lime";
	    }
	    ctx.beginPath();
	    ctx.arc(x, y, 20, 0, 2 * Math.PI); // circle 20 radius
	    ctx.fill();
	    ctx.fillStyle = "black";
	    ctx.fillText(UNICS[loc.roll], x, y);
	}
    }
}

/* Testing while sitting at home? Instead of getting permission
   to use geolocation that's not gonna be interesting, just
   jiggle the coordinates a little every so often.
 */
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

function isDev() {
    return document.location.protocol == 'file:';
}

function fetchCurPosDesperate() {
    fetchCurPos(true);
}

function fetchCurPos(desperateP) {
    if (isDev()) { // If I'm testing on desktop in my apartment, don't try to gelocate
        receivedCurPos(fakePos())
        return
    }
    options = {
        enableHighAccuracy: true,
        timeout: 1000 * 20,
        maximumAge: 1000 * 1,
    }
    if (desperateP) {
        options.enableHighAccuracy = false;
    }
    var error;
    for (var retries = 0; retries < 3; retries++) {
	error = false;
	navigator.geolocation.getCurrentPosition(receivedCurPos, function(err) {
	    error = err
	})
	if (!error) break
    }
    if (error) {
	var s = 'Phone failed to get its location. ' + err.message + ' ';
	switch(err.code) {
	case 1:
	    s += '(permission denied)';
	    break
	case 2:
	    s += '(position unavailable)';
	    break
	case 3:
	    s += '(timeout)';
	    break
	}
	newMsg(s);
    }
}

function refreshNearbyBlocks() {
    const now = Date.now();
    const anHourAgo = now - (60 * 60 * 1000)
    const threeHoursAgo = now - (3 * 60 * 60 * 1000)
    for (var northOffset = -1; northOffset < 2; northOffset++) {
	for (var eastOffset = -1; eastOffset < 2; eastOffset++) {
	    const neighborLat = cachedPos.lat + (northOffset / cachedPos.kmPerLat);
	    const neighborLng = cachedPos.lng + (eastOffset / cachedPos.kmPerLng);
	    const neighborLatX10 = Math.floor(neighborLat * 10.0)
	    const neighborLngX10 = Math.floor(neighborLng * 10.0)
	    const blockKey = "" + neighborLatX10 + "," + neighborLngX10;
	    if (locTimes[blockKey] > anHourAgo) {
		continue
	    }
	    if (! (blockKey in locs)) {
		locs[blockKey] = [];
	    }
	    if (locTimes[blockKey] < threeHoursAgo) {
		locs[blockKey] = locs[blockKey].filter((l) => {return l.roll > 0})
	    }
	    locTimes[blockKey] = now;
	    if (neighborLat > 80) continue
	    if (neighborLat < -80) continue
	    var collisions = 0;
	    var locKeyInt = 0;
	    while (collisions < 1000) {
		locKeyInt++;
		const potentialLat = (neighborLatX10 + Math.random()) / 10.0;
		const potentialLng = (neighborLngX10 + Math.random()) / 10.0;
		var collided = false;
		for (const [otherBlockKey, otherLocs] of Object.entries(locs)) {
		    otherLocs.forEach((otherLoc) => {
			if (dist(potentialLat, potentialLng, otherLoc.lat, otherLoc.lng) < 0.2) {
			    collided = true;
			    return
			}
		    });
		}
		if (collided) {
		    collisions += 1
		    continue
		}
		collisions = 0
		const dieRoll = Math.floor(Math.random() * 6 + 1);
		const locKey = "" + locKeyInt;
		locs[blockKey].push({
		    lat: potentialLat,
		    lng: potentialLng,
		    roll: dieRoll,
		});
		collisions += 1;
	    } // done while (collisions < 1000)
	    
	    locs[blockKey] = locs[blockKey].filter((l) => {return l.roll > 0})
	}
    }
    persist();
}

/*
 * If the user's on an around-the-world cruise, the locs data structure
 * might get big.  So maybe cull some old blocks we haven't seen for a while
*/
function cullOldBlocks() {
    const blockKeys = Object.keys(locs);
    if (blockKeys.length < 10) return

    var deleted = 0;
    for (var ix = 0; ix < blockKeys.length; ix++) {
	const blockKey = blockKeys[ix];
	if (! (blockKey in locTimes)) {
	    delete locs[blockKey];
	    deleted++;
	}
    }
    if (deleted) { return }
    
    const now = Date.now();
    const twoHoursAgo = now - (2 * 60 * 60 * 1000);
    const randKey = blockKeys[Math.floor(Math.random() * blockKeys.length)];
    const longAgo = Math.min(twoHoursAgo, locTimes[randKey])
    
    for (var ix = 0; ix < blockKeys.length; ix++) {
	const blockKey = blockKeys[ix];
	if (locTimes[blockKey] < longAgo) {
	    delete locs[blockKey];
	    delete locTimes[blockKey];
	    deleted++;
	}
    }
}

function receivedCurPos(pos) {
    cachedPos = {
        lat: pos.coords.latitude,
        lng: pos.coords.longitude,
        kmPerLat: 111.1,
        kmPerLng: 111.1 * Math.cos(pos.coords.latitude * 3.14159 / 180.0),
    }
    const rf = document.getElementById("refresh");
    refreshNearbyBlocks();
    redraw();
    
    // The geolocation API might give an olllld location, but not
    // give us an error message or anything. What a scamp! 
    if (Date.now() - pos.timestamp > 9 * 1000) {
        // Maybe we should retry? Maybe?
        // (An earlier implementation always retried, w/short delay. That
        //  pretty much locked up my phone when I went through a tunnel, so
        //  that wasn't a good handler. 
        //  So... uhm, randomly maybe retry with some random delay or
	//  something. Not well thought-out, obviously, but worked OK
        //  for many years in Troubadour Tour Board, so good enough?)
        if (Math.random() < 0.7) {
            setTimeout(fetchCurPosDesperate, 1000 * (Math.random() + 0.01));
        }
	
    }
}

function addHand() {
    while (true) {
	const ruleKey = Object.keys(HANDRULE)[Math.floor(Math.random() * Object.keys(HANDRULE).length)];
	var collision = false
	for (ix = 0; ix < hands.length; ix++) {
	    if (hands[ix].rule == ruleKey) { collision = true; break }
	}
	if (collision) { continue }
	hands.push({ rule: ruleKey, dice: [0, 0, 0, 0, 0]})
	break
    }
}

function showHopper() {
    const hopperDiv = document.getElementById("hopper");
    hopperDiv.replaceChildren();
    for (var ix = 0; ix < hopper.length; ix++) {
	var dieButton = document.createElement("button");
	dieButton.appendChild(document.createTextNode(UNICS[hopper[ix]]));
	dieButton.payload = { d: ix, }
	if ((claw.area == "hop") && (claw.die == ix)) {
	    dieButton.className = "claw";
	} else {
	    dieButton.className = "noclaw";
	}
	dieButton.addEventListener("click", hopBtnClick);
	hopperDiv.appendChild(dieButton);
    }
}
function showHands() {
    const handsDiv = document.getElementById("hands");
    handsDiv.replaceChildren();
    for (var hix = 0; hix < hands.length; hix++) {
	hand = hands[hix];
	var handDiv = document.createElement("div");
	
	var ruleSpan = document.createElement("div");
	ruleSpan.className = "hand1rule";
	ruleSpan.innerHTML = HANDRULE[hand.rule].l;
	handDiv.appendChild(ruleSpan);

	handDiv.appendChild(document.createTextNode(" "));

	for (var dix = 0; dix < hand.dice.length; dix++) {
	    var dieButton = document.createElement("button");
	    dieButton.appendChild(document.createTextNode(UNICS[hand.dice[dix]]));
	    dieButton.payload = { h: hix, d: dix, }
	    dieButton.addEventListener("click", dieBtnClick);
	    if ((claw.area == "hands") && (claw.hand == hix) && (claw.die == dix)) {
		dieButton.className = "claw";
	    } else {
		dieButton.className = "noclaw";
	    }
	    handDiv.appendChild(dieButton);
	}

	handDiv.appendChild(document.createTextNode(" "));

	var claimButton = document.createElement("button");
	claimButton.appendChild(document.createTextNode("🏆"));
	claimButton.disabled = !HANDRULE[hand.rule].j(hand.dice);
	claimButton.payload = { h: hix };
	claimButton.addEventListener("click", claimBtnClick);
	handDiv.appendChild(claimButton);
	
	handsDiv.appendChild(handDiv);
    }    
}

function hopBtnClick(ev) {
    const payload = ev.target.payload;    
    var die = hopper[payload.d];
    if (die) {
	claw = {
	    area: "hop",
	    die: payload.d,
	}
	newMsg("Where do you want it?");
	redraw();
	return
    } 
    rightsizeHopper();
    if (claw.area == "void") return
    if (claw.area == "hands") {
	hopper[payload.d] = hands[claw.hand].dice[claw.die];
	hands[claw.hand].dice[claw.die] = 0;
	claw = { area: "void" };
	redraw();
	newMsg();
    }
    if (claw.area == "hop") {
	hopper[payload.d] = hopper[claw.die];
	hopper[claw.die] = 0;
	claw = { area: "void" };
	redraw();
	newMsg();
    }
    if (claw.area == "map") {
	if (!(claw.block in locs)) { claw = { area: "void" }; return }
	if (locs[claw.block].length <= claw.ix) { claw = { area: "void" }; return }
	hopper[payload.d] = locs[claw.block][claw.ix].roll;
	locs[claw.block][claw.ix].roll = 0;
	claw = { area: "void" };
	redraw();
	newMsg();
    }

    persist();
}
function dieBtnClick(ev) {
    const payload = ev.target.payload;
    var hand = hands[payload.h];
    var die = hand.dice[payload.d];
    if (die) {
	claw = {
	    area: "hands",
	    hand: payload.h,
	    die: payload.d,
	}
	newMsg("Where do you want it?");
	redraw();
    } else {
	if (claw.area == "void") return
	if (claw.area == "hands") {
	    const tmp = hands[claw.hand].dice[claw.die];
	    hands[claw.hand].dice[claw.die] = 0;
	    hands[payload.h].dice[payload.d] = tmp;
	    newMsg("Working towards that " + HANDRULE[hands[payload.h].rule].p);
	    claw = { area: "void" };
	    redraw();
	}
	if (claw.area == "hop") {
	    hands[payload.h].dice[payload.d] = hopper[claw.die];
	    hopper[claw.die] = 0;
	    claw = { area: "void" };
	    newMsg("Working towards that " + HANDRULE[hands[payload.h].rule].p);
	    rightsizeHopper();
	    redraw();
	}
	if (claw.area == "map") {
	    if (!(claw.block in locs)) { claw = { area: "void" }; return }
	    if (locs[claw.block].length <= claw.ix) { claw = { area: "void" }; return }
	    hands[payload.h].dice[payload.d] = locs[claw.block][claw.ix].roll;
	    locs[claw.block][claw.ix].roll = 0;
	    newMsg("Working towards that " + HANDRULE[hands[payload.h].rule].p);
	    claw = { area: "void" }; 
	    redraw();
	}
    }
    persist();
}

function claimBtnClick(ev) {
    const beforeScore = score;
    if (score < 2) score = 2;
    const rule = HANDRULE[hands[ev.target.payload.h].rule];
    score += Math.floor( Math.log2(score) * rule.pt);
    newMsg("Nice work! Your score went up! " + beforeScore + " ↗ " + score);
    addHand();
    var salvageHands = [];
    for (var ix = 0; ix < hands.length; ix++) {
	if (ev.target.payload.h == ix) continue
	salvageHands.push(hands[ix]);
    }
    hands = salvageHands;
    redraw();
    persist();
}

function redraw() {
    drawMap();
    showHands();
    showHopper();
}

function persist() {
    localStorage.setItem("wz-locs", JSON.stringify(locs));
    localStorage.setItem("wz-locTimes", JSON.stringify(locTimes));
    localStorage.setItem("wz-hopper", JSON.stringify(hopper));
    localStorage.setItem("wz-hands", JSON.stringify(hands));
    localStorage.setItem("wz-score", JSON.stringify(score));
}

function tickMinute() {
    if (document.hidden) { return }
    if (cachedPos.lat || cachedPos.lng) {
	cullOldBlocks();
    }
    setTimeout(fetchCurPos, 5 * 1000);
}
