const GRAMS = 'e t a in ch her'.split(' ').sort();

const WORDS_L = 'ache aint ate chain chat china each eat etch ether herein tat tea tech tee tet there tina tine tint'.split(' ').sort();

var WORDS = {};
for (var i in WORDS_L) {
    WORDS[WORDS_L[i]] = true
}
