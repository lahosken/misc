// UI states
PROMPTING = "PROMPTING";
GRADING = "GRADING";
TRIUMPH = "TRIUMPH";

var answers= ["stern", "natty", "abbey"];
var ui_state = PROMPTING;
var hist = [[], [], []];
var done = [false, false, false];
var current_guess = "";

function update_view() {
    for (var col_i = 0; col_i < 3; col_i++) {
	var html = '';
	for (var hist_i = 0; hist_i < hist[col_i].length; hist_i++) {
	    html += '<div class=row>';
	    for (var lett_i = 0; lett_i < 5; lett_i++) {
		html += '<div class="' + hist[col_i][hist_i].grade[lett_i] + '">'
		html += hist[col_i][hist_i].guess[lett_i];
		html += '</div>';
	    }
	    html += '</div>';
	}
	if (done[col_i]) {
	    html += '<center>YAY</center>';
	} else {
	    var cg = current_guess + '_____';
	    if (current_guess.length >= 5 && !PROBES[current_guess]) {
		html += '<div class="row sadguess">';
	    } else {
		html += '<div class="row">';
	    }
	    for (var lett_i = 0; lett_i < 5; lett_i++) {
		html += '<div class=blank>' + cg[lett_i] + '</div>';
	    }
	    html += '</div>';
	}
	document.getElementById('col' + col_i).innerHTML = html;
    }
}

function new_game() {
    answers = [];
    while (answers.length < 3) {
	candidate = CANDIDATES[Math.floor(Math.random() * CANDIDATES.length)]
	if (answers.find(a => a == candidate)) { continue }
	answers.push(candidate);
    }
    ui_state = PROMPTING;
    hist = [[], [], []];
    done = [false, false, false];
    current_guess = "";
    document.getElementById('newgame').style.display = 'none';
    update_view();
}

function winnow_candidates() {
    var retval = [];
    for (var col_i = 0; col_i < 3; col_i++) {
	if (done[col_i]) {
	    retval.push([]);
	    continue
	}
	c_l = CANDIDATES;
	for (var hist_i = 0; hist_i < hist[col_i].length; hist_i++) {
	    t_c_l = [];
	    for (var cl_i = 0; cl_i < c_l.length; cl_i++) {
		var guess = hist[col_i][hist_i].guess;
		var candidate = c_l[cl_i];
		if (grade_one(guess, candidate) == hist[col_i][hist_i].grade) {
		    t_c_l.push(candidate)
		}
	    }
	    c_l = t_c_l;
	}
	retval.push(c_l);
    }
    return retval;
}

function do_suggest() {
    var sugg = choose_suggestion();
    if (!sugg) { return; }
    current_guess = choose_suggestion();
    update_view();
}

function choose_suggestion() {
    // special case: If we haven't guessed yet and there are
    // 1000+ "candidates", this brute-force loop takes an unacceptable minute.
    // In that case, return the known answer:
    if (!hist[0].length) {
	return 'raise';
    }
    var winnowed_candidates = winnow_candidates();
    // if a column has just one possible answer,
    // return that answer
    for (var col_i = 0; col_i < 3; col_i++) {
	if (winnowed_candidates[col_i].length == 1) {
	    return winnowed_candidates[col_i][0];
	}
    }
    var some_probes = CANDIDATES;
    // If there are many, many candidates left, then don't consider
    // 1000+ possible probes; that could take several seconds.
    if (winnowed_candidates[0].length +
	winnowed_candidates[1].length +
	winnowed_candidates[2].length > 300) {
	some_probes = [
	    "arose", "blare", "broil",
	    "could", "crate", "craze",
	    "flare", "flour", "glory", "grate",
	    "irate", "irony", "joist", "juicy", "later",
	    "minor", "parse", "quirk", "quote", "raise",
	    "saute", "scorn", "scour", "share", "shorn",
	    "slink", "snare", "solid", "spoil", "swirl",
	    "taker", "tamer", "teary", "trade", "unzip",
	    "valet", "visor", "water",
	];
    }
    var best_probe = 'raise';
    var best_score = 0;
	
    for (var probe_i = 0; probe_i < some_probes.length; probe_i++) {
	var probe = some_probes[probe_i];
	score = 0;
	for (var col_i = 0; col_i < 3; col_i++) {
	    if (winnowed_candidates[col_i].length > 1) {
		buckets = {}
		for (var wc_i = 0; wc_i < winnowed_candidates[col_i].length; wc_i++) {
		    var candidate = winnowed_candidates[col_i][wc_i];
		    var grade = grade_one(probe, candidate);
		    if (!buckets[grade]) { buckets[grade] = 0; }
		    buckets[grade]++;
		}
		for (var key in buckets) {
		    var v = buckets[key];
		    score += v * (CANDIDATES.length - v)
		}
	    }
	}
	if (score > best_score) {
	    best_probe = probe;
	    best_score = score;
	}
    }
    return best_probe;
}

function do_grades() {
    for (var col_i = 0; col_i < 3; col_i++) {
	if (done[col_i]) { continue }
	var gr = grade_one(current_guess, answers[col_i]);
	hist[col_i].push({
	    guess: current_guess,
	    grade: gr,
	})
	if (gr == 'GGGGG') { done[col_i] = true; }
    }
    current_guess = "";
    if (done[0] && done[1] && done[2]) {
	ui_state = TRIUMPH;
	document.getElementById('newgame').style.display = 'inline';
    } else {
	ui_state = PROMPTING;
    }
    update_view();
}

function grade_one(guess, answr) {
    answr_l = answr.split('');
    var retval = ['w', 'w', 'w', 'w', 'w'];
    for (var i = 0; i < 5; i++) {
	if (guess[i] == answr_l[i]) {
	    retval[i] = 'G';
	    answr_l[i] = '_';
	}
    }
    for (var i = 0; i < 5; i++) {
      if (retval[i] != 'w') { continue }
      for (var j = 0; j < 5; j++) {
	if (guess[i] == answr_l[j]) {
	    retval[i] = 'y';
	    answr_l[j] = '_';
	    break
	}
      }
    }
    return retval.join('');
}

function do_key(e) {
    if (ui_state == GRADING) { return }
    if (ui_state == TRIUMPH) { return new_game() }
    // ui_state is PROMPTING
    if ({'Delete': true, 'Backspace': true}[e.code] && current_guess.length) {
	current_guess = current_guess.substring(0, current_guess.length-1);
    }
    if ({'?': true, '/': true}[e.key]) {
	do_suggest();
	return
    }
    var m = RegExp('Key([A-Z])').exec(e.code);
    if (m && current_guess.length < 5) {
	current_guess += m[1].toLowerCase();
    }
    if (e.key == 'Enter' && current_guess.length >= 5) {
	if (PROBES[current_guess]) {
	    ui_state = GRADING;
	    setTimeout(do_grades);
	}
    }
    update_view();
}

document.addEventListener('keyup', do_key);
window.addEventListener('load', new_game);
