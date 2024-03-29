/* (c) 2023 Kenny Young
 * This code is licensed under the MIT License.
 * https://github.com/tabascq/PuzzleJS
 */
var puzzleJsFolderPath = document.currentScript.src.replace("puzzle.js", "");

// register some puzzle modes; a mode is just a set of options,
// so the options do not need to all be learned and manually applied to each puzzle.
// a puzzle can have multiple modes and multiple custom options.
var puzzleModes = {};
puzzleModes["default"] = {
    // text
    "data-text": null,
    "data-text-replacements": null,
    "data-text-characters": "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
    "data-text-shift-key": "rebus",
    "data-text-shift-lock": false,
    "data-text-solution": null,
    "data-text-advance-on-type": false,

    // fills
    "data-fill-classes": null,
    "data-fills": null,
    "data-fill-cycle": true,

    // paths
    "data-paths": null,
    "data-path-style": "straight",

    // edges
    "data-edges": null,
    "data-edge-style": "box",

    // clues
    "data-clue-locations": null,
    "data-clue-indicators": null,
    "data-top-clues": null,
    "data-bottom-clues": null,
    "data-left-clues": null,
    "data-right-clues": null,

    // misc
    "data-drag-paint-fill": true,
    "data-drag-draw-path": false,
    "data-drag-draw-edge": false,
    "data-unselectable-givens": false,
    "data-extracts": null,
    "data-no-input": false,
    "data-show-commands": false,
    "data-state-key": null
};

puzzleModes["linear"] = {
    "data-text-advance-on-type": true,
    "data-unselectable-givens": true
}

puzzleModes["crossword"] = {
    "data-text-advance-on-type": true,
    "data-clue-locations": "crossword"
};

puzzleModes["notext"] = {
    "data-text-characters": ""
}

puzzleModes["sudoku"] = {
    "data-text-characters": "123456789",
    "data-edges": "3x3",
    "data-text": "9x9",
    "data-text-shift-key": "candidates"
};

puzzleModes["pathpaint"] = {
    "data-path-style": "curved",
    "data-drag-draw-path": true,
    "data-fill-cycle": false
}

puzzleModes["trains"] = {
    "data-path-style": "track",
    "data-drag-paint-fill": false,
    "data-fill-cycle": false,
    "data-drag-draw-path": true
}

puzzleModes["slitherlink"] = {
    "data-drag-draw-edge" : true,
    "data-edge-style": "dots"
}

puzzleModes["solution"] = {
    "data-no-input": true
}

// Parse string as raw JS objects. e.g. "false" -> false
// (if ("false") is truthy in JS)
function parseFalseStrings(s) {
    return s.toLowerCase() === "false" ? false : s;
}

// Go through all puzzles and give them a PuzzleEntry object
document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll(".puzzle-entry").forEach((p, index) => { new PuzzleEntry(p, index); });
});

function resetAllPuzzleStateOnPage() {
    document.querySelectorAll(".puzzle-entry").forEach((p) => { p.puzzleEntry.prepareToReset(); });
    window.location.reload();
}

function UndoManager() {
    this.undoStack = [];
    this.redoStack = [];
    this.activeGroup = null;

    // undo/redo support
    this.redoUnit = function(unit) {
        var extractId = unit.elem.getAttribute("data-extract-id");
        var elems = extractId ? document.querySelectorAll("." + extractId) : [unit.elem];

        elems.forEach(elem =>{
            if (unit.adds) { unit.adds.forEach((a) => { elem.classList.add(a); }); }
            if (unit.removes) { unit.removes.forEach((a) => { elem.classList.remove(a); }); }
            if (unit.attribute) { elem.setAttribute(unit.attribute, unit.newValue); }
            if (unit.oldText || unit.newText) { elem.innerText = unit.newText; }
        });
    }

    this.undoUnit = function(unit) {
        var extractId = unit.elem.getAttribute("data-extract-id");
        var elems = extractId ? document.querySelectorAll("." + extractId) : [unit.elem];

        elems.forEach(elem =>{
            if (unit.adds) { unit.adds.forEach((a) => { elem.classList.remove(a); }); }
            if (unit.removes) { unit.removes.forEach((a) => { elem.classList.add(a); }); }
            if (unit.attribute) { elem.setAttribute(unit.attribute, unit.oldValue); }
            if (unit.oldText || unit.newText) { elem.innerText = unit.oldText; }
        });
    }

    this.undo = function() {
        if (this.activeGroup) { this.endGroup(); }
        if (this.undoStack.length == 0) { return; }

        var group = this.undoStack.pop();
        group.units.forEach((unit) => { this.undoUnit(unit); });
        this.redoStack.push(group);
        this.notify(group);
    }

    this.redo = function() {
        if (this.activeGroup) { this.endGroup(); }
        if (this.redoStack.length == 0) { return; }

        var group = this.redoStack.pop();
        group.units.forEach((unit) => { this.redoUnit(unit); });
        this.undoStack.push(group);
        this.notify(group);
    }

    this.startGroup = function(puzzleEntry) {
        if (this.activeGroup) { this.endGroup(); }
        this.activeGroup = { puzzleEntry: puzzleEntry, units: [] };
    }

    this.endGroup = function() {
        var retVal = false;

        if (this.activeGroup && this.activeGroup.units.length) {
            this.undoStack.push(this.activeGroup);
            this.redoStack = [];
            this.notify(this.activeGroup);
            retVal = true;
        }

        this.activeGroup = null;
        return retVal;
    }

    this.modifyClass = function(elem, adds, removes) {
        var trueAdds = [];
        var trueRemoves = [];

        adds.forEach((a) => { if (a && !elem.classList.contains(a)) { trueAdds.push(a); }});
        removes.forEach((r) => { if (r && elem.classList.contains(r)) { trueRemoves.push(r); }});

        if (trueAdds.length == 0 && trueRemoves.length == 0) { return; }

        var unit = { "elem": elem };
        if (trueAdds.length > 0) { unit.adds = trueAdds; }
        if (trueRemoves.length > 0) { unit.removes = trueRemoves; }

        this.redoUnit(unit);
        this.activeGroup.units.push(unit);
    }

    this.modifyAttribute = function(elem, attribute, newValue) {
        var oldValue = elem.getAttribute(attribute);
        if (oldValue == newValue) { return; }

        var unit = { "elem": elem, "attribute": attribute, "oldValue": oldValue, "newValue": newValue };

        this.redoUnit(unit);
        this.activeGroup.units.push(unit);
    }

    this.modifyText = function(elem, newText) {
        var oldText = elem.innerText;
        if (oldText == newText) { return; }

        var unit = { "elem": elem, "oldText": oldText, "newText": newText };

        this.redoUnit(unit);
        this.activeGroup.units.push(unit);
    }

    this.notify = function(group) {
        group.puzzleEntry.onUndoRedo(group.units);
    }
}

function PuzzleEntry(p, index) {
    this.container = p;
    p.puzzleEntry = this;

    // Assign all options by applying all properties from all modes. Modes specified earliest in data-mode get precedence.
    var modes = p.getAttribute("data-mode");
    modes = modes ? modes.split(" ") : [];
    modes.push("default"); 
    modes.reverse();

    this.options = {};
    modes.forEach(m => { for (const [key, value] of Object.entries(puzzleModes[m])) { this.options[key] = value; } });

    // Finally, any explicitly-specified attributes win.
    for (const [key, value] of Object.entries(this.options)) {
        if (this.container.hasAttribute(key)) { this.options[key] = parseFalseStrings(this.container.getAttribute(key)); }
    }

    if (this.container.firstChild && this.container.firstChild.nodeType === Node.TEXT_NODE) {
        try {
            var json = JSON.parse(this.container.firstChild.textContent);
            for (const[key, value] of Object.entries(json)) { this.options[key] = value; }
            this.container.removeChild(this.container.firstChild);
        } catch {}
    }

    this.pointerIsDown = false;
    this.lastCell = null;
    this.currentFill = null;
    this.stateKey = this.options["data-state-key"];
    if (!this.stateKey) { this.stateKey = window.location.href + "|" + index; }
    this.inhibitSave = false;
    this.xKeyMode = false;

    this.locateScope = function(scopeId) {
        var ancestor = this.container;

        while (ancestor) {
            if (ancestor.getAttribute(scopeId) != undefined) { return ancestor; }
            ancestor = ancestor.parentElement;
        }

        return this.container;
    }

    var undoScope = this.locateScope("data-undo-scope");
    this.undoManager = undoScope.puzzleUndoManager;
    if (!this.undoManager) { this.undoManager = new UndoManager(); undoScope.puzzleUndoManager = this.undoManager; }

    // Assume that if a button with class 'clipboard-button' exists, we're using copyjack.
    this.isUsingCopyjack = document.querySelector('button.clipboard-button') !== null;

    // --- Functions to update state ---
    // keyboard support
    this.move = function(td, drow, dcol) {
        var col = dcol + Array.prototype.indexOf.call(td.parentElement.children, td) - this.leftClueDepth;
        var row = drow + Array.prototype.indexOf.call(td.parentElement.parentElement.children, td.parentElement) - this.topClueDepth;

        if (this.fShift && this.options["data-drag-draw-path"]) {
            var tdTo = this.table.querySelector("tr:nth-child(" + (row + this.topClueDepth + 1) + ") td:nth-child(" + (col + this.leftClueDepth + 1) + ")");
            if (tdTo && tdTo.classList.contains("inner-cell")) {
                this.lastCell = td;
                this.currentFill = this.findClassInList(td, this.fillClasses);
                this.dragPaintAndPath(tdTo);
            }
        }

        while (true) {
            var td = this.table.querySelector("tr:nth-child(" + (row + this.topClueDepth + 1) + ") td:nth-child(" + (col + this.leftClueDepth + 1) + ")");

            if (!td) {
                return false;
            }

            var text = td.querySelector(".text span");

            if (text && !td.classList.contains("unselectable")) {
                if (this.options["data-text-advance-on-type"]) {
                    this.dx = Math.abs(dcol);
                    this.dy = Math.abs(drow);
                }
                this.updateCenterFocus(td);
                return true;
            }

            col += dcol;
            row += drow;
        }
    }

    this.moveCorner = function(dy, dx) {
        var newX = dx + this.cornerFocusX;
        var newY = dy + this.cornerFocusY;

        newX = Math.max(newX, 0);
        newX = Math.min(newX, this.numCols);
        newY = Math.max(newY, 0);
        newY = Math.min(newY, this.numRows);

        if (this.fShift && (this.cornerFocusX != newX || this.cornerFocusY != newY)) {
            var xMin = Math.min(this.cornerFocusX, newX);
            var xMax = Math.max(this.cornerFocusX, newX);
            var yMin = Math.min(this.cornerFocusY, newY);
            var yMax = Math.max(this.cornerFocusY, newY);
            var xCell = Math.min(xMin, this.numCols - 1);
            var yCell = Math.min(yMin, this.numRows - 1);
            var edgeState = { cell: this.table.children[this.topClueDepth + yCell].children[this.leftClueDepth + xCell], edgeCode: 0 };

            if (xMax != xMin) { edgeState.edgeCode |= (yCell == yMax ? 1 : 2); }
            if (yMax != yMin) { edgeState.edgeCode |= (xCell == xMax ? 4 : 8); }
            this.lastEdgeState = null;
            this.setEdgeState(edgeState, this.xKeyMode ? "toggle-x" : "toggle-line");
            this.lastEdgeState = null;
        }

        this.cornerFocusX = newX;
        this.cornerFocusY = newY;
        this.updateCornerFocus();
    }

    this.findClassInList = function(td, classes) {
        var cls = "";
        if (classes) { classes.forEach(c => { if (td.classList.contains(c)) cls = c; }) }
        return cls;
    }

    this.cycleClasses = function(td, classes, reverse) {
        var cls = this.findClassInList(td, classes);

        if (cls) {
            this.undoManager.startGroup(this);
            cls = classes[(classes.indexOf(cls) + classes.length + (reverse ? -1 : 1)) % classes.length];
            this.setClassInCycle(td, classes, cls);
            this.undoManager.endGroup();
        }

        return cls;
    }

    this.setClassInCycle = function(td, classes, cls) {
        if (classes && !td.classList.contains(cls)) {
            this.undoManager.modifyClass(td, [cls], classes);
        }
    }

    this.handleEventChar = function(e, ch) {
        if (e.shiftKey || this.options["data-text-shift-lock"]) {
            var val = this.getText(e.target).replace("\xa0", " ");
            if (this.options["data-text-shift-key"] == "rebus" || !val.includes(ch)) { val = val + ch; }
            else { val = val.replace(ch, ""); }

            this.setText(e.target, ["small-text"], [], val);
        } else {
            this.setText(e.target, [], ["small-text"], ch);
            if (this.options["data-text-advance-on-type"]) { this.move(e.target, this.dy, this.dx); }
        }
    }

    this.handleBackspaceChar = function(e) {
        var newVal = "";

        if ((e.shiftKey || this.options["data-text-shift-lock"]) && this.options["data-text-shift-key"] == "rebus") {
            newVal = this.getText(e.target).replace("\xa0", " ");
            newVal = newVal.substring(0, newVal.length - 1);
            if (newVal.length && newVal[newVal.length - 1] == " ") {
                newVal = newVal.substring(0, newVal.length - 1) + "\xa0";
            }
        }

        if (newVal) {
            this.setText(e.target, [], [], newVal);
        } else {
            this.setText(e.target, [], ["small-text"], "");
            if (this.options["data-text-advance-on-type"]) { this.move(e.target, -this.dy, -this.dx); }
        }
    }

    this.beforeInput = function(e) {
        e.target.dispatchEvent(new KeyboardEvent("keydown", { keyCode: (e.data ? e.data.toUpperCase().charCodeAt(0) : 46) }));
        e.preventDefault();
    }

    this.keyDown = function(e) {
        this.fShift = e.shiftKey;

        if (e.keyCode == 9) return;

        e.preventDefault();
        if (this.options["data-text-solution"] || this.options["data-no-input"]) { return; }
        
        if (e.ctrlKey && e.keyCode == 90) { this.undoManager.undo(); } // Ctrl-Z
        else if (e.ctrlKey && e.keyCode == 89) { this.undoManager.redo(); } // Ctrl-Y
        else if (e.keyCode == 37) { this.move(e.target, 0, -1); } // left
        else if (e.keyCode == 38) { this.move(e.target, -1, 0); } // up
        else if (e.keyCode == 39) { this.move(e.target, 0, 1); } // right
        else if (e.keyCode == 40) { this.move(e.target, 1, 0); } // down
        else if (e.keyCode == 190 && this.canHaveCornerFocus) { this.setCornerFocusMode(); } // period
        else if (e.keyCode == 32) { // space
            if (e.ctrlKey) {
                e.currentTarget.classList.toggle("interesting");
            } else if (this.options["data-text-characters"].includes(" ")) {
                this.handleEventChar(e, "\xa0");
            } else {
                if (this.options["data-text-advance-on-type"] && this.numCols > 1 && this.numRows > 1) { this.dx = 1 - this.dx; this.dy = 1 - this.dy; }
                if (this.options["data-clue-locations"]) { this.unmark(e.target); this.mark(e.target); }
                if (e.currentTarget.classList.contains("given-fill")) return;
                if (this.options["data-fill-cycle"]) { this.currentFill = this.cycleClasses(e.target, this.fillClasses, e.shiftKey); }
            }
        } else if (e.keyCode == 8) { // backspace
            this.handleBackspaceChar(e);
        } else if (e.keyCode == 46) { // delete
            this.setText(e.target, [], [], "");
        } else {
            var code = e.keyCode;
            if (code >= 96 && code <= 105) { code -= 48; }
            var ch = String.fromCharCode(code);

            if (this.options["data-text-characters"].includes(ch)) {
                this.handleEventChar(e, ch);
            }
        }
    }

    this.keyDownCorner = function(e) {
        this.fShift = e.shiftKey;

        if (e.keyCode == 9) return;

        e.preventDefault();
        if (this.options["data-no-input"]) { return; }
        
        if (e.ctrlKey && e.keyCode == 90) { this.undoManager.undo(); } // Ctrl-Z
        else if (e.ctrlKey && e.keyCode == 89) { this.undoManager.redo(); } // Ctrl-Y
        else if (e.keyCode == 37) { this.moveCorner(0, -1); } // left
        else if (e.keyCode == 38) { this.moveCorner(-1, 0); } // up
        else if (e.keyCode == 39) { this.moveCorner(0, 1); } // right
        else if (e.keyCode == 40) { this.moveCorner(1, 0); } // down
        else if (e.keyCode == 190 && this.canHaveCenterFocus) { this.setCenterFocusMode(); } // period
        else if (e.keyCode == 88) { this.xKeyMode = !this.xKeyMode; this.cornerFocus.classList.toggle("x-mode"); } // toggle "x" mode
    }

    this.setText = function(target, adds, removes, text) {
        var textElement = target.querySelector(".text span");
        if (textElement.innerText != text && !target.classList.contains("given-text")) {
            this.undoManager.startGroup(this);
            this.undoManager.modifyClass(target, adds, removes);
            this.undoManager.modifyText(textElement, text);
            this.undoManager.endGroup();
        }
    }

    this.getText = function(target) {
        return target.querySelector(".text span").innerText;
    }

    this.onUndoRedo = function(units) {
        units.forEach((u) => {
            if (u.attribute == "data-path-code" || u.attribute == "data-edge-code" || u.attribute == "data-x-edge-code") {
                this.updateSvg(u.elem);
            }
            if (u.elem instanceof HTMLTableCellElement) {
                this.processTdForCopyjack(u.elem);
            } else if (u.elem.parentElement.parentElement instanceof HTMLTableCellElement) {
                this.processTdForCopyjack(u.elem.parentElement.parentElement);
            }
        });
    }

    this.getEventEdgeState = function(e) {
        var tolerance = e.currentTarget.offsetWidth/5;
        var cell = e.currentTarget;
        var closeTop = (e.offsetY <= tolerance);
        var closeBottom = (e.offsetY >= e.currentTarget.offsetHeight - tolerance);
        var closeLeft = (e.offsetX <= tolerance);
        var closeRight = (e.offsetX >= e.currentTarget.offsetWidth - tolerance);

        if (closeBottom || closeRight) {
            var col = Array.prototype.indexOf.call(cell.parentElement.children, cell) - this.leftClueDepth;
            var row = Array.prototype.indexOf.call(cell.parentElement.parentElement.children, cell.parentElement) - this.topClueDepth;

            if (closeBottom && row < this.numRows - 1) { closeBottom = false; closeTop = true; row++; }
            if (closeRight && col < this.numCols - 1) { closeRight = false; closeLeft = true; col++; }

            cell = cell.parentElement.parentElement.children[row + this.topClueDepth].children[col + this.leftClueDepth];
        }

        var any = closeLeft || closeRight || closeTop || closeBottom;

        // edgecode is a "four-bit integer":
        //  - The rightmost bit is 1 iff the top border is shaded.
        //  - The 2nd-to-rightmost bit is 1 iff the bottom border is shaded.
        //  - The 2nd-to-leftmost bit is 1 iff the left border is shaded.
        //  - The leftmost bit is 1 iff the right border is shaded.
        var edgeCode = 0;
        if (closeTop && !closeLeft && !closeRight) { edgeCode = 1; }
        else if (closeBottom && !closeLeft && !closeRight) { edgeCode = 2; }
        else if (closeLeft && !closeTop && !closeBottom) { edgeCode = 4; }
        else if (closeRight && !closeTop && !closeBottom) { edgeCode = 8; }

        if (edgeCode == 0 && (closeTop || closeBottom || closeLeft || closeRight)) {
            var col = Array.prototype.indexOf.call(cell.parentElement.children, cell) - this.leftClueDepth;
            var row = Array.prototype.indexOf.call(cell.parentElement.parentElement.children, cell.parentElement) - this.topClueDepth;
            if (this.canHaveCornerFocus) {
                this.setCornerFocusMode();
                this.cornerFocusX = col + (closeRight ? 1 : 0);
                this.cornerFocusY = row + (closeBottom ? 1 : 0);
                this.updateCornerFocus();
            }
        }

        return { cell: cell, edgeCode: edgeCode, any: any };
    }

    this.setEdgeState = function(edgeState, mode) {
        if (edgeState.edgeCode == 0) return;
        if (this.lastEdgeState != null && this.lastEdgeState.cell === edgeState.cell && this.lastEdgeState.edgeCode === edgeState.edgeCode) return;

        var curEdgeCode = edgeState.cell.getAttribute("data-edge-code");
        var curXEdgeCode = edgeState.cell.getAttribute("data-x-edge-code");
        var curEdgeVal = (curEdgeCode & edgeState.edgeCode) ? 1 : ((curXEdgeCode & edgeState.edgeCode) ? -1 : 0);

        if (!this.lastEdgeState) {
            this.fromEdgeVal = curEdgeVal;
            switch(mode) {
                case "cycle-front": this.toEdgeVal = this.fromEdgeVal + 1; break;
                case "cycle-back": this.toEdgeVal = this.fromEdgeVal - 1; break;
                case "toggle-line": this.toEdgeVal = (this.fromEdgeVal == 1) ? 0 : 1; break;
                case "toggle-x": this.toEdgeVal = (this.fromEdgeVal == -1) ? 0 : -1; break;
            }
            if (this.toEdgeVal > 1) { this.toEdgeVal -= 3; }
            if (this.toEdgeVal < -1) { this.toEdgeVal += 3; }
        }

        this.lastEdgeState = edgeState;

        if (curEdgeVal != this.fromEdgeVal) return;

        this.undoManager.startGroup(this);
        if (this.fromEdgeVal == 1 || this.toEdgeVal == 1) this.undoManager.modifyAttribute(edgeState.cell, "data-edge-code", curEdgeCode ^ edgeState.edgeCode);
        if (this.fromEdgeVal == -1 || this.toEdgeVal == -1) this.undoManager.modifyAttribute(edgeState.cell, "data-x-edge-code", curXEdgeCode ^ edgeState.edgeCode);
        this.undoManager.endGroup();
        this.updateSvg(edgeState.cell);
    }

    this.pointerDown = function(e) {
        if (this.pointerIsDown) { this.pointerIsDown = false; return; }

        this.pointerIsDown = true;

        if (e.target.hasPointerCapture(e.pointerId)) { e.target.releasePointerCapture(e.pointerId); }

        if ((document.activeElement == e.currentTarget) && this.options["data-text-advance-on-type"] && this.numCols > 1 && this.numRows > 1) {
            this.dx = 1 - this.dx; this.dy = 1 - this.dy;
            e.currentTarget.blur(); e.currentTarget.focus(); // Re-render the highlighting direction.
        }
        this.lastCell = e.currentTarget;
        this.currentFill = null;

        if (e.ctrlKey) {
            e.currentTarget.classList.toggle("interesting");
            e.preventDefault();
            return;
        }

        if (this.canDrawOnEdges) {
            var edgeState = this.getEventEdgeState(e);
            this.lastEdgeState = null;
            this.setEdgeState(edgeState, (e.button > 0 || e.shiftKey) ? "cycle-back" : "cycle-front");
            if (edgeState.any) {
                e.preventDefault();
                return;
            }
        }
        
        if (this.options["data-fill-cycle"] && !e.currentTarget.classList.contains("given-fill")) { this.currentFill = this.cycleClasses(e.currentTarget, this.fillClasses, e.button > 0 || e.shiftKey); }
        else { this.currentFill = this.findClassInList(e.currentTarget, this.fillClasses); }
        
        if (this.canHaveCenterFocus) {
            this.setCenterFocusMode();
            this.updateCenterFocus(e.currentTarget);
        }
        e.preventDefault();
    }

    this.pointerMove = function(e) {
        if (!this.pointerIsDown) return;

        e.preventDefault();

        if (this.canDrawOnEdges && !this.currentFill) {
            var edgeState = this.getEventEdgeState(e);
            this.setEdgeState(edgeState, (e.button > 0 || e.shiftKey) ? "cycle-back" : "cycle-front");
        }
    }

    this.pointerCancel = function(e) {
        this.pointerIsDown = false;
    }

    this.dragPaintAndPath = function(to) {
        var wantPaint = this.options["data-drag-paint-fill"] && !!this.currentFill;
        var canPaint = wantPaint;

        this.undoManager.startGroup(this);

        if (this.options["data-drag-draw-path"]) {
            var targetFill = this.findClassInList(to, this.fillClasses);
            var setLast = false; 
            if (wantPaint && this.currentFill == this.fillClasses[0]) { this.currentFill = targetFill; setLast = true; }
            if (wantPaint && targetFill != this.fillClasses[0] && targetFill != this.currentFill) { canPaint = false; }
            else { canPaint &= this.LinkCells(this.lastCell, to); }
        }

        if (canPaint && !to.classList.contains("given-fill")) {
            this.setClassInCycle(to, this.fillClasses, this.currentFill);
        }

        if (canPaint && setLast) { this.setClassInCycle(this.lastCell, this.fillClasses, this.currentFill); }

        var didWork = this.undoManager.endGroup();

        if (!wantPaint || canPaint) {
            this.lastCell = to;
            if (didWork) { this.updateCenterFocus(to); }
        }
    }

    this.pointerEnter = function(e) {
        if (!this.pointerIsDown) return;
        if (this.lastCell === e.currentTarget) return;

        this.dragPaintAndPath(e.currentTarget);
    }

    this.getOptionArray = function(option, splitchar, special) {
        var val = this.options[option];
        if (!val || Array.isArray(val) || val == special) { return val; }
        return val.split(splitchar);
    }

    this.getOptionDict = function(option) {
        var val = this.options[option];
        // TODO attribute version
        return val;
    }

    this.translate = function(ch, replacements) {
        if (!replacements || !replacements[ch]) return ch;
        return replacements[ch];
    }

    this.cluePointerEnter = function(e) {
        var acrosscluenumber = e.currentTarget.getAttribute("data-across-cluenumber");
        var downcluenumber = e.currentTarget.getAttribute("data-down-cluenumber");
        if (acrosscluenumber) { this.table.querySelectorAll("td[data-across-cluenumber='" + acrosscluenumber + "']").forEach(td => { td.classList.add("hovered"); }); }
        if (downcluenumber) { this.table.querySelectorAll("td[data-down-cluenumber='" + downcluenumber + "']").forEach(td => { td.classList.add("hovered"); }); }
    }

    this.cluePointerLeave = function(e) {
        var acrosscluenumber = e.currentTarget.getAttribute("data-across-cluenumber");
        var downcluenumber = e.currentTarget.getAttribute("data-down-cluenumber");
        if (acrosscluenumber) { this.table.querySelectorAll("td[data-across-cluenumber='" + acrosscluenumber + "']").forEach(td => { td.classList.remove("hovered"); }); }
        if (downcluenumber) { this.table.querySelectorAll("td[data-down-cluenumber='" + downcluenumber + "']").forEach(td => { td.classList.remove("hovered"); }); }
    }

    this.clueClick = function(e) {
        var acrosscluenumber = e.currentTarget.getAttribute("data-across-cluenumber");
        var downcluenumber = e.currentTarget.getAttribute("data-down-cluenumber");
        if (acrosscluenumber) { this.dx = 1; this.dy = 0; this.table.querySelector("td[data-across-cluenumber='" + acrosscluenumber + "']").focus(); }
        if (downcluenumber) { this.dx = 0; this.dy = 1; this.table.querySelector("td[data-down-cluenumber='" + downcluenumber + "']").focus(); }
    }

    this.scrollClue = function(li) {
        const ol = li.parentElement;
        if (li.offsetTop < ol.scrollTop ||
                li.offsetTop + li.offsetHeight > ol.scrollTop + ol.clientHeight) {
            ol.scrollTop = li.offsetTop + (li.offsetHeight - ol.clientHeight) / 2 - ol.offsetTop;
      }
    }

    this.mark = function(cell) {
        if (this.options["data-clue-locations"] !== "crossword") return;
        
        // Strip highlighting on all cells.
        this.table.querySelectorAll("td[data-across-cluenumber]").forEach(td => { td.classList.remove("marked"); });
        this.table.querySelectorAll("td[data-down-cluenumber]").forEach(td => { td.classList.remove("marked"); });
        // Now reapply the highlighting to relevant cells and clues.
        var acrosscluenumber = cell.getAttribute("data-across-cluenumber");
        var downcluenumber = cell.getAttribute("data-down-cluenumber");
        if (acrosscluenumber) {
            const li = this.container.querySelector("li[data-across-cluenumber='" + acrosscluenumber + "']");
            if (li) {
                li.classList.add("marked");
                this.scrollClue(li);
            }
            if (this.dx !== 0) {
                this.table.querySelectorAll("td[data-across-cluenumber='" + acrosscluenumber + "']").forEach(td => { td.classList.add("marked"); });
            }
        }
        if (downcluenumber) {
            const li = this.container.querySelector("li[data-down-cluenumber='" + downcluenumber + "']");
            if (li) {
                li.classList.add("marked");
                this.scrollClue(li);
            }
            if (this.dy !== 0) {
                this.table.querySelectorAll("td[data-down-cluenumber='" + downcluenumber + "']").forEach(td => { td.classList.add("marked"); });
            }
        }
    }

    this.unmark = function(cell) {
        if (this.options["data-clue-locations"] !== "crossword") return;

        var acrosscluenumber = cell.getAttribute("data-across-cluenumber");
        var downcluenumber = cell.getAttribute("data-down-cluenumber");
        if (acrosscluenumber) {
            this.container.querySelectorAll("li[data-across-cluenumber='" + acrosscluenumber + "']").forEach(li => { li.classList.remove("marked"); });
            if (this.dx !== 0) {
                this.table.querySelectorAll("td[data-across-cluenumber='" + acrosscluenumber + "']").forEach(td => { td.classList.remove("marked"); });
            }
        }
        if (downcluenumber) {
            this.container.querySelectorAll("li[data-down-cluenumber='" + downcluenumber + "']").forEach(li => { li.classList.remove("marked"); });
            if (this.dy !== 0) {
                this.table.querySelectorAll("td[data-down-cluenumber='" + downcluenumber + "']").forEach(td => { td.classList.remove("marked"); });
            }
        }
    }

    this.addEdgeToSvg = function(svg, edgeName) {
        var use = document.createElementNS("http://www.w3.org/2000/svg", "use");
        use.classList.add(edgeName);
        var edgePath = this.options["data-edge-style"];
        if (!edgePath.endsWith(".svg")) { edgePath = puzzleJsFolderPath + "edge-" + edgePath + ".svg"; }
        use.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", edgePath + "#" + edgeName);
        svg.appendChild(use);
    }

    this.pathTranslate = ["o0", "i2", "i0", "l0", "i1", "r2", "r1", "t1", "i3", "r3", "r0", "t3", "l1", "t2", "t0", "x0"];
    this.pathCopyjack = [" ", "╵", "╷", "│", "╴", "┘", "┐", "┤", "╶", "└", "┌", "├", "─", "┴", "┬", "┼"];
    this.updateSvg = function(td) {
        var svg = td.querySelector("svg");
        if (!svg) { svg = document.createElementNS("http://www.w3.org/2000/svg", "svg"); svg.setAttribute("viewBox", "-15 -15 30 30"); td.appendChild(svg); }

        var pathCode = td.getAttribute("data-path-code");
        if (pathCode) { pathCode = parseInt(pathCode); } else { pathCode = 0; }
        var translatedData = this.pathTranslate[pathCode];

        svg.innerHTML = "";

        if (pathCode) {
            var use = document.createElementNS("http://www.w3.org/2000/svg", "use");
            use.classList.add("path");
            var pathPath = this.options["data-path-style"];
            if (!pathPath.endsWith(".svg")) { pathPath = puzzleJsFolderPath + "path-" + pathPath + ".svg"; }
            use.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", pathPath + "#path-" + translatedData[0]);
            if (translatedData[1] != "0") { use.setAttributeNS(null, "transform", "rotate(" + parseInt(translatedData[1] * 90) + ")"); }
            svg.appendChild(use);
        }
        
        if (!td.classList.contains("unselectable")) { this.addEdgeToSvg(svg, "edge-base"); }

        var edgeCode = td.getAttribute("data-edge-code");
        if (edgeCode & 1) { this.addEdgeToSvg(svg, "edge-top"); }
        if (edgeCode & 2) { this.addEdgeToSvg(svg, "edge-bottom"); }
        if (edgeCode & 4) { this.addEdgeToSvg(svg, "edge-left"); }
        if (edgeCode & 8) { this.addEdgeToSvg(svg, "edge-right"); }

        edgeCode = td.getAttribute("data-x-edge-code");
        if (edgeCode & 1) { this.addEdgeToSvg(svg, "x-edge-top"); }
        if (edgeCode & 2) { this.addEdgeToSvg(svg, "x-edge-bottom"); }
        if (edgeCode & 4) { this.addEdgeToSvg(svg, "x-edge-left"); }
        if (edgeCode & 8) { this.addEdgeToSvg(svg, "x-edge-right"); }
    }

    this.IsFullyLinked = function(code) {
        var linkCount = 0;
        while (code) { linkCount++; code &= (code - 1); }
        return (linkCount >= 2);
    }

    this.LinkCellsDirectional = function(cellFrom, directionFrom, cellTo, directionTo) {
        var codeFrom = cellFrom.getAttribute("data-path-code");
        var codeTo = cellTo.getAttribute("data-path-code");
        if (!codeFrom) { codeFrom = 0; }
        if (!codeTo) { codeTo = 0; }

        if (!(codeFrom & directionFrom) && !(codeTo & directionTo) && !this.IsFullyLinked(codeFrom) && !this.IsFullyLinked(codeTo)) {
            this.undoManager.modifyAttribute(cellFrom, "data-path-code", codeFrom | directionFrom);
            this.undoManager.modifyAttribute(cellTo, "data-path-code", codeTo | directionTo);
            return true;
        }
        else if ((codeFrom & directionFrom) && (codeTo & directionTo)) {
            var givenCodeFrom = cellFrom.getAttribute("data-given-path-code");
            var givenCodeTo = cellTo.getAttribute("data-given-path-code");
            if (!(givenCodeFrom & directionFrom) && !(givenCodeTo & directionTo)) {
                this.undoManager.modifyAttribute(cellFrom, "data-path-code", codeFrom & ~directionFrom);
                this.undoManager.modifyAttribute(cellTo, "data-path-code", codeTo & ~directionTo);

                if (this.options["data-drag-paint-fill"]) {
                    if (cellFrom.getAttribute("data-path-code") == 0 && !cellFrom.classList.contains("given-fill")) {
                        this.setClassInCycle(cellFrom, this.fillClasses, this.fillClasses[0]);
                    }
                    if (cellTo.getAttribute("data-path-code") == 0 && !cellTo.classList.contains("given-fill")) {
                        this.setClassInCycle(cellTo, this.fillClasses, this.fillClasses[0]);
                    }
                }
                return true;
            }
        }

        return false;
    }

    this.LinkCells = function(cellFrom, cellTo)
    {
        var colFrom = Array.prototype.indexOf.call(cellFrom.parentElement.children, cellFrom) - this.leftClueDepth;
        var rowFrom = Array.prototype.indexOf.call(cellFrom.parentElement.parentElement.children, cellFrom.parentElement) - this.topClueDepth;
        var colTo = Array.prototype.indexOf.call(cellTo.parentElement.children, cellTo) - this.leftClueDepth;
        var rowTo = Array.prototype.indexOf.call(cellTo.parentElement.parentElement.children, cellTo.parentElement) - this.topClueDepth;

        if (colFrom === colTo) {
            if (rowFrom === rowTo - 1) { return this.LinkCellsDirectional(cellFrom, 2, cellTo, 1); }
            else if (rowFrom === rowTo + 1) { return this.LinkCellsDirectional(cellFrom, 1, cellTo, 2); }
        }
        else if (rowFrom === rowTo) {
            if (colFrom === colTo - 1) { return this.LinkCellsDirectional(cellFrom, 8, cellTo, 4); }
            else if (colFrom === colTo + 1) { return this.LinkCellsDirectional(cellFrom, 4, cellTo, 8); }
        }

        return false;
    }

    this.parseOuterClues = function(clues) {
        var clueDepth = 0;
        if (clues) { for (var i = 0; i < clues.length; i++) { clues[i] = clues[i].split(" "); clueDepth = Math.max(clues[i].length, clueDepth); } }
        return clueDepth;    
    }

    this.addEmptyOuterCell = function(tr) {
        var td = document.createElement("td");
        td.classList.add("cell");
        td.classList.add("outer-cell");
        td.classList.add("unselectable");
        tr.appendChild(td);
    }

    this.addOuterClue = function(tr, clues, clueIndex, cls) {
        var td = document.createElement("td");
        td.classList.add("cell");
        td.classList.add("outer-cell");
        if (clueIndex >= 0 && clueIndex < clues.length && clues[clueIndex]) {
            td.textContent = clues[clueIndex];
            td.classList.add(cls);
            td.addEventListener("pointerdown", e => { if (e.ctrlKey) { e.target.classList.toggle("interesting"); e.preventDefault(); } else if (e.shiftKey) { e.target.classList.toggle("strikethrough"); e.preventDefault(); } });
            td.addEventListener("contextmenu", e => { e.target.classList.toggle("strikethrough"); e.preventDefault(); });
        } else { td.classList.add("unselectable"); }

        tr.appendChild(td);
    }

    // Copyjack support.
    //
    // Reads from inputTd, a td that's part of an interactive puzzle element.
    // Expects that inputTd.dataset.copyjack is copyTd, a td that's part of this.copyjackVersion.
    // Styles copyTd such that clipboard copies it correctly.
    this.processTdForCopyjack = function(inputTd) {
        if (!this.isUsingCopyjack) return;
        const [i, j] = inputTd.dataset.coord.split(",")
        const copyTd = this.copyjackVersion.getElementsByTagName('tr')[i].getElementsByTagName('td')[j];
        // Set class names.
        // Remove "transient" class names such as 'marked'.
        copyTd.className = inputTd.className.replace(/marked/g, '');
        // Reset the font size to avoid row overflow.
        copyTd.style.fontSize = '1em';
        // Copy any text inside the td. This includes text inside divs within the td.
        copyTd.innerText = inputTd.innerText;
        // If the td has a "value", overwrite the innertext.
        const text = inputTd.querySelector('.text span');
        if (text && text.innerText) {
            copyTd.innerText = text.innerText;
        }

        // Do edges.
        const edgeCode = inputTd.dataset.edgeCode;
        copyTd.style.borderTop = (edgeCode & 1) ? '3px solid black' : '';
        copyTd.style.borderBottom = (edgeCode & 2) ? '3px solid black' : '';
        copyTd.style.borderLeft = (edgeCode & 4) ? '3px solid black' : '';
        copyTd.style.borderRight = (edgeCode & 8) ? '3px solid black' : '';

        // Do paths.
        const pathCode = inputTd.dataset.pathCode;
        if (!copyTd.innerText && pathCode) {
            // If the cell has content, don't overwrite it.
            copyTd.innerText = this.pathCopyjack[pathCode];
        }
    }

    this.prepareToReset = function() {
        localStorage.removeItem(this.stateKey);

        this.table.querySelectorAll(".inner-cell.extract .text span").forEach(s => {
            var extractId = s.getAttribute("data-extract-id");

            if (extractId) {
                document.querySelectorAll("." + extractId).forEach(elem => { elem.innerText = ""; });
            }
        });

        this.inhibitSave = true;
    }

    this.saveState = function() {
        if (this.inhibitSave) return;

        var stateArray = [];
        var hasState = false;

        this.table.querySelectorAll(".inner-cell").forEach(td => {
            var fillIndex = 0;
            if (this.fillClasses && !td.classList.contains("given-fill")) { fillIndex = this.fillClasses.indexOf(this.findClassInList(td, this.fillClasses)); }

            var edgeCode = td.getAttribute("data-edge-code");
            var givenEdgeCode = td.getAttribute("data-given-edge-code");
            if (!edgeCode) edgeCode = 0;
            if (!givenEdgeCode) givenEdgeCode = 0;
            var edgeCodeDelta = edgeCode ^ givenEdgeCode;

            var pathCode = td.getAttribute("data-path-code");
            var givenPathCode = td.getAttribute("data-given-path-code");
            if (!pathCode) pathCode = 0;
            if (!givenPathCode) givenPathCode = 0;
            var pathCodeDelta = pathCode ^ givenPathCode;

            var text = td.classList.contains("given-text") ? "" : td.querySelector(".text").innerText.trim();

            var cellState = "";
            if (fillIndex || edgeCodeDelta || pathCodeDelta || text) {
                hasState = true;
                cellState = fillIndex.toString(36) + edgeCodeDelta.toString(16) + pathCodeDelta.toString(16);
                if (text) { cellState += "," + text; }
            }

            stateArray.push(cellState);
        });

        if (hasState) { localStorage.setItem(this.stateKey, stateArray.join("|")); }
        else { localStorage.removeItem(this.stateKey); }
    }

    this.setCornerFocusMode = function(notyet) {
        if (this.firstCenterFocus) { this.firstCenterFocus.tabIndex = -1; }

        if (this.cornerFocus == null) {
            this.cornerFocus = document.createElement("div");
            this.cornerFocus.classList.add("corner-focus");
            this.cornerFocus.addEventListener("keydown",  e => { this.keyDownCorner(e); });
            this.table.appendChild(this.cornerFocus);

            this.cornerFocusX = 0;
            this.cornerFocusY = 0;
        }

        this.cornerFocus.tabIndex = 0;
        if (!notyet) { this.cornerFocus.focus(); }
        this.updateCornerFocus();
    }

    this.setCenterFocusMode = function(notyet) {
        if (this.cornerFocus) { this.cornerFocus.tabIndex = -1; }
        this.firstCenterFocus.tabIndex = 0;
        if (!notyet && this.currentCenterFocus) { this.currentCenterFocus.focus(); }
    }

    this.updateCornerFocus = function() {
        var topLeftTD = this.table.children[this.topClueDepth].children[this.leftClueDepth];
        this.cornerFocus.style.left = (topLeftTD.offsetLeft + this.cornerFocusX * topLeftTD.offsetWidth) + "px";
        this.cornerFocus.style.top = (topLeftTD.offsetTop + this.cornerFocusY * topLeftTD.offsetHeight) + "px";
    }

    this.updateCenterFocus = function(center) {
        this.currentCenterFocus = center;
        this.currentCenterFocus.focus();
    }

    this.closeAbout = function() {
        this.container.querySelector(".puzzle-about-back").remove();
        this.container.querySelector(".puzzle-about").remove();

        if (this.oldFocus) { this.oldFocus.focus(); this.oldFocus = null; }
    }

    this.aboutPopup = function() {
        var lines = [];
        lines.push("<div class='puzzle-about-back no-copy'></div>");
        lines.push("<div class='puzzle-about no-copy'>");
        lines.push("<div class='puzzle-about-scroller'>");
        lines.push("<div class='puzzle-about-savedstate'>This puzzle will save its state when you leave/refresh the page.</div>");
        lines.push("<table>");
        lines.push("<tr><th>Function</th><th>Keyboard</th><th>Mouse/Touch</th></tr>");
        lines.push("<tr><td>Reset saved state</td><td>N/A</td><td>Reset Button</td></tr>");
        lines.push("<tr><td>Undo/Redo</td><td>Ctrl+Z/Y</td><td>Undo/Redo Buttons</td></tr>");
        if (this.options["data-text-characters"]) {
            if (this.options["data-text-shift-lock"]) {
                lines.push("<tr><td>Multiple-character text entry</td><td>Type to append; Backspace to remove, Del to clear</td><td>N/A</td></tr>");
            } else {
                lines.push("<tr><td>Single-character text entry</td><td>Type to replace; Backspace/Del to clear</td><td>N/A</td></tr>");
                if (this.options["data-text-shift-key"] == "rebus") {
                    lines.push("<tr><td>Rebus clue text entry</td><td>Shift-Type to append; Shift-Backspace to remove</td><td>N/A</td></tr>");
                } else {
                    lines.push("<tr><td>Candidate-value text entry</td><td>Shift-Type to toggle a character</td><td>N/A</td></tr>");
                }
            }
        }
        if (this.canHaveCenterFocus) {
            lines.push("<tr><td>Navigate between cells</td><td>Arrow keys</td><td>Click a cell</td></tr>");
        }
        if (this.fillClasses && this.fillClasses.length > 0 && this.options["data-fill-cycle"]) {
            lines.push("<tr><td>Change cell background (forwards or backwards)</td><td>Space or Shift-Space</td><td>Click/Left-Click or Right/Shift-Click</td></tr>");
        }
        if (this.options["data-drag-draw-path"]) {
            lines.push("<tr><td>Draw a path between cells</td><td>Shift-arrow keys</td><td>Click one cell, drag to others</td></tr>");
        }
        if (this.options["data-drag-draw-edge"]) {
            if (this.canHaveCenterFocus) {
                lines.push("<tr><td>Draw an edge between cells</td><td>'.' to enter/exit corner mode, then Shift-arrow keys</td><td>Click one corner or edge, drag to others</td></tr>");
            } else {
                lines.push("<tr><td>Draw an edge between cells</td><td>Shift-arrow keys</td><td>Click one corner or edge, drag to others</td></tr>");
            }
        }
        lines.push("<tr><td>Mark a cell as 'interesting'</td><td>Ctrl+Space</td><td>Ctrl+Click</td></tr>");
        if (this.leftClueDepth || this.rightClueDepth || this.topClueDepth || this.bottomClueDepth) {
            lines.push("<tr><td>Mark an external clue as 'satisfied'</td><td>N/A</td><td>Right+Click</td></tr>");
        }
        lines.push("</table></div>");
        lines.push("<div class='puzzle-about-credits'>Made with <a href='https://github.com/tabascq/PuzzleJS' target='_blank'>Puzzle.js</a></div>")
        lines.push("<button type='button' class='puzzle-about-close'>Close</button>")
        lines.push("</div>");
        this.container.insertAdjacentHTML("beforeend", lines.join(""));

        this.oldFocus = document.activeElement;

        this.container.querySelector(".puzzle-about-back").addEventListener("pointerdown", e => { this.closeAbout(); });
        this.container.querySelector(".puzzle-about-close").focus();
        this.container.querySelector(".puzzle-about-close").addEventListener("click", e => { this.closeAbout(); });
        this.container.querySelector(".puzzle-about-close").addEventListener("keyup", e => { if (e.keyCode == 27 || e.keyCode == 13) this.closeAbout(); });
    }

    // --- Construct the interactive player. ---
    this.fillClasses = this.getOptionArray("data-fill-classes", " ");

    var clueIndicators = this.getOptionArray("data-clue-indicators", " ", "auto");
    var textLines = this.getOptionArray("data-text", "|");
    var textReplacements = this.getOptionDict("data-text-replacements");
    var fills = this.getOptionArray("data-fills", "|");
    var solution = this.getOptionArray("data-text-solution", "|");
    var edges = this.getOptionArray("data-edges", "|");
    var paths = this.getOptionArray("data-paths", "|");
    var extracts = this.getOptionArray("data-extracts", " ");
    var unselectableGivens = this.options["data-unselectable-givens"];
    var topClues = this.getOptionArray("data-top-clues", "|");
    var bottomClues = this.getOptionArray("data-bottom-clues", "|");
    var leftClues = this.getOptionArray("data-left-clues", "|");
    var rightClues = this.getOptionArray("data-right-clues", "|");

    this.topClueDepth = this.parseOuterClues(topClues);
    this.bottomClueDepth = this.parseOuterClues(bottomClues);
    this.leftClueDepth = this.parseOuterClues(leftClues);
    this.rightClueDepth = this.parseOuterClues(rightClues);

    if (!textLines) { textLines = solution; }

    var allowInput = !this.options["data-no-input"];
    var table = document.createElement("table");
    var clueNum = 0;
    var extractNum = 0;

    var acrossClues = this.container.querySelectorAll(".crossword-clues.across li");
    var acrossClueIndex = 0;
    var downClues = this.container.querySelectorAll(".crossword-clues.down li");
    var downClueIndex = 0;

    this.canDrawOnEdges = this.options["data-drag-draw-edge"] && !this.options["data-no-input"];

    this.canHaveCenterFocus = this.options["data-text-characters"] || (fills && this.options["data-fill-cycle"]) || this.options["data-drag-draw-path"];
    this.canHaveCornerFocus = this.canDrawOnEdges;
    this.keyboardFocusModel = this.options["data-no-input"] ? "none" : (this.canHaveCornerFocus ? "corner" : "center");
    this.cornerFocus = null;
    this.firstCenterFocus = null;

    var regularRowBorder = 0;
    var regularColBorder = 0;

    var savedState = localStorage.getItem(this.stateKey);
    if (savedState) { savedState = savedState.split("|"); }

    if (!allowInput) {
        this.container.classList.add("no-input");
    }

    if (textLines.length == 1 && /^\d+x\d+$/.test(textLines[0])) {
        var dim = textLines[0].split("x");
        textLines = [];
        for (r = 0; r < dim[1]; r++) {
            textLines[r] = [];
            for (c = 0; c < dim[0]; c++) { textLines[r][c] = "."; }
        }
    }

    if (edges && edges.length == 1 && /^\d+x\d+$/.test(edges[0])) {
        var dim = edges[0].split("x");
        edges = null;
        regularColBorder = dim[0];
        regularRowBorder = dim[1];
    }

    for (var i = 0; i < this.topClueDepth; i++) {
        var tr = document.createElement("tr");
        for (var j = 0; j < this.leftClueDepth; j++) { this.addEmptyOuterCell(tr); }
        for (var j = 0; j < topClues.length; j++) { this.addOuterClue(tr, topClues[j], i - this.topClueDepth + topClues[j].length, "top-clue"); }
        for (var j = 0; j < this.rightClueDepth; j++) { this.addEmptyOuterCell(tr); }

        table.appendChild(tr);
    }

    this.numRows = textLines.length;
    this.numCols = 0;

    var stateIndex = 0;

    for (var r = 0; r < textLines.length; r++) {
        var tr = document.createElement("tr");

        for (var j = 0; j < this.leftClueDepth; j++) { this.addOuterClue(tr, leftClues[r], j - this.leftClueDepth + leftClues[r].length, "left-clue"); }

        this.numCols = Math.max(this.numCols, textLines[r].length);
        for (var c = 0; c < textLines[r].length; c++) {
            var cellSavedState = null;
            if (savedState) { cellSavedState = savedState[stateIndex++]; }

            var td = document.createElement("td");
            td.classList.add("cell");
            td.classList.add("inner-cell");
            var ch = textLines[r][c];
            
            var textwrapper = document.createElement("div");
            textwrapper.classList.add("text");

            var text = document.createElement("span");
            textwrapper.appendChild(text);

            if (ch == '.') {
                if (solution) { text.innerText = this.translate(solution[r][c], textReplacements); }
                else if (allowInput && this.options["data-text-characters"]) { td.contentEditable = true; td.autocapitalize="off"; }
            }
            else if (ch == '#') {
                td.classList.add("extract");
                if (solution) { text.innerText = this.translate(solution[r][c], textReplacements); }
                else if (allowInput && this.options["data-text-characters"]) { td.contentEditable = true; td.autocapitalize="off"; }

                if (extracts) {
                    var code = extracts[extractNum++];
                    var id = "extract-id-" + code;
                    text.setAttribute("data-extract-id", id);
                    text.classList.add(id);

                    var extractCode = document.createElement("div");
                    extractCode.contentEditable = false;
                    extractCode.classList.add("extract-code");
                    extractCode.innerText = code;
                    td.appendChild(extractCode);    
                }
            }
            else if (ch == '@') {
                td.classList.add("black-cell");
                td.classList.add("unselectable");
            }
            else {
                text.innerText = this.translate(ch, textReplacements);
                td.classList.add("given-text");
                if (unselectableGivens) { td.classList.add("unselectable"); }
            }

            if (cellSavedState && cellSavedState.indexOf(",") >= 0) {
                var savedText = cellSavedState.substring(cellSavedState.indexOf(",") + 1).trim();
                text.innerText = savedText;
                if (savedText && savedText.length > 1) { td.classList.add("small-text"); }
            }

            if (!td.classList.contains("unselectable")) {
                if (allowInput) {
                    td.tabIndex = this.firstCenterFocus ? -1 : 0;
                    if (!this.firstCenterFocus) { this.firstCenterFocus = td; }
                    td.addEventListener("keydown",  e => { this.keyDown(e); });
                    td.addEventListener("beforeinput", e => { this.beforeInput(e); });
                    td.addEventListener("pointerdown",  e => { this.pointerDown(e); });
                    td.addEventListener("pointermove",  e => { this.pointerMove(e); });
                    td.addEventListener("pointerenter",  e => { this.pointerEnter(e); });
                    td.addEventListener("pointercancel",  e => { this.pointerCancel(e); });
                    td.addEventListener("contextmenu",  e => { e.preventDefault(); });
                    if (this.options["data-clue-locations"] === "crossword") {
                        td.addEventListener("focus",  e => { this.mark(e.target); });
                        td.addEventListener("blur",  e => { this.unmark(e.target); });
                    }
                }
            }

            td.appendChild(textwrapper);

            var edgeCode = 0;
            if (edges || regularRowBorder || regularColBorder) {
                if (regularRowBorder) {
                    if ((r % regularRowBorder) == 0) { edgeCode |= 1; }
                    if (r == textLines.length - 1) { edgeCode |= 2; }
                }
                if (regularColBorder) {
                    if ((c % regularColBorder) == 0) { edgeCode |= 4; }
                    if (c == textLines[r].length - 1) { edgeCode |= 8; }
                }
    
                if (edges) {
                    if (edges.length == textLines.length) {
                        edgeCode |= parseInt(edges[r][c], 16);
                    }
                    else if (edges.length == textLines.length * 2 + 1) {
                        var topRow = edges[r * 2];
                        var midRow = edges[r * 2 + 1];
                        var botRow = edges[r * 2 + 2];
                        var chTop = (topRow.length == textLines[r].length) ? topRow[c] : topRow[c * 2 + 1];
                        var chLeft = (midRow.length == textLines[r].length + 1) ? midRow[c] : midRow[c * 2];
                        var chRight = (midRow.length == textLines[r].length + 1) ? midRow[c + 1] : midRow[c * 2 + 2];
                        var chBottom = (botRow.length == textLines[r].length) ? botRow[c] : botRow[c * 2 + 1];
                        if (chTop != " " && chTop != ".") { edgeCode |= 1; }
                        if (chBottom != " " && chBottom != ".") { edgeCode |= 2; }
                        if (chLeft != " " && chLeft != ".") { edgeCode |= 4; }
                        if (chRight != " " && chRight != ".") { edgeCode |= 8; }
                    }
                }
    
                if (edgeCode) { td.setAttribute("data-given-edge-code", edgeCode); }
            }
            if (cellSavedState) { edgeCode ^= parseInt(cellSavedState[1], 16); }
            if (edgeCode) { td.setAttribute("data-edge-code", edgeCode); }

            var pathCode = 0;
            if (paths) {
                if (paths.length == textLines.length) {
                    pathCode |= parseInt(paths[r][c], 16);
                }
                else if (paths.length == textLines.length * 2 + 1) {
                    var topRow = paths[r * 2];
                    var midRow = paths[r * 2 + 1];
                    var botRow = paths[r * 2 + 2];
                    var chTop = (topRow.length == textLines[r].length) ? topRow[c] : topRow[c * 2 + 1];
                    var chLeft = (midRow.length == textLines[r].length + 1) ? midRow[c] : midRow[c * 2];
                    var chRight = (midRow.length == textLines[r].length + 1) ? midRow[c + 1] : midRow[c * 2 + 2];
                    var chBottom = (botRow.length == textLines[r].length) ? botRow[c] : botRow[c * 2 + 1];
                    if (chTop != " " && chTop != ".") { pathCode |= 1; }
                    if (chBottom != " " && chBottom != ".") { pathCode |= 2; }
                    if (chLeft != " " && chLeft != ".") { pathCode |= 4; }
                    if (chRight != " " && chRight != ".") { pathCode |= 8; }
                }

                if (pathCode) { td.setAttribute("data-path-code", pathCode); td.setAttribute("data-given-path-code", pathCode); }
            }
            if (cellSavedState) { pathCode ^= parseInt(cellSavedState[2], 16); }
            if (pathCode) { td.setAttribute("data-path-code", pathCode); }

            if (this.options["data-clue-locations"] && textLines[r][c] != '@') {
                var acrossClue = (c == 0 || textLines[r][c-1] == '@' || (edgeCode & 4)) && c < textLines[r].length - 1 && textLines[r][c+1] != '@' && !(edgeCode & 8); // block/edge left, letter right
                var downClue = (r == 0 || textLines[r-1][c] == '@' || (edgeCode & 1)) && r < textLines.length - 1 && textLines[r+1][c] != '@' && !(edgeCode & 2); // block/edge above, letter below
                
                if (acrossClue || downClue || this.options["data-clue-locations"] == "all") {
                    var clueIndicator = (!clueIndicators) ? ++clueNum : clueIndicators[clueNum++];
                    const isClueEmpty = String(clueIndicator).trim() === "";

                    if (!isClueEmpty && this.options["data-clue-locations"] == "crossword") {
                        if (acrossClue) { td.setAttribute("data-across-cluenumber", clueIndicator); }
                        if (downClue) { td.setAttribute("data-down-cluenumber", clueIndicator); }
                        if (acrossClue && acrossClues[acrossClueIndex]) {
                          acrossClues[acrossClueIndex].setAttribute("data-across-cluenumber", clueIndicator);
                          acrossClues[acrossClueIndex].setAttribute("value", clueIndicator);
                          acrossClueIndex++;
                        }
                        if (downClue && downClues[downClueIndex]) {
                          downClues[downClueIndex].setAttribute("data-down-cluenumber", clueIndicator);
                          downClues[downClueIndex].setAttribute("value", clueIndicator);
                          downClueIndex++;
                        }
                    }

                    var clue = document.createElement("div");
                    clue.contentEditable = false;
                    clue.classList.add("clue");
                    clue.innerText = clueIndicator;
                    td.appendChild(clue);
                }

                if (this.options["data-clue-locations"] == "crossword") {
                    if (!acrossClue && c > 0 && textLines[r][c-1] != '@' && !(edgeCode & 4)) { td.setAttribute("data-across-cluenumber", tr.children[c-1].getAttribute("data-across-cluenumber")); }
                    if (!downClue && r > 0 && textLines[r-1][c] != '@' && !(edgeCode & 1)) { td.setAttribute("data-down-cluenumber", table.children[r-1].children[c].getAttribute("data-down-cluenumber")); }
                }
            }

            if (this.fillClasses) {
                var fillIndex = 0;
                if (fills && fills[r][c] != '.') {
                    fillIndex = parseInt(fills[r][c], 36);
                    td.classList.add("given-fill");
                } else {
                    fillIndex = cellSavedState ? parseInt(cellSavedState[0], 36) : 0;
                }
                td.classList.add(this.fillClasses[fillIndex]);
            }

            this.updateSvg(td);
            tr.appendChild(td);
        }

        for (var j = 0; j < this.rightClueDepth; j++) { this.addOuterClue(tr, rightClues[r], j, "right-clue"); }

        table.appendChild(tr);
    }

    for (var i = 0; i < this.bottomClueDepth; i++) {
        var tr = document.createElement("tr");
        for (var j = 0; j < this.leftClueDepth; j++) { this.addEmptyOuterCell(tr); }
        for (var j = 0; j < bottomClues.length; j++) { this.addOuterClue(tr, bottomClues[j], i, "bottom-clue"); }
        for (var j = 0; j < this.rightClueDepth; j++) { this.addEmptyOuterCell(tr); }

        table.appendChild(tr);
    }

    this.table = table;

    this.container.insertBefore(table, this.container.firstChild);

    this.dx = 0;
    this.dy = 0;
    if (this.options["data-text-advance-on-type"]) {
        if (this.numCols > 1) { this.dx = 1; }        
        else if (this.numRows > 1) { this.dy = 1; }        
    }

    if (this.options["data-show-commands"]) {
        this.commands = document.createElement("div");
        this.commands.classList.add("puzzle-commands");
        this.commands.classList.add("no-copy");
        this.commands.innerHTML = "<button type='button' class='puzzle-about-button'>About</button><button type='button' class='puzzle-undo-button'>Undo</button><button type='button' class='puzzle-redo-button'>Redo</button><button type='button' class='puzzle-reset-button'>Reset</button>";
        this.commands.querySelector(".puzzle-about-button").addEventListener("click", e => { this.aboutPopup(); });
        this.commands.querySelector(".puzzle-undo-button").addEventListener("click", e => { this.undoManager.undo(); });
        this.commands.querySelector(".puzzle-redo-button").addEventListener("click", e => { this.undoManager.redo(); });
        // TODO shouldn't need a reload
        this.commands.querySelector(".puzzle-reset-button").addEventListener("click", e => { this.prepareToReset(); window.location.reload(); });

        // add as an extra row of the table because everything else breaks layout somewhere
        var row = table.insertRow(-1);
        row.classList.add("commands");
        for (var i = 0; i < this.leftClueDepth; i++) { row.insertCell(-1); }
        var cell = row.insertCell(-1);
        cell.style.maxWidth = table.offsetWidth;
        cell.colSpan = this.numCols;
        cell.insertBefore(this.commands, null);
        for (var i = 0; i < this.rightClueDepth; i++) { row.insertCell(-1); }
    }

    if (this.keyboardFocusModel == "corner") {
        this.setCornerFocusMode(true);        
    }

    // Copyjack support: initialize a copyjack version of the table.
    // This table will be modified as the user takes actions.

    // Put the table inside this.container to ensure that styling works.
    if (this.isUsingCopyjack) {
        // Set no-copy on this table.
        table.classList.add('no-copy');
        // Create a copy-only table and insert it.
        this.copyjackVersion = document.createElement('table');
        this.copyjackVersion.classList.add('copy-only');
        this.copyjackVersion.style.userSelect = 'auto'; // Needed for Firefox compatibility.
        this.container.insertBefore(this.copyjackVersion, this.table);
        // Populate the copy-only table.
        for (const [i, tr] of Array.from(table.getElementsByTagName('tr')).entries()) {
            const copyTr = document.createElement('tr');
            copyTr.style.userSelect = 'auto';
            this.copyjackVersion.appendChild(copyTr);
            for (const [j, td] of Array.from(tr.getElementsByTagName('td')).entries()) {
                td.dataset.coord = [i, j];
                const copyTd = document.createElement('td');
                copyTd.style.userSelect = 'auto';
                copyTr.appendChild(copyTd);
                this.processTdForCopyjack(td);
            }
        }
    }

    if (allowInput) {
        this.container.querySelectorAll(".crossword-clues li").forEach((clue) => {
            clue.addEventListener("pointerenter", e => { this.cluePointerEnter(e); });
            clue.addEventListener("pointerleave", e => { this.cluePointerLeave(e); });
            clue.addEventListener("click", e => { this.clueClick(e); });
            clue.addEventListener("contextmenu", e => { e.target.classList.toggle("strikethrough"); e.preventDefault(); });
        });

        window.addEventListener("pointerup", e => {this.pointerIsDown = false; });

        document.addEventListener("keyup", function(e) { this.fShift = e.shiftKey; });
        document.addEventListener("keydown", function(e) { this.fShift = e.shiftKey; });

        window.addEventListener("beforeunload", e => { this.saveState(); });
    }
}
