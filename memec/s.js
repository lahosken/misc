function muffle(e) {
    e.preventDefault();
    e.stopPropagation();
}

function inputtery_ondrop(e) {
    e.preventDefault();
    e.stopPropagation();

    const bkgimg = document.getElementById('bkgimg');

    if (e.dataTransfer && e.dataTransfer.files && e.dataTransfer.files.length) {
        file = e.dataTransfer.files[0];
        URL.revokeObjectURL(bkgimg.src);
        bkgimg.src = URL.createObjectURL(file);
    }
    document.getElementById('inputtery').style.backgroundColor = 'transparent';
}

function infile_onchange(e) {
    const infile = document.getElementById('infile');
    const bkgimg = document.getElementById('bkgimg');

    if (infile.files.length) {
        file = infile.files[0];
        URL.revokeObjectURL(bkgimg.src);
        bkgimg.src = URL.createObjectURL(file);
    }
}

function render() {
    const bkgimg = document.getElementById('bkgimg');
    const canv = document.getElementById('canv');
    const ctx = canv.getContext('2d');
    const fill = document.getElementById('fill');
    const outline = document.getElementById('outline');
    const outimg = document.getElementById('outimg');

    const toptext = document.getElementById('toptext');
    const bottomtext = document.getElementById('bottomtext');
    const allovertext = document.getElementById('allovertext');

    canv.width = bkgimg.width;
    canv.height = bkgimg.height;
    ctx.fillRect(0, 0, canv.width, canv.height);

    if (bkgimg.src && bkgimg.complete) {
	try {
            ctx.drawImage(bkgimg, 0, 0);
	} catch (error) {
	    console.log(error);
	    bkgimg.src = "builtin/squinting-fry-not-sure-if.jpg";
	    return;
	}
    }

    ctx.fillStyle = fill.value;
    ctx.strokeStyle = outline.value;
    ctx.textAlign = 'center';
    ctx.lineWidth = 2;
    
    var fontSize =  Number.parseFloat(document.getElementById('fontSize').value);
    if (fontSize < 12) { fontSize = 12 }
    if (fontSize > 1200) { fontSize = 1200 }
    var typeface = document.getElementById('typeface').value;
    if (typeface == '' || !typeface) { typeface = 'Impact'; }
    const BOLD_THESE = {
	'serif': 1,
	'sans-serif': 1,
	'monspace': 1,
	'cursive': 1,
    };

    if (BOLD_THESE[typeface]) {
	ctx.font = `bold ${fontSize}px ${typeface}`;
    } else {
	ctx.font = `${fontSize}px ${typeface}`;
    }

    var lines;
    var carriage;
    lines = toptext.value.split('\n');
    carriage = 0;
    for (var i = 0; i < lines.length; i++) {
        carriage += fontSize;
        ctx.fillText(lines[i], canv.width/2, carriage);
        ctx.strokeText(lines[i], canv.width/2, carriage);
    }

    lines = bottomtext.value.split('\n');
    carriage = canv.height - 10 - (lines.length * fontSize);
    for (var i = 0; i < lines.length; i++) {
        carriage += fontSize;
        ctx.fillText(lines[i], canv.width/2, carriage);
        ctx.strokeText(lines[i], canv.width/2, carriage);
    }

    ctx.textAlign = 'start';
    lines = allovertext.value.split('\n');
    carriage = 0;
    for (var i = 0; i < lines.length; i++) {
        carriage += fontSize;
        ctx.fillText(lines[i], fontSize, carriage);
        ctx.strokeText(lines[i], fontSize, carriage);
    }

    URL.revokeObjectURL(outimg.src);
    outimg.src = canv.toDataURL("image/jpeg", 0.9);
}

function bkgimg_onerror(e) {
    console.log("Background image got error. ", e);
    document.getElementById('bkgimg').src = "builtin/squinting-fry-not-sure-if.jpg";
}

function builtin_handle(e) {
    const bkgimg = document.getElementById('bkgimg');
    const v = document.getElementById('builtin').value;
    for (var i = 0; i < BUILTIN.length; i++) {
	if (v == BUILTIN[i]) {
	    bkgimg.src = `builtin/${v}`;
	    break
	}
    }
}

function init_listeners() {
    const bkgimg = document.getElementById('bkgimg');
    const builtin = document.getElementById('builtin');
    const infile = document.getElementById('infile');
    const inputtery = document.getElementById('inputtery');

    bkgimg.addEventListener('event', bkgimg_onerror);

    builtin.addEventListener('change', builtin_handle);
    builtin.addEventListener('input', builtin_handle);

    inputtery.addEventListener('dragenter', muffle);
    inputtery.addEventListener('dragover', muffle);
    inputtery.addEventListener('dragleave', muffle);

    inputtery.addEventListener('drop', inputtery_ondrop);
    infile.addEventListener('change', infile_onchange);
}

function init_builtin_options() {
    var h = "";
    for (var i = 0; i < BUILTIN.length; i++) {
	h += `<option value="${BUILTIN[i]}">`;
    }
    document.getElementById('builtinOptions').innerHTML = h;
}

function init() {
    init_listeners();
    init_builtin_options();
    setInterval(render, 30);
}
init()


