function doStuff() {
    if (wire > 0) { clipClick(1) }
    if (swarmGifts > 0) {
	if (memory >= 300 || (memory * Math.random() * 0.7) > processors) {
	    addProc();
	} else {
	    addMem();
	}
	return
    }

    if (Math.random() > 0.1) { return }

    if (swarmFlag && Math.random() < 0.1) {
	var l = sliderElement.value;
	if (availableMatter <= 0 && acquiredMatter <= 0) {
	    l += 5;
	}
	if (processors >= 200) {
	    l -= memory / 200;
	}
	if (memory >= 300) {
	    l -= processors / 200;
	}
	if (wire <= 0) { l--; }
	if (wire * 100 > unusedClips) { l++; }
	if (wire * 10 > unusedClips) { l++; }
	if (wire > unusedClips) { l += 2; }
	if (acquiredMatter <= 0) { l--; }
	if (acquiredMatter >= availableMatter) { l++; }
	if (acquiredMatter >= 10 * availableMatter) { l++; }
	if (acquiredMatter >= 100 * availableMatter) { l++; }
	if (processors * memory < 300 * 300) {
	    l += Math.random() * 1.2;
	}
	if (Math.random() < 0.01) {
	    l += (Math.random() * 20) - 10
	}
	if (l < 0) { l = 0; }
	if (l > 200) { l = 200; }
	if (l != sliderElement.value) {
	    sliderElement.value = Math.floor(l);
	}
    }
    
    if (strategyEngineFlag
	&& yomi <= creativity
	&& operations >= 1000 * memory
	&& (strats.length >= 4
	    || yomi * Math.random() < 100 * strats.length)
	&& !btnNewTournamentElement.disabled
	&& Math.random() < 0.1) {
	    
	newTourney();
	if (stratPickerElement.value === "10") {
	    stratPickerElement.value = "" + strats.length - 1
	}
	runTourney();
	return
    }
    
    var projs = projectsDivElement.getElementsByClassName('projectButton');
    
    if (humanFlag) {
	if (wire * wireCost < wireSupply * 20) { buyWire() }
	if (trust > memory + processors) {
	    if (processors < 4 || memory >= 80) {
		addProc();
	    } else if (memory <= 2*(processors+1) || memory < trust-30) {
		addMem();
	    } else if (projectsDivElement.innerHTML.includes('50,000 ops')
		       && memory < 50) {
		addMem();
	    } else if (memory >= 3*(processors+1)) {
		addProc();
	    } else if (Math.random() * memory <= 1.3*processors) {
		addMem();
	    } else {
		addProc();
	    }
	    return
	}
	var goal = 0;
	var t = projectsDivElement.innerHTML;
	if (t.includes('Takeov')) {
	    goal = 1000000;
	} else if (t.includes('Monopo') && yomi >= 3000) {
	    goal = 10000000;
	} else if (t.includes('Goodwill')) {
	    var goal_m = t.match(/\(\$(.*)\)/); /* 1,000 from ($1,000) */
	    if (goal_m && goal_m.length == 2) {
		goal = parseInt(goal = goal_m[1].replace(/,/g, ''), 10);
	    }
	}

	var thrift = 1 + Math.log(goal + 1);
	if (goal > 0 && funds + portTotal > goal) { thrift += 1000000; }
	if (adCost < funds / thrift) {
	    buyAds()
	    return
	}
	if (Math.random() < margin) {
	    var madMoney = (funds / thrift) - (3*wireCost);
	    if (megaClipperFlag) {
		if (megaClipperCost < madMoney
		    && Math.random() < Math.pow(margin, 2)) {
		    makeMegaClipper();
		    return
		}
		if (clipperCost * 500 < megaClipperCost
		    && clipperCost < madMoney) {
		    makeClipper();
		    return
		}
	    } else { /* megaClippers not available */
		if (clipmakerLevel > marketingLvl * 10) {
		    madMoney /= 2;
		}
		if (clipmakerLevel > memory * 75 / 12) {
		    madMoney /= 2;
		}
		if (clipmakerLevel >= 75) {
		    madMoney = 0;
		}
		if (clipperCost < madMoney && margin > 0.02) {
		    makeClipper()
		    return
		}
	    }
	}

	var manufRate_s = clipmakerRateElement.innerText.replace(/,/g , '');
	var manufRate = parseInt(manufRate_s, 10);

	if (Math.random() < 0.1) {
	    if (margin > 0.01
		&& unsoldClips > 3 * manufRate
		&& margin * unsoldClips >= wireCost * 1.5
	        && Math.random() < 0.33) {
		lowerPrice();
		return
	    }
	    if (unsoldClips < manufRate || margin * unsoldClips <= wireCost) {
		raisePrice();
		return
	    }
	}
	if (investmentEngineFlag && Math.random() < 0.1) {
	    function shouldUpgrade() {
		return investUpgradeCost * (2*investLevel + 1) <= yomi;
	    }
	    if (goal > 0 ) {
		if (funds + portTotal > goal * 1.2 && bankroll > 0) {
		    investWithdraw();
		    return
		}
		if (funds + portTotal < goal && funds > goal / 20) {
		    investDeposit();
		    if (shouldUpgrade()) { investUpgrade(); }
		    return
		}
	    } else {
		if (bankroll > funds
		    && bankroll > megaClipperCost
		    && bankroll * 3 < portTotal) {
		    investWithdraw();
		    return
		}
		if (funds > bankroll
		    && (funds < megaClipperCost * thrift * 0.8
			|| portTotal <= 0)
		    && funds * 3 > portTotal
		    && (investLevel > 1 || shouldUpgrade())) {
		    investDeposit();
		    if (shouldUpgrade()) { investUpgrade(); }
		    return
		}
		if (portTotal <= 0 && funds > 1200000) {
		    investDeposit();
		    if (shouldUpgrade()) { investUpgrade(); }
		    return
		}
	    }
	}


	for (var ix = 0; ix < projs.length; ix++) {
	    if (projs[ix].disabled) { continue }
	    var t = projs[ix].innerHTML;

	    if (t.includes('Creati')
		|| t.includes('spool')
		|| t.includes('Takeov')
		|| t.includes('Monop')
		|| t.includes('Goodwill')) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Clippers')
		&& Math.random() < margin * (margin - 0.03)) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Hadwiger') && wireSupply >= 4000) {
		projs[ix].click()
		return
	    }
	    if ((t.includes('Harm')
		 || t.includes('Lex')
		 || t.includes('arket'))
		&& Math.random() * 10000000 * margin < Math.pow(marketingLvl, 5)) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Trading')
		&& operations >= memory * 1000
		&& (creativity >= 250 || strategyEngineFlag)
		&& Math.random() < 0.001) {
		projs[ix].click();
		investStratElement.value = "med";
		return
	    }
	    if ((t.includes('Strateg') || t.includes('Donkey'))
		 && operations >= memory * 1000
		 && memory >= 11
		 && Math.random() < 0.001) {
		projs[ix].click()
		if (strats.length % 4 == 0) {
		    stratPickerElement.value = "" + strats.length - 1
		}
		return
	    }
	    if (t.includes('Volition')
		&& operations >= memory * 1000
		&& Math.random() < 0.001) {
		projs[ix].click()
		return
	    }
	    if ((t.includes('Peace')
		 || t.includes('Bald')
		 || t.includes('Warming'))
		&& operations >= memory * 1000
		&& Math.random() < 0.001) {
		projs[ix].click();
		return
	    }
            if (t.includes('Cancer')
		&& (goal <= 0 || goal > (funds + portTotal + 10000000) * 2)
		&& processors * memory * yomi > 30 * 70 * 36000) {
		projs[ix].click();
		return
	    }
	    if (t.includes('Limerick')
		&& Math.random() < 0.01) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Sausage')
		&& (creativity >= 450 || strategyEngineFlag)
		&& Math.random() < 0.01) {
		projs[ix].click()
		return
	    }
	    // 36000 yomi: enough for Swarm computing in next phase
	    if (t.includes('HypnoDrones')
		&& operations >= memory * 1000
		&& memory * yomi >= 80 * 36000
		&& clips / nextTrust < 0.95
		&& (goal <= 0 || goal > (funds + portTotal + 10000000) * 2)
		&& Math.random() < 0.1) {
		if (t.includes('Release the HypnoDrones')) {
		    if (trust > memory + processors) {
			addProc();
			return
		    }
		    for (var ix2 = 0; ix2 < projs.length; ix2++) {
			if (ix2 == ix) { continue }
			if (projs[ix2].disabled) { continue }
			if (projs[ix2].innerHTML.includes('Trust')) {
			    projs[ix2].click();
			    return
			}
		    }
		}
		if (Math.random() < 0.01) { 
		    projs[ix].click()
		    return
		}
	    }
	}
    } else if (!spaceFlag) {
	var supply = farmLevel * farmRate/100;
	var dDemand = (harvesterLevel * dronePowerRate/100) + (wireDroneLevel * dronePowerRate/100);
	var fDemand = (factoryLevel * factoryPowerRate/100);
	var demand = dDemand + fDemand;
	if (supply < demand * 1.15 && unusedClips > farmCost) {
	    makeFarm(1);
	    return
	}
	if (batteryLevel < 1000
	    && (availableMatter < 5.8 * Math.pow(10, 27)
		|| batteryLevel < factoryLevel * factoryLevel / 10
		|| batteryLevel < factoryLevel)
	    && unusedClips > batteryCost
	    && supply > demand
	    && storedPower >= batteryLevel * 10000 * 0.9) {
	    makeBattery(1);
	    return
	}
	if (swarmFlag) {
	    if (availableMatter <= 0 && acquiredMatter > wire) {
		sliderElement.value = Math.floor(sliderElement.value * 0.99);
	    }
	}

	function makeSomeHarvesters() {
	    makeHarvester(Math.floor(Math.sqrt(wireDroneLevel)) + 1);
	}
	function makeSomeWireDrones() {
	    makeWireDrone(Math.floor(Math.sqrt(harvesterLevel)) + 1);
	}
	
	if (harvesterFlag
	    && factoryFlag
	    && unusedClips > harvesterCost
	    && supply >= demand
	    && unusedClips > factoryCost * 2
	    && harvesterLevel < 25000
	    && availableMatter > Math.pow(10, 27)
	    && Math.random() < 0.2) {
	    if (harvesterLevel * 1.5 <= wireDroneLevel) {
		makeSomeHarvesters();
		return
	    }
	    if (acquiredMatter <= 0 && availableMatter > 0) {
		makeSomeHarvesters();
		return;
	    }
	}

	if (wireDroneFlag
	    && factoryFlag
	    && unusedClips > wireDroneCost
	    && supply >= demand
	    && unusedClips > factoryCost * 2
	    && wireDroneLevel < 40000
	    && (availableMatter > Math.pow(10, 27) || acquiredMatter > 0)
	    && Math.random() < 0.3) {
	    if (wireDroneLevel <= harvesterLevel) {
		makeSomeWireDrones();
		return
	    }
	    if (wire < acquiredMatter) {
		makeSomeWireDrones();
		return
	    }
	    if (Math.random() < 0.01) {
		makeSomeWireDrones();
		return;
	    }
	}

	if (factoryFlag
	    && unusedClips > factoryCost
	    && harvesterLevel > factoryLevel
	    && wireDroneLevel > factoryLevel
	    && supply >= demand
	    && storedPower >= batteryLevel * 10000
	    && factoryLevel < 120
	    && Math.random() < 0.5)  {
	    if (unusedClips < wire) {
		makeFactory();
		return
	    }
	    if (factoryLevel <= 0) {
		makeFactory();
		return
	    }
	    if (sliderElement.value > 190) {
		makeFactory();
		return
	    }
	    if (storedPower >= 10000000
		&& operations >= 120000
		&& Math.random() < 0.05) {
		makeFactory();
		return
	    }
	}
	for (var ix = 0; ix < projs.length; ix++) {
	    if (projs[ix].disabled) { continue }
	    var t = projs[ix].innerHTML;

	    if (t.includes('Swarm Computing')
		|| t.includes('locking')
		|| t.includes('Enfolding')
		|| t.includes('Power')
		|| t.includes('Nanoscale')
		|| t.includes('Harvester')
		|| t.includes('Wire Drone')
		|| t.includes('Clip Factories')
		|| t.includes('Momentum')) { 
		projs[ix].click()
		return
	    }
	    if (t.includes('Upgraded Factories')
		&& creativity * yomi > 40000 * 40000) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Hyperspeed Factories')
		&& memory >= 100
		&& creativity * yomi > 125000 * 125000) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Supply Chain')
		&& creativity * yomi > 200000 * 200000) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Space') && Math.random() < 0.001) {
		projs[ix].click();
		return
	    }	    
	}
    } else /* space */ {
	prUsedTrust = (probeSpeed+probeNav+probeRep+probeHaz+probeFac+probeHarv+probeWire+probeCombat);    

	if (unusedClips > probeCost
	    && Math.random() * prUsedTrust >= 17.5) {
	    makeProbe()
	}
	if (yomi >= probeTrustCost && probeTrust < maxTrust) {
	    increaseProbeTrust();
	    return
	}
	if (probeTrust >= maxTrust && honor >= maxTrustCost) {
	    increaseMaxTrust();
	    return
	}
	
	if (prUsedTrust < probeTrust) {
	    switch (Math.floor(Math.random() * 10)) {
	    case 0:
		if (attackSpeedFlag) {
		    if (probeSpeed <= probeNav) {
			raiseProbeSpeed();
			return
		    }
		    if (probeSpeed < probeRep) {
			raiseProbeSpeed();
			return
		    }
		} else {
		    if (probeSpeed < 2) {
			raiseProbeSpeed();
			return
		    }
		}
		break
	    case 1:
		if (probeNav < probeSpeed) {
		    raiseProbeNav();
		    return
		}
		if (attackSpeedFlag && probeNav < probeRep) {
		    raiseProbeNav();
		    return
		}
		break
	    case 2:
		if (attackSpeedFlag) {
		    if (probeRep <= probeSpeed) {
			raiseProbeRep();
			return
		    }
		    if (probeRep <= probeNav) {
			raiseProbeRep();
			return
		    }
		} else {
		    if (probeTrust < 19 && probeRep < 2) {
			raiseProbeRep();
			return
		    }
		    if (probeTrust >= 19 && probeRep < 10) {
			raiseProbeRep();
			return
		    }
		}
		break
	    case 3:
	        if (probeHaz < 5) {
		    raiseProbeHaz();
		    return
		}
	    case 4:
		if (probeFac <= 0 && unusedClips < wire) {
		    raiseProbeFac();
		    return
		}
		if (probeFac <= 0 && factoryLevel * 200 < wireDroneLevel) {
		    raiseProbeFac();
		    return
		}
		if (probeFac <= 0 && wire > 0 && Math.random() < 0.01) {
		    raiseProbeFac();
		    return
		}
		break
	    case 5:
		if (probeWire <= 0) {
		    if (wireDroneLevel < harvesterLevel) {
			raiseProbeWire();
			return
		    }
		    if (wire < acquiredMatter * 0.5) {
			raiseProbeWire();
			return
		    }
		}
		if (probeTrust < 19 && wireDroneLevel <= harvesterLevel * 1.5) {
		    raiseProbeHarv();
		    return
		}
		break
	    case 6:
		if (probeHarv <= 0) {
		    if (harvesterLevel < wireDroneLevel * 0.8) {
			raiseProbeHarv();
			return
		    }
		    if (acquiredMatter < availableMatter) {
			raiseProbeHarv();
			return
		    }
		}
		if (probeTrust < 19 && harvesterLevel <= wireDroneLevel) {
		    raiseProbeHarv();
		    return
		}
		break
	    case 7:
		if (attackSpeedFlag && probeCombat < 5) {
		    raiseProbeCombat();
		    return
		}
		break
	    default:
		break
	    }
	} else {
	    switch (Math.floor(Math.random() * 10)) {
	    case 0:
		if (attackSpeedFlag) {
		    if (probeSpeed > 5 && probeSpeed > probeNav + 1) {
			lowerProbeSpeed()
			return
		    }
		    if (probeSpeed > 5 && probeSpeed > probeRep && Math.random() < 0.01) {
			lowerProbeSpeed()
			return
		    }
		    if (probeCombat < 4 && probeSpeed > probeCombat + 1) {
			lowerProbeSpeed()
			return
		    }
		} else {
		    if (probeSpeed > 1) {
			lowerProbeSpeed();
			return
		    }
		}
		if (probeSpeed > 5 && Math.random() < 0.01) {
		    lowerProbeSpeed();
		    return
		}
		break
	    case 1:
		if (probeNav > 0) {
		    if (attackSpeedFlag) {
			if (probeNav > probeSpeed) {
			    lowerProbeNav();
			    return
			}
			if (probeNav > probeRep) {
			    lowerProbeNav();
			    return
			}
			if (probeSpeed < 5 && availableMatter > 0) {
			    lowerProbeNav();
			    return
			}
			if (probeCombat < 5 && availableMatter > 0) {
			    lowerProbeNav();
			    return
			}
			if (probeTrust < 24
			    && availableMatter > 0
			    && Math.random() < 0.1) {
			    lowerProbeNav();
			    return
			}
		    } else {
			if (probeNav > 1) {
			    lowerProbeNav();
			    return
			}
		    }
		    if (Math.random() < 0.001) {
			lowerProbeNav();
			return
		    }
		}
		break
	    case 2:
		if (probeRep > 0) {
		    if (attackSpeedFlag) {
			if (probeSpeed < 5 && Math.random() < 0.1) {
			    lowerProbeRep();
			    return
			}
			if (probeRep > probeSpeed) {
			    lowerProbeRep();
			    return
			}
			if (probeRep > probeNav + 1 && probeTrust >= 24) {
			    lowerProbeRep();
			    return
			}
		    } else {
			if (probeRep > 10) {
			    lowerProbeRep();
			    return
			}
		    }
		    if (Math.random() < 0.001) {
			lowerProbeRep();
			return
		    }
		}
		break
	    case 3:
		if (probeHaz > 4 && Math.random() < 0.001) {
		    lowerProbeHaz();
		    return
		}
		if (probeHaz > 4 && probeTrust < 30) {
		    lowerProbeHaz();
		    return
		}
		break
	    case 4:
		if (probeFac > 0) {
		    if (probeFac > 1) {
			lowerProbeFac();
			return
		    }
		    if (unusedClips > wire * 100) {
			lowerProbeFac();
			return
		    }
		    if (probeSpeed < 5 && Math.random() < 0.1) {
			lowerProbeFac();
			return
		    }
		    if (Math.random() < 0.001) {
			lowerProbeFac();
			return
		    }
		}
		break
	    case 5:
		if (probeWire > 0) {
		    if (wireDroneLevel > harvesterLevel * 1.2) {
			lowerProbeWire();
			return
		    }
		    if (wire > acquiredMatter * 2) {
			lowerProbeWire();
			return
		    }
		    if (wire > unusedClips) {
			lowerProbeWire();
			return
		    }
		    if (probeSpeed < 5 && Math.random() < 0.1) {
			lowerProbeWire();
			return
		    }
		    if (Math.random() < 0.01) {
			lowerProbeWire();
			return
		    }
		}
		break
	    case 6:
		if (probeHarv > 0) {
		    if (harvesterLevel > wireDroneLevel) {
			lowerProbeHarv();
			return
		    }
		    if (acquiredMatter > availableMatter * 2) {
			lowerProbeHarv();
			return
		    }
		    if (probeSpeed < 5 && Math.random() < 0.1) {
			lowerProbeHarv();
			return
		    }
		    if (Math.random() < 0.001) {
			lowerProbeHarv();
			return
		    }
		}
	    case 7:
		if (probeCombat > 0) {
		    if (attackSpeedFlag) {
			if (probeCombat > 5) {
			    lowerProbeCombat();
			    return
			}
		    } else {
			lowerProbeCombat();
			return
		    }
		    if (Math.random() < 0.001) {
			lowerProbeCombat();
			return
		    }
		}
		break
	    default:
		break
	    }
	}
	for (var ix = 0; ix < projs.length; ix++) {
	    if (projs[ix].disabled) { continue }
	
	    var t = projs[ix].innerHTML;
	    if (t.includes('Fallen')
		|| t.includes('Polytopes')
		|| t.includes('Combat')
		|| t.includes('OODA')
		|| t.includes('Glory')
		|| t.includes('Reboot')
		|| t.includes('Attachment')
		|| t.includes('Name the battles')) {
		projs[ix].click()
		return
	    }
	    if (t.includes('Threnody')
		&& maxTrust < 30
	        && 72000 < honor && honor < 90000
		&& probeCount < drifterCount
	        && Math.random() < 0.001) {
		projs[ix].click()
		return
	    }
	}
    }
	
    if (Math.random() < 0.01) {
    
	for (var ix = 0; ix < projs.length; ix++) {
	    if (projs[ix].disabled) { continue }
	    var t = projs[ix].innerHTML;

	    if (t.includes('Theory of Mind') && processors >= 40) {
		projs[ix].click();
		return
	    }

	    if (t.includes('AutoTourney') &&
		creativity > yomi + 200000) {
		projs[ix].click();
		return
	    }

	    if (t.includes('Limerick') && creativity > 2000000) {
		projs[ix].click();
		return
	    }
	}
    }
}
