var oi = document.getElementById('origimg');
var oc = document.getElementById('origcnv');
var dc = document.getElementById('drawcnv');
var octx = oc.getContext('2d');
var dctx = dc.getContext('2d');
var r;
var x;
var y;
var attract = 0;
var preload_ix = 0;
var galload_ix = 0;

IMGS = [
    // https://sfgov.org/sfc/bac/san-francisco-bicycle-plan
    ['Bike', 'bike_ggbridge.jpg'], 

    ['Canyon', 'USA_10654_Bryce_Canyon_Luca_Galuzzi_2007.jpg'],

    // https://www.flickr.com/photos/mattblaze/50246351407
    ['Curta', 'curta.jpg'],

    // https://pixabay.com/illustrations/ganesh-india-ganesha-hinduism-god-4784991/
    ['Ganesh', 'ganesh.jpg'], 

    ['Labyrinth', 'labyrinth.png'],
    ['Nerd', 'new-glasses-selfie.jpg'],
    ['Rosie', 'rosie.jpg'],
    ['Stop', 'stop-sign.jpg'],
    ['U.S.', 'unclesam.jpg'],

    // https://withgoodreasonradio.org/episode/election-episode/
    ['Voted', 'voted.jpg'], 

    ['Wave', 'Tsunami_by_hokusai_19th_century.jpg'],
    ['Xing', 'ped-xing.jpg'],
];

GALS = [
    { i: "gaal", s: "gal/american-landscape.png"},
    { i: "gabl", s: "gal/blackbird.png"},
    { i: "gach", s: "gal/childs-bath.png"},
    { i: "gado", s: "gal/doge.png"},
    { i: "gape", s: "gal/persistence-kat-clocks.png"},
    { i: "gare", s: "gal/red-dots.png"},
    { i: "gase", s: "gal/self-portrait-with-monkey.png"},
    { i: "gasl", s: "gal/sluggo-is-lit.png"},
];
    
for (var ix = 0; ix < IMGS.length; ix++) {
    n = IMGS[ix][0];
    f = IMGS[ix][1];
    var presets = document.getElementById('presets');
    var btn = document.createElement('button');
    btn.innerHTML = n;
    btn.value = f;
    btn.addEventListener('click', function() {
	oi.src = this.value;
	if (oi.complete) { reset(); }
    });
    presets.appendChild(btn);
    presets.appendChild(document.createTextNode(' '));
}


function reset() {
    if (attract) {
	clearTimeout(attract);
	attract = 0;
    }
    var w = oi.width;
    var h = oi.height;
    var innerHeight = window.innerHeight - 170;
    if (h > innerHeight) {
	w *= innerHeight / h;
	h = innerHeight;
    }
    if (w > window.innerWidth) {
	h *= window.innerWidth / w;
	w = window.innerWidth;
    }
    oc.width = w;
    oc.height = h;
    dc.width = w;
    dc.height = h;
    
    octx.drawImage(oi, 0, 0, w, h);
    r = Math.sqrt(w * oi.height);
    if (!r) { r = 1000; }
    setTimeout(drip, 0);
    x = Math.floor(Math.random() * w);
    y = Math.floor(Math.random() * h);
}

function drip() {
    if (r < 0.7) {
	if (!attract) {
	    attract = setTimeout(function() {
		var ix = Math.floor(Math.random() * IMGS.length);
		oi.src = IMGS[ix][1];
		reset();
	    }, 5000);
	}
	return
    }
    if (attract) {
	clearTimeout(attract);
	attract = 0;
    }
    setTimeout(drip, 150);
    var sparse = document.getElementById('sparse').value;

    for (var count = 0; count < oc.width * 10 / sparse; count += r) {

	var orig = octx.getImageData(x, y, 1, 1);
	var drawn = dctx.getImageData(x, y, 1, 1);
	color_dist_2 = (Math.pow(orig.data[0]-drawn.data[0], 2) +
			Math.pow(orig.data[1]-drawn.data[1], 2) +
			Math.pow(orig.data[2]-drawn.data[2], 2));
	if (color_dist_2 >= 1000) { 
	    dctx.fillStyle = ('rgb(' + orig.data[0] + ',' + orig.data[1]
			      + ',' + orig.data[2] + ')');
	    dctx.beginPath();
	    dctx.arc(x, y, r, 0, 2 * Math.PI);
	    dctx.fill();
	}
	x = Math.floor(x + sparse*r);
	if (x >= oc.width) {
	    x = x % oc.width;
	    y = Math.floor(y + sparse*r);
	}
	if (y >= oc.height) {
	    y = y % oc.height;
	    r = 0.71 * r;
	}
    }
}

reset();
oi.addEventListener('load', reset);
document.getElementById('preloadimg').addEventListener('load', preload);

document.getElementById('toggle').addEventListener('click', function() {
    var div = document.getElementById('about');
    if (div.style.display == 'none') {
	div.style.display = 'block';
    } else {
	div.style.display = 'none';
    }
});

function preload() {
    if (preload_ix < IMGS.length) {
	var preloadimg = document.getElementById('preloadimg');
	preloadimg.src = IMGS[preload_ix][1];
	preload_ix++;
    } else {
	galload();
    }
}

function galload() {
    if (galload_ix < GALS.length) {
	var gal = GALS[galload_ix];
	var img = document.getElementById(gal.i);
	img.addEventListener('load', galload);
	img.src = gal.s;
	galload_ix++;
    }
}
setTimeout(preload, 5000);
