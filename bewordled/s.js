var GRID_H = 7;
var GRID_W = 5;

var BASE_TICK_TIME = 500;
var TICK_TIME = BASE_TICK_TIME;

// Scoring is fun. I haven't thought much about how it should work, tho.
var score = 0;
var scoreBoost = 1.0;

// previously-tapped cell. Player taps two cells to swap their
// contents; previous_tapped keeps track of the first-tapped cell.
var previous_tapped = false;

// ui_state: are we listening for taps?
var ui_state = false;

// grid:
// state of the game grid,
// an array of array of strings.
// For example:
// [['a', 'b', 'c', 'd'],
//  ['e', 'f', 'g', 'h'],
//  ['i', 'j', 'k', 'l']]
// would look like this in game:
// a e i
// b f j
// c g k
// d h l

// ! is a cloud
// ? is an unexploded firecracker
// !! is an exploding firecracker
var grid = [];

// init the grid global variable to an empty state
function initGrid() {
    row = [];
    for (var j = 0; j < GRID_H; j++) {
	row.push('');
    }
    grid = [];
    for (var i = 0; i < GRID_W; i++) {
	grid.push(Array.from(row));
    }
}

// get a deep-copy of the grid global variable
function copyGrid() {
    return JSON.parse(JSON.stringify(grid))
}

function randChoice(l) {
    return l[Math.floor(Math.random() * l.length)];
}

function fillOneTopCell(gr) {
    const offset = Math.floor(Math.random() * GRID_W);
    for (var jj = 0; jj < GRID_W; jj++) {
	const j = (offset + jj) % GRID_W;
	if (gr[j][0]) { continue }
	var gramO = {}; // dict of grams we're considering dropping
	for (var ix in GRAMS) { gramO[GRAMS[ix]] = true; }
	/*
	 * Try to avoid dropping tiles already overrepresented:
         * randomly look at a few tiles from the grid and remove
         * them from consideration
	 */
	for (var ix = 0; ix < GRAMS.length; ix++) {
	    var g = randChoice(randChoice(gr))
	    if (gramO[g]) {
		delete gramO[g];
	    }
	}
	/*
         *  If we drop a 'c' on top of 'at', that gives the player
         *  a word "for free." Hardcore players might find this too
         *  easy, so try to detect this case and wriggle out of it.
         */
	var colTop = "";
	var colTopCount = 0;
	for (var i = 0; i < gr[j].length; i++) {
	    if (gr[j][i].match(/[a-z]/)) {
		colTop += gr[j][i];
		colTopCount += 1;
		if (colTopCount >= 2) { break }
	    }
	}
	for (var ix = 0; ix < GRAMS.length; ix++) {
	    const g = GRAMS[ix];
	    if (!gramO[g]) { continue }
	    if (WORDS[g + colTop]) {
		delete gramO[g];
	    }
	}

	// Wriggle out of many "lucky" horizontal words, part 1
	// Don't drop a tile which would be the first tile in a "lucky" word
	if (j < gr.length - 2) {
	    for (var i = 0; i < gr[j].length; i++) {
		if (gr[j][i].match(/[a-z]/)) { break }
		for (var ix = 0; ix < GRAMS.length; ix++) {
		    const g = GRAMS[ix];
		    if (!gramO[g]) { continue }
		    if (WORDS[g + gr[j+1][i] + gr[j+2][i]]) {
			delete gramO[g];
		    }
		}
	    }
	}

	// Wriggle out of many "lucky" horizontal words, part 2
	// Don't drop a tile which would be the middle tile in a "lucky" word
	if (j >= 1 && j < gr.length - 1) {
	    for (var i = 0; i < gr[j].length; i++) {
		if (gr[j][i].match(/[a-z]/)) { break }
		for (var ix = 0; ix < GRAMS.length; ix++) {
		    const g = GRAMS[ix];
		    if (!gramO[g]) { continue }
		    if (WORDS[gr[j-1][i] + g + gr[j+1][i]]) {
			delete gramO[g];
		    }
		}
	    }
	}

	// Wriggle out of many "lucky" horizontal words, part 3
	// Don't drop a tile which would be the last tile in a "lucky" word
	if (j >= 2) {
	    for (var i = 0; i < gr[j].length; i++) {
		if (gr[j][i].match(/[a-z]/)) { break }
		for (var ix = 0; ix < GRAMS.length; ix++) {
		    const g = GRAMS[ix];
		    if (!gramO[g]) { continue }
		    if (WORDS[gr[j-2][i] + gr[j-1][i] + g]) {
			delete gramO[g];
		    }
		}
	    }
	}

	if (Object.keys(gramO).length) {
	    gr[j][0] = randChoice(Object.keys(gramO));
	} else {
	    gr[j][0] = randChoice(GRAMS);
	}
	return true
    }
    return false
}

// Call f(i,j) repeatedly for i in [0..height) and j in [0..width)
function foreachCell(f) {
    for (var i = 0; i < GRID_H; i++) {
	for (var j = 0; j < GRID_W; j++) {
	    f(i, j);
	}
    }
}

function newGame() {
    initGrid();
    score = 0;
    scoreBoost = 1.0;
    for (var j = 0; j < GRID_W; j++) { fillOneTopCell(grid); }
    TICK_TIME = BASE_TICK_TIME / 2;
    setTimeout(animFrame, 0);
    document.getElementById('words').innerHTML = WORDS_L.join(', ');
    display();
}

function animFrame() {
    var gr2 = physicsTick(grid);
    gridj = JSON.stringify(grid);
    gr2j = JSON.stringify(gr2);
    score += Math.floor(scoreBoost * gr2j.split('!').length / 2.5);
    scoreBoost += Math.log(score + 3);
    if (gr2j.includes('!') || gr2j.includes('""') || gr2j != gridj) {
	TICK_TIME = Math.floor(0.95 * TICK_TIME);
	if (TICK_TIME < (BASE_TICK_TIME / 10)) TICK_TIME = BASE_TICK_TIME / 10;
	setTimeout(animFrame, TICK_TIME);
	ui_state = false;
    } else {
	const bombCount = gr2j.split(/\?/).length - 1;
	if (Math.random() * (bombCount + 0.25) < 0.002 * GRID_W * GRID_H) {
	    var i = Math.floor((Math.random() + Math.random()) * GRID_H / 2);
	    var j = Math.floor((Math.random() + Math.random()) * GRID_W / 2);
	    gr2[j][i] = '?';
	}
	TICK_TIME = BASE_TICK_TIME;
	scoreBoost = Math.max(1.0 * score / 200.0, 1.0);
	ui_state = true;
    }
    grid = gr2;
    display();
}

function display() {
    const gr2 = physicsTick(grid);
    var rows = [];
    for (var i = 0; i < GRID_H; i++) {
	var row = document.createElement('tr');
	for (var j = 0; j < GRID_W; j++) {
	    var td = document.createElement('td');
	    var s = grid[j][i];
	    if (s == '!') { s = '🌩'; }
	    if (s == '!!') { s = '💥'; }
	    if (s == '?') { s = '🧨'; }
	    td.appendChild(document.createTextNode(s));
	    td.onclick = cellClicked;
	    td.i = i;
	    td.j = j;
	    td.id = 'td_i_' + i + '_j_' + j;
	    if (gr2[j][i] == '!'
		|| grid[j][i].includes('!')
		|| (grid[j][i] == '?' && gr2[j][i] != '?')) {
		td.className = 'mark';
	    }
	    row.appendChild(td);
	}
	rows.push(row);
    }
    document.getElementById('gridtbody').replaceChildren(...rows);
    document.getElementById('complaints').innerHTML = '&nbsp;'
    document.getElementById('score').innerHTML = readableScore(score);
}

function readableScore(score) {
    if (score > 5000000000) {
	return `${score.toExponential(3)} points`
    }
    if (score > 5000000) {
	var s = 1.0 * score / 1000000.0;
	return `${s.toFixed(2)} million points`
    }
    if (score > 5000) {
	var s = 1.0 * score / 1000.0;
	return `${s.toFixed(2)} thousand points`
    }
    return `${score} points`
}

function complain(html) {
    document.getElementById('complaints').innerHTML = html;
}

function suggest() {
    var bestSScore = 0;
    var bestI = 0;
    var bestJ = 0;
    var bestDirection = 'b';
    // look for verticals
    foreachCell(function(i, j) {
	if (i >= GRID_H-1) { return }
	gr2 = copyGrid();
	gr2[j][i] = grid[j][i+1];
	gr2[j][i+1] = grid[j][i];
	const p = physicsTick(gr2);
	const sscore = JSON.stringify(p).split('!').length - 1;
	if (sscore && sscore + i > bestSScore) {
	    bestSScore = sscore + i;
	    bestI = i;
	    bestJ = j;
	    bestDirection = 'b';
	}
    });
    // look for horizontals
    foreachCell(function(i, j) {
	if (j >= GRID_W-1) { return }
	gr2 = copyGrid();
	gr2[j][i] = grid[j+1][i];
	gr2[j+1][i] = grid[j][i];
	const p = physicsTick(gr2);
	const sscore = JSON.stringify(p).split('!').length - 1;
	if (sscore && sscore + i > bestSScore) {
	    bestSScore = sscore + i;
	    bestI = i;
	    bestJ = j;
	    bestDirection = 'r';
	}
    });
    if (bestSScore > 0) {
	var td = document.getElementById('td_i_' + bestI + '_j_' + bestJ);
	td.className += ' suggest' + bestDirection;
    } else {
	complain("Sorry. I'm out of ideas.");
    }
}

function wider(i) {
    GRID_W += i;
    if (GRID_W < 3) { GRID_W = 3; }
    initGrid();
    newGame();
}

function taller(i) {
    GRID_H += i;
    if (GRID_H < 3) { GRID_H = 3; }
    initGrid();
    newGame();
}



function cellClicked(e) {
    if (!ui_state) { return true; }
    if (previous_tapped) {
	var delta_i = Math.abs(e.target.i - previous_tapped.i);
	var delta_j = Math.abs(e.target.j - previous_tapped.j);
	if (delta_i + delta_j != 1) {
	    previous_tapped.className = '';
	    previous_tapped = false;
	    return false;
	}
    } else {
	previous_tapped = e.target;
	previous_tapped.className = 'sel';
	return false;
    }
    var gr2 = copyGrid();
    gr2[previous_tapped.j][previous_tapped.i] = grid[e.target.j][e.target.i];
    gr2[e.target.j][e.target.i] = grid[previous_tapped.j][previous_tapped.i];
    if (JSON.stringify(physicsTick(gr2)).includes('!')) {
	grid = gr2;
	previous_tapped.className = '';
	previous_tapped = false;
	ui_state = false;
	display();
	setTimeout(animFrame, 500);
    } else {
	previous_tapped.className = '';
	previous_tapped = false;
	display();
	complain("Switching those two squares wouldn't form a word. Undoing!");
    }
}

function physicsTick(gr) {
    var nex = copyGrid();
    // Any empty cells in top row? Fill some in.
    for (var j = 0; j < Math.sqrt(GRID_W); j++) { fillOneTopCell(grid); }
    // gravity: letters above voids should "fall"
    for (var j = 0; j < GRID_W; j++) {
	for (var bottom = GRID_H-1; bottom > 0; bottom--) {
	    if(nex[j][bottom]) { continue }
	    const top = bottom-1;
	    if(!nex[j][top]) { continue }
	    if (nex[j][top] == '?') { nex[j][top] = '!!'; }
	    nex[j][bottom] = nex[j][top];
	    nex[j][top] = '';
	}
    }
    // explode firecrackers
    foreachCell(function(i, j) {
	if (gr[j][i] == '!!') {
	    for (var ii = i - 1; ii < i - 1 + 3; ii++) {
		for (var jj = j - 1; jj < j - 1 + 3; jj++) {
		    if (ii < 0) { continue; }
		    if (ii >= GRID_H) { continue; }
		    if (jj < 0) { continue; }
		    if (jj >= GRID_W) { continue; }
		    nex[jj][ii] = '!';
		    if (gr[jj][ii] == '?') {
			nex[jj][ii] = '!!';
		    }
		}
	    }
	}
    })
    // replace explosions with void
    foreachCell(function(i, j) {
	if (gr[j][i] == '!') {
	    nex[j][i] = '';
	    // tickle nearby firecrackers
	    for (var ii = i - 1; ii < i - 1 + 3; ii++) {
		for (var jj = j - 1; jj < j - 1 + 3; jj++) {
		    if (ii < 0) { continue; }
		    if (ii >= GRID_H) { continue; }
		    if (jj < 0) { continue; }
		    if (jj >= GRID_W) { continue; }
		    if (gr[jj][ii] == '?') {
			nex[jj][ii] = '!!';
		    }
		}
	    }
	}
    });
    // replace horizontal words with explosions
    foreachCell(function(i, j) {
	if (j >= GRID_W-2) { return }
	var word = gr[j][i] + gr[j+1][i] + gr[j+2][i];
	if (WORDS[word]) {
	    nex[j][i] = '!';
	    nex[j+1][i] = '!';
	    nex[j+2][i] = '!';
	}
    });
    // replace vertical words with explosions
    foreachCell(function(i, j) {
	if (i >= GRID_H-2) { return }
	var word = gr[j][i] + gr[j][i+1] + gr[j][i+2];
	if (WORDS[word]) {
	    nex[j][i] = '!';
	    nex[j][i+1] = '!';
	    nex[j][i+2] = '!';
	}
    });   
    return nex;
}
