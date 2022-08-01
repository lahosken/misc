#!/usr/bin/env python3

REJECT_LETTERS = "qzxjy"

DENY = """aaa
aas
alla
arie
ala
ata
atl
atm
att
ans
asa
aron
beal
bebes
bels
beame
carls
carte
chem
cher
chet
cocos
cols
comea
comin
comte
cosa
eee
eeg
een
ein
enco
ene
ento
enarm
ese
eine
hmm
iii
laa
las
lat
laine
lapat
lats
lst
marts
mell
mels
mets
mma
mmm
nae
oneal
ooo
pepes
riche
ritt
tse
ste
sse
sst
stl
sint
tal
tas
tnn
sta
stn
sts
stro
rons
romes
theres
wwe
www""".splitlines()

for line in open("/home/lahosken/words_500K.txt"):
    score_s, word = line.strip().split("\t")
    score_i = int(score_s, 10)
    if score_i < 4000: break
    if len(word) < 3: continue
    if len(word) > 9: continue
    if word in DENY: continue
    if len([c for c in REJECT_LETTERS if c in word]): continue
    print(line.strip())
print("3899\trein")
print("3740\tteal")
print("3727\ttaro")
print("3718\tchar")
print("3692\teon")
print("3683\tparse")
print("3683\trend")
print("3570\ttet")
print("3557\ttine")
print("3546\ttonal")
print("3520\tsinge")
print("3484\tstoat")
print("3484\trasta")
print("2917\tell")
print("2737\tess")
print("2567\therald")
print("2293\talee")
print("1969\therein")
print("1411\tstrongarm")
print("1377\trotund")
print("783\tstatin")
print("740\tpinecones")
