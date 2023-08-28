/* LOL todo */
var ipuzx = {}
{
    // user pressed a key in a grid canvas. 
    function gridKeyDown(e) {
	console.log("this:", this);
	console.log("event:", e);
	if (!this._ipuzxGrid.ui.focSquare) { return }
	if (e.altKey) { return }
	if (e.ctrlKey) { return }
	if (e.metaKey) { return }
	const w = this._ipuzxWrangler;
	const g = this._ipuzxGrid;
	var rowIx = g.ui.focSquare.row;
	var colIx = g.ui.focSquare.col;

	function goOneSquareRight() {
	    if (!w._data.puzzle[rowIx][colIx].blocked.right) {
		g.ui.focSquare.col++;
		colIx = g.ui.focSquare.col;
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].acrossEntry;
	    }
	}
	function goOneSquareLeft() {
	    if (!w._data.puzzle[rowIx][colIx].blocked.left) {
		g.ui.focSquare.col--;
		colIx = g.ui.focSquare.col;
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].acrossEntry;
	    }
	}
	function goOneSquareDown() {
	    if (!w._data.puzzle[rowIx][colIx].blocked.down) {
		g.ui.focSquare.row++;
		rowIx = g.ui.focSquare.row;
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].downEntry;
	    }
	}
	function goOneSquareUp() {
	    if (!w._data.puzzle[rowIx][colIx].blocked.up) {
		g.ui.focSquare.row--;
		rowIx = g.ui.focSquare.row;
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].downEntry;
	    }
	}
	function goOneSquare() {
	    if (g.ui.direction == "across") {
		goOneSquareRight();
	    } else {
		goOneSquareDown();
	    }
	}
	function goOneSquareBack() {
	    if (g.ui.direction == "across") {
		goOneSquareLeft();
	    } else {
		goOneSquareUp();
	    }
	}
	function findFirstAcrossEntryAfter(excludeEntry, afterRowIx, afterColIx) {
	    if (arguments.length < 2) {
		afterRoxIx = -1;
		afterColIx = -1;
	    }
	    for (var rowIx = 0; rowIx < w._data.puzzle.length; rowIx++) {
		if (rowIx < afterRowIx) { continue }
		for (var colIx = 0; colIx < w._data.puzzle[rowIx].length; colIx++) {
		    if (rowIx == afterRowIx && colIx <= afterColIx) { continue }
		    const rv = w._data.puzzle[rowIx][colIx].acrossEntry;
		    if (!rv) { continue }
		    if (rv == excludeEntry) { continue }
		    return rv
		}
	    }
	    return false
	}
	function findFirstDownEntryAfter(excludeEntry, afterRowIx, afterColIx) {
	    if (arguments.length < 2) {
		afterRoxIx = -1;
		afterColIx = -1;
	    }
	    for (var rowIx = 0; rowIx < w._data.puzzle.length; rowIx++) {
		if (rowIx < afterRowIx) { continue }
		for (var colIx = 0; colIx < w._data.puzzle[rowIx].length; colIx++) {
		    if (rowIx == afterRowIx && colIx <= afterColIx) { continue }
		    const rv = w._data.puzzle[rowIx][colIx].downEntry;
		    if (!rv) { continue }
		    if (rv == excludeEntry) { continue }
		    if (rv.startRow != rowIx) { continue }
		    return rv
		}
	    }
	    return false
	}
	
	const blockStr = w._data.block || "#";
	if (w._data.puzzle[rowIx][colIx].cell == blockStr) {
	    return;
	}
	switch (e.key) {
	case "Tab":
	    if (e.shiftKey) {
		console.log("TODO back-tab");
	    } else {
		var nextEntry = false;
		if (g.ui.direction == "across") {
		    if (w._data.puzzle[rowIx][colIx].acrossEntry) {
			const ae = w._data.puzzle[rowIx][colIx].acrossEntry;
			nextEntry = findFirstAcrossEntryAfter(ae, ae.startRow, ae.startCol);
		    } else {
			nextEntry = findFirstAcrossEntryAfter("no entry", rowIx, colIx);
		    }
		    if (!nextEntry) {
			nextEntry = findFirstDownEntryAfter("no entry");
		    }
		    if (!nextEntry) {
			nextEntry = findFirstAcrossEntryAfter("no entry");
		    }
		    if (!nextEntry) { return }
		} else {
		    if (w._data.puzzle[rowIx][colIx].downEntry) {
			const de = w._data.puzzle[rowIx][colIx].downEntry;
			nextEntry = findFirstDownEntryAfter(de, de.startRow, de.startCol);
		    } else {
			nextEntry = findFirstDownEntryAfter("no entry", rowIx, colIx);
		    }
		    if (!nextEntry) {
			nextEntry = findFirstAcrossEntryAfter("no entry");
		    }
		    if (!nextEntry) {
			nextEntry = findFirstDownEntryAfter("no entry");
		    }
		    if (!nextEntry) { return }
		}
		g.ui.selEntry = nextEntry;
		g.ui.focSquare.row = nextEntry.startRow;
		g.ui.focSquare.col = nextEntry.startCol;
		g.ui.direction = nextEntry.direction;
	    }
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	case "Backspace":
	    g.ui.guess[rowIx][colIx] = "";
	    goOneSquareBack();
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	case "Delete":
	case " ": // spacebar
	    g.ui.guess[rowIx][colIx] = "";
	    goOneSquare();
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	case "ArrowRight":
	    if (g.ui.direction == "across") {
		goOneSquareRight();
	    } else {
		g.ui.direction = "across";
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].acrossEntry;
	    }
	    w.renderGrid(g);
	    e.preventDefault();
	    return
	case "ArrowLeft":
	    if (g.ui.direction == "across") {
		goOneSquareLeft();
	    } else {
		g.ui.direction = "across";
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].acrossEntry;
	    }
	    w.renderGrid(g);
	    e.preventDefault();
	    return
	case "ArrowUp":
	    if (g.ui.direction == "across") {
		g.ui.direction = "down";
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].downEntry;
	    } else {
		goOneSquareUp();
	    }
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	case "ArrowDown":
	    if (g.ui.direction == "across") {
		g.ui.direction = "down";
		g.ui.selEntry = w._data.puzzle[rowIx][colIx].downEntry;
	    } else {
		goOneSquareDown();
	    }
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	default:
	    // fall through
	}
	if (e.key.length == 1) {
	    g.ui.guess[rowIx][colIx] = e.key.toUpperCase();
	    goOneSquare();
	    e.preventDefault();
	    w.renderGrid(g);
	    return
	}
	console.log("tra la la", e.key);
    }
    class Wrangler {
	constructor() {
	    this._data = false;
	    this._grids = [];
	    this._acrossCluesDiv = false;
	    this._downCluesDiv = false;
	}
	get data() {
	    return this._data;
	}
	get gridCanvas() {
	    if (this._grids.length) {
		return this._grids[0].canvas;
	    } else {
		return false;
	    }
	}
	set data(d) {
	    this._data = JSON.parse(JSON.stringify(d));
	    // The JSON for each grid square might be a full struct or might just
	    // be the cell contents. We'll use a full struct for each square, so
	    // replace [ …, foo, …] with [ …, {"cell": foo}, …]
	    for (var rowIx = 0; rowIx < this._data.puzzle.length; rowIx++) {
		for (var colIx = 0; colIx < this._data.puzzle[rowIx].length; colIx++) {
		    if (typeof({}) == typeof(this._data.puzzle[rowIx][colIx])) { continue }
		    this._data.puzzle[rowIx][colIx] = { cell: this._data.puzzle[rowIx][colIx] }
		}
	    }
	    const blockStr = d.block || "#";
	    const emptyStr = this._data.empty || 0;
	    // Annotate our squares: are they blocked in each direction?
	    // Ipuz supports blocked squares and "walls" between squares.
	    for (var rowIx = 0; rowIx < this._data.puzzle.length; rowIx++) {
		for (var colIx = 0; colIx < this._data.puzzle[rowIx].length; colIx++) {
		    var blocked = {};
		    blocked.up = (rowIx <= 0 ||
				 this._data.puzzle[rowIx][colIx].cell == blockStr ||
				 this._data.puzzle[rowIx-1][colIx].cell == blockStr ||
				 ( this._data.puzzle[rowIx][colIx].barred &&
				   this._data.puzzle[rowIx][colIx].barred.includes("T") ) ||
				 ( this._data.puzzle[rowIx-1][colIx].barred &&
				   this._data.puzzle[rowIx-1][colIx].barred.includes("B") ) );
		    blocked.down = (rowIx >= this._data.puzzle.length-1 ||
				   this._data.puzzle[rowIx][colIx].cell == blockStr ||
				   this._data.puzzle[rowIx+1][colIx].cell == blockStr ||
				 ( this._data.puzzle[rowIx][colIx].barred &&
				   this._data.puzzle[rowIx][colIx].barred.includes("B") ) ||
				 ( this._data.puzzle[rowIx+1][colIx].barred &&
				   this._data.puzzle[rowIx+1][colIx].barred.includes("T") ) );
		    blocked.right = (colIx >= this._data.puzzle[rowIx].length-1 ||
				    this._data.puzzle[rowIx][colIx].cell == blockStr ||
				    this._data.puzzle[rowIx][colIx+1].cell == blockStr ||
				    ( this._data.puzzle[rowIx][colIx].barred &&
				      this._data.puzzle[rowIx][colIx].barred.includes("R") ) ||
				    ( this._data.puzzle[rowIx][colIx+1].barred &&
				      this._data.puzzle[rowIx][colIx+1].barred.includes("L") ) );
		    blocked.left = (colIx <= 0 ||
				    this._data.puzzle[rowIx][colIx].cell == blockStr ||
				    this._data.puzzle[rowIx][colIx-1].cell == blockStr ||
				    ( this._data.puzzle[rowIx][colIx].barred &&
				      this._data.puzzle[rowIx][colIx].barred.includes("L") ) ||
				    ( this._data.puzzle[rowIx][colIx-1].barred &&
				      this._data.puzzle[rowIx][colIx-1].barred.includes("R") ) );
		    this._data.puzzle[rowIx][colIx].blocked = blocked;
		}
	    }
	    var acrossEntries = [];
	    for (var rowIx = 0; rowIx < this._data.puzzle.length; rowIx++) {
		for (var colIx = 0; colIx < this._data.puzzle[rowIx].length-1; colIx++) {
		    if (this._data.puzzle[rowIx][colIx].cell == blockStr) { continue }
		    // Are we in the first square of an across-entry?
		    //   If so, we should be blocked on the left, but not on the right:
		    if (!this._data.puzzle[rowIx][colIx].blocked.left) { continue }
		    if (this._data.puzzle[rowIx][colIx].blocked.right) { continue }
		    var entry = {
			direction: "across",
			startRow: rowIx,
			endRow: rowIx,
			startCol: colIx,
		    }
		    if (this._data.puzzle[rowIx][colIx].cell && this._data.puzzle[rowIx][colIx].cell != emptyStr) {
			entry.label = String(this._data.puzzle[rowIx][colIx].cell);
		    }
		    for (var endColIx = colIx; endColIx < this._data.puzzle[rowIx].length; endColIx++) {
			this._data.puzzle[rowIx][endColIx].acrossEntry = entry;
			if (!this._data.puzzle[rowIx][endColIx].blocked.right) { continue }
			entry.endCol = endColIx;
			acrossEntries.push(entry);
			break
		    }
		}
	    }
	    if (acrossEntries.length) {
		this._data.acrossEntries = acrossEntries;
	    }
	    var downEntries = [];
	    for (var rowIx = 0; rowIx < this._data.puzzle.length-1; rowIx++) {
		for (var colIx = 0; colIx < this._data.puzzle[rowIx].length; colIx++) {
		    if (this._data.puzzle[rowIx][colIx].cell == blockStr) { continue }
		    // Are we in the first square of an across-entry?
		    //   If so, we should be blocked on top, but not underneath
		    if (!this._data.puzzle[rowIx][colIx].blocked.up) { continue }
		    if (this._data.puzzle[rowIx][colIx].blocked.down) { continue }
		    var entry = {
			dir: "down",
			startRow: rowIx,
			startCol: colIx,
			endCol: colIx,
		    }
		    if (this._data.puzzle[rowIx][colIx].cell && this._data.puzzle[rowIx][colIx].cell != emptyStr) {
			entry.label = String(this._data.puzzle[rowIx][colIx].cell);
		    }
		    for (var endRowIx = rowIx; endRowIx < this._data.puzzle.length; endRowIx++) {
			this._data.puzzle[endRowIx][colIx].downEntry = entry;
			if (!this._data.puzzle[endRowIx][colIx].blocked.down) { continue }
			entry.endRow = endRowIx;
			downEntries.push(entry);
			break
		    }
		}
	    }
	    if (downEntries.length) {
		this._data.downEntries = downEntries;
	    }
	    if (acrossEntries.length) {
		this._grids.forEach((g) => g.ui = this.initUI());
	    }
	    this.renderGrids();
	}
	addGridCanvas(c) {
	    if (!c) { return }
	    if (!c.hasAttribute('tabindex')) {
		c.setAttribute('tabindex', 0);
	    }
	    var g = {
		canvas: c,
		ui: this.initUI(),
	    };
	    c.addEventListener("keydown", gridKeyDown);
	    c._ipuzxWrangler = this;
	    c._ipuzxGrid = g;
	    this._grids.push(g);
	    this.renderGrids();
	}
	set gridCanvas(c) {
	    this._grids = [];
	    this.addGridCanvas(c)
	}
	renderGrids() {
	    this._grids.forEach((g) => this.renderGrid(g))
	}
	initUI() {
	    var ui = {
		selEntry: false,
		focSquare: false,
		direction: "across",
		guess: [],
	    };
	    if (this._data.puzzle && this._data.puzzle.length) {
		for (var rowIx = 0; rowIx < this._data.puzzle.length; rowIx++) {
		    var row = [];
		    for (var colIx = 0; colIx < this._data.puzzle[rowIx].length; colIx++) {
			row.push("");
		    }
		    ui.guess.push(row);
		}
	    }
	    if (this._data.acrossEntries && this._data.acrossEntries.length) {
		ui.selEntry = this._data.acrossEntries[0];
		ui.focSquare = 	{
		    row: ui.selEntry.startRow,
		    col: ui.selEntry.startCol,
		}
	    }
	    return ui
	}
	renderGrid(g) {
	    var context = g.canvas.getContext("2d");
	    if (!context) { return; }
	    if (!this._data) { return; }
	    const sqSz = Math.floor(Math.min(g.canvas.width / this._data.dimensions.width,
					     g.canvas.height / this._data.dimensions.height))
	    const xOffset = Math.floor((g.canvas.width - sqSz *  this._data.dimensions.width) / 2)
	    const yOffset = Math.floor((g.canvas.height - sqSz *  this._data.dimensions.height) / 2)
	    const blockStr = this._data.block || "#";
	    const emptyStr = this._data.empty || 0;	    
	    for (var rowIx = 0; rowIx < this._data.dimensions.height; rowIx++) {
		for (var colIx = 0; colIx < this._data.dimensions.width; colIx++) {
		    const sq = this._data.puzzle[rowIx][colIx];
		    var soln = {};
		    var style = {};
		    if (sq.style) { style = sq.style; }
		    if (typeof(sq.style) == typeof("string")) {
			style =  this._data.styles[sq.style];
		    }

		    context.beginPath();
		    context.rect(colIx * sqSz + xOffset, rowIx * sqSz + yOffset, sqSz, sqSz);
		    context.closePath();
		    var bgFillStyle = "rgb(0, 0, 0)";
		    if (style.colorbar) {
			bgFillStyle = "#" + style.colorbar;
		    }
		    if (sq.cell != blockStr) {
			bgFillStyle = style.color || "rgb(255, 255, 255)";
			if (style.color) {
			    bgFillStyle = "#" + style.color;
			}
		    }
		    if (g.ui.selEntry &&
			rowIx >= g.ui.selEntry.startRow &&
			rowIx <= g.ui.selEntry.endRow &&
			colIx >= g.ui.selEntry.startCol &&
			colIx <= g.ui.selEntry.endCol) {
			bgFillStyle = "rgb(200, 255, 255)";
		    }
		    if (g.ui.focSquare &&
			rowIx == g.ui.focSquare.row &&
			colIx == g.ui.focSquare.col) {
			bgFillStyle = "rgb(200, 255, 200)";
		    }
		    context.fillStyle = bgFillStyle;
		    context.fill();
		    context.lineWidth = 1;
		    context.strokeStyle = "rgb(0, 0, 0)";
		    context.stroke();

		    if (style && style.shapebg) {
			if (style.shapebg == "circle") {
			    context.beginPath();
			    context.arc((colIx + 0.5) * sqSz + xOffset,
					(rowIx + 0.5) * sqSz + yOffset,
					sqSz/2,
					0, 2 * Math.PI);
			    context.closePath();
			    context.strokeStyle = "rgb(0, 0, 0)";
			    context.stroke();
			}
		    }

		    if (sq.style && sq.style.barred) {
			context.lineWidth = 3;
			context.strokeStyle = "rgb(0, 0, 0)";
			context.beginPath();
			if (sq.style.barred.includes("T")) {
			    context.moveTo(colIx * sqSz + xOffset+1, rowIx * sqSz + yOffset);
			    context.lineTo(colIx * sqSz + xOffset+sqSz-2, rowIx * sqSz + yOffset);
			}
			if (sq.style.barred.includes("B")) {
			    // TODO: untested, Crossword Compiler only uses "T" and "L", and
			    //  I created my test xwd in Crossword Compiler
			    context.moveTo(colIx * sqSz + xOffset+1, rowIx * sqSz + yOffset+sqSz);
			    context.lineTo(colIx * sqSz + xOffset+sqSz-2, rowIx * sqSz + yOffset+sqSz);
			}
			if (sq.style.barred.includes("L")) {
			    context.moveTo(colIx * sqSz + xOffset, rowIx * sqSz + yOffset+1);
			    context.lineTo(colIx * sqSz + xOffset, rowIx * sqSz + yOffset+sqSz-2);
			}
			if (sq.style.barred.includes("R")) {
			    // TODO: untested, Crossword Compiler only uses "T" and "L", and
			    //  I created my test xwd in Crossword Compiler
			    context.moveTo(colIx * sqSz + xOffset+sqSz, rowIx * sqSz + yOffset+1);
			    context.lineTo(colIx * sqSz + xOffset+sqSz, rowIx * sqSz + yOffset+sqSz-2);
			}
			context.closePath();
			context.stroke();
		    }

		    if (sq.cell &&
			sq.cell != blockStr &&
			sq.cell != emptyStr) {
			context.fillStyle = bgFillStyle;
			context.fillRect(colIx * sqSz + xOffset + 2, rowIx * sqSz + yOffset + 2, sqSz/3, sqSz/3);
			context.font = "" + Math.floor(sqSz/4) + "px serif";
			context.fillStyle = "rgb(0, 0, 0)";
			context.fillText(String(sq.cell), colIx * sqSz + xOffset+2, (rowIx + 0.25) * sqSz + yOffset);
		    }
		    if (g.ui.guess[rowIx][colIx]) {
			const guess = g.ui.guess[rowIx][colIx];
			context.fillStyle = "rgb(0, 0, 0)";
			var fontSize = sqSz * 4 / 5;
			var m = {}
			while (fontSize > 5) {
			    fontSize = fontSize * 0.95;
			    context.font = "" + fontSize + "px sans-serif"
			    m = context.measureText(String(guess));
			    if (m.width < sqSz * 4 / 5) { break }
			}
			const a = m.actualBoundingBoxAscent;
			const d = Math.max(0, m.actualBoundingBoxDescent);
			context.fillText(String(guess),
					 colIx * sqSz + (sqSz - m.width)/2,
					 rowIx * sqSz + (sqSz + a -d)/2 + 1);
		    }
		}
	    }
	}
    }
    ipuzx.Factory = function () { return new Wrangler(); }
}
