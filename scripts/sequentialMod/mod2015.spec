title: Model Data Through 2014
model: mod
buildData: no
buildModel: yes

// data settings
strats1: state, aoDt
sampleSize1: 15000000
strats2: fcstMonth, month, aoDq
sampleSize2: 2000000
where1: AND aoAge >= 0 AND aoDq >= 0 AND aoDq <= 24 AND aoUpb > 10000 AND mon.zb='00' AND aoMod='N' AND aoBap in ['N', '7', '9']
where2:  AND trgZb='00' AND fcstMonth > 0 
mtgDb: mtg.fannie
mtgFields: fannie
econDb: econGo.final
econFields: zip3

// model settings
target: targetAssist
targetType: cat
cat: purpose, propType, occ, amType, standard, nsDoc, nsUw, coBorr, hasSecond, aoPrior30, aoPrior60, aoPrior90p, harp, channel
cts: fico, trgAge, term, trgUpbExp, units, dti, trgUnempRate, trgEltv, aoMonthsCur, trgPti50, trgRefiIncentive, lbrGrowth, spread
emb:  aoDq: 5, state: 2, servMapped: 2
addlCats: aoMaxDq12, vintage, aoDqCap6, fcstMonth
addlKeep: lnId, fcstMonth
layer1: DropOut(.1)
layer2: FC(size:20, activation:relu)
layer3: FC(size:2, activation:softmax)
batchSize: 5000
epochs: 500
earlyStopping: 40
learningRateStart: .0005
learningRateEnd: .00025
modelQuery: SELECT %s FROM %s WHERE bucket < 10 AND year(month) < 2015
validateQuery: SELECT %s FROM %s WHERE bucket in (10,11,12,13,14) AND year(month) < 2015
assessQuery: SELECT %s FROM %s WHERE bucket in (15,16,17,18,19)

// output locations
outDir: /home/will/goMortgage/sequential/mod/mod2015
pass1Strat: tmp.strat1Mod
pass1Sample: tmp.sample1Mod
pass2Strat: tmp.strat2Mod
pass2Sample: tmp.sample2Mod
modelTable: tmp.modelMod
log: log.txt

// save Assess Data + model output
saveTable: tmp.outMod2015
saveTableTargets: pMod: 1

// assessment
assessAddl: fcstMonth, aoIncome50, aoEItb50, pPen, newPti90, newPti50, newPti10, trgPti10, trgPti90, ltv, aoMaxDq12, trgPti50, newPti50, expPti50, aoUpb, state, aoIncome90, msaLoc

assessNameaoDq: pModified
assessTargetaoDq: 1
assessSliceraoDq: aoDqCap6

// curves
curvesNameyrQtr: Target Quarter, pModified
curvesTargetyrQtr: 1
curvesSliceryrQtr: trgYrQtr

curvesNamevintage: Vintage, pModified
curvesTargetvintage: 1
curvesSlicervintage: vintage

curvesNamefcstMonth: Forecast Month, pModified
curvesTargetfcstMonth: 1
curvesSlicerfcstMonth: fcstMonth

// general
plotShow: no
plotHeight: 1200
plotWidth: 1600



