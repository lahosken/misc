/* LOL todo */
var ipuzx = {}
{
    class Wrangler {
	constructor() {
	    this._data = false;
	    this._gridCanvas = false;
	    this._acrossCluesDiv = false;
	    this._downCluesDiv = false;

	    this._gridCtx = false; // TODO: why not just make a new context each render?

	    this._ui = {
		"entry": false,
		"focus": false,
	    }
	}
	get data() {
	    return this._data;
	}
	get gridCanvas() {
	    return this._gridCanvas;
	}
	set data(d) {
	    this._data = JSON.parse(JSON.stringify(d));
	    // The JSON for each grid square might be a full struct or might just
	    // be the cell contents. We'll use a full struct for each square, so
	    // replace [ …, foo, …] with [ …, {"cell": foo}, …]
	    for (var rowIx = 0; rowIx < this._data.puzzle.length; rowIx++) {
		for (var colIx = 0; colIx < this._data.puzzle[rowIx].length; colIx++) {
		    if (typeof({}) == typeof(this._data.puzzle[rowIx][colIx])) { continue }
		    this._data.puzzle[rowIx][colIx] = { "cell": this._data.puzzle[rowIx][colIx] }
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
			"dir": "across",
			"startRow": rowIx,
			"endRow": rowIx,
			"startCol": colIx,
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
		this._ui.entry = acrossEntries[0];
		this._ui.focus = {
		    "row": this._ui.entry.startRow,
		    "col": this._ui.entry.startCol,
		}
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
			"dir": "down",
			"startRow": rowIx,
			"startCol": colIx,
			"endCol": colIx,
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
	    this.renderGrid();
	}
	set gridCanvas(c) {
	    this._gridCanvas = c;
	    this._gridCtx = c.getContext('2d');
	    this.renderGrid();
	}
	renderGrid() {
	    var context = this._gridCtx;
	    if (!context) { return; }
	    if (!this._data) { return; }
	    const sqSz = Math.floor(Math.min(this._gridCanvas.width / this._data.dimensions.width,
					     this._gridCanvas.height / this._data.dimensions.height))
	    const xOffset = Math.floor((this._gridCanvas.width - sqSz *  this._data.dimensions.width) / 2)
	    const yOffset = Math.floor((this._gridCanvas.height - sqSz *  this._data.dimensions.height) / 2)
	    const blockStr = this._data.block || "#";
	    const emptyStr = this._data.empty || 0;	    
	    for (var rowIx = 0; rowIx < this._data.dimensions.height; rowIx++) {
		for (var colIx = 0; colIx < this._data.dimensions.width; colIx++) {
		    const sq = this._data.puzzle[rowIx][colIx];
		    context.beginPath();
		    context.rect(colIx * sqSz + xOffset , rowIx * sqSz + yOffset, sqSz, sqSz);		    
		    context.closePath();
		    var bgFillStyle = "rgb(0, 0, 0)";
		    if (sq.cell != blockStr) {
			bgFillStyle = "rgb(255, 255, 255)";
		    }
		    if (this._ui.entry &&
			rowIx >= this._ui.entry.startRow &&
			rowIx <= this._ui.entry.endRow &&
			colIx >= this._ui.entry.startCol &&
			colIx <= this._ui.entry.endCol) {
			bgFillStyle = "rgb(200, 255, 255)";
		    }
		    if (this._ui.focus &&
			rowIx == this._ui.focus.row &&
			colIx == this._ui.focus.col) {
			bgFillStyle = "rgb(200, 255, 200)";
		    }
		    context.fillStyle = bgFillStyle;
		    context.fill();
		    context.lineWidth = 1;
		    context.strokeStyle = "rgb(0, 0, 0)";
		    context.stroke();

		    if (sq.style && sq.style.shapebg) {
			if (sq.style.shapebg == "circle") {
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

		    // TODO: I haven't tested this at all, realized I didn't have a
		    // test puzzle with walls instead of block-squares
		    if (sq.style && sq.style.barred) {
			context.lineWidth = 3;
			context.strokeStyle = "rgb(0, 0, 0)";
			context.beginPath();
			if (sq.style.barred.includes("T")) {
			    context.moveTo(colIx * sqSz + xOffset+1, rowIx * sqSz + yOffset);
			    context.lineTo(colIx * sqSz + xOffset+sqSz-2, rowIx * sqSz + yOffset);
			}
			if (sq.style.barred.includes("B")) {
			    context.moveTo(colIx * sqSz + xOffset+1, rowIx * sqSz + yOffset+sqSz);
			    context.lineTo(colIx * sqSz + xOffset+sqSz-2, rowIx * sqSz + yOffset+sqSz);
			}
			if (sq.style.barred.includes("L")) {
			    context.moveTo(colIx * sqSz + xOffset, rowIx * sqSz + yOffset+1);
			    context.lineTo(colIx * sqSz + xOffset, rowIx * sqSz + yOffset+sqSz-2);
			}
			if (sq.style.barred.includes("R")) {
			    context.moveTo(colIx * sqSz + xOffset+sqSz, rowIx * sqSz + yOffset+1);
			    context.lineTo(colIx * sqSz + xOffset+sqSz, rowIx * sqSz + yOffset+sqSz-2);
			}
			context.closePath();
			context.stroke();
		    }

		    if (sq.cell != blockStr &&
			sq.cell != emptyStr) {
			context.beginPath();
			context.rect(colIx * sqSz + xOffset + 2, rowIx * sqSz + yOffset + 2, sqSz/3, sqSz/3);		    
			context.closePath();
			context.fillStyle = bgFillStyle;
			context.fill();
			context.font = "" + Math.floor(sqSz/4) + "px serif";
			context.fillStyle = "rgb(0, 0, 0)";
			context.fillText(String(sq.cell), colIx * sqSz + xOffset+2, (rowIx + 0.25) * sqSz + yOffset);
		    }
		}
	    }
	}
    }
    ipuzx.Factory = function () { return new Wrangler(); }
}
