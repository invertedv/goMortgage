title: Model Data Through 2009
model: death
buildData: no
buildModel: yes

// data settings
strats1: state, aoDt
sampleSize1: 15000000
strats2: fcstMonth, month, aoDqCap6
sampleSize2: 3000000
where1: AND aoAge >= 0 AND aoDq >= 0 AND aoDq <= 24 AND aoUpb > 10000 AND mon.zb='00'
where2:  AND trgZb IN ('00', '01', '03', '09') AND fcstMonth > 0 
mtgDb: mtg.fannie
mtgFields: fannie
econDb: econGo.final
econFields: zip3

// model settings
target: targetDeath
targetType: cat
cat: propType, occ, standard, nsDoc, nsUw, coBorr, hasSecond, harp, aoMod, aoBap, canBe12, fcType, covid
cts: trgAge, term, units, trgUnempRate, trgEltv, trgRefiIncentive, spread, d120, pModDirect, fcTime
addlCats: targetAssist, targetDq, aoMaxDq12,fcstMonth, vintage, aoDqCap6, purpose, aoPrior30, aoPrior60, aoPrior90p, state, amType, aoDq, channel, servMapped
addlKeep: lnId, fico, aoPropVal, dti, aoMonthsCur,lbrGrowth
layer1: FC(size:20, activation:relu)
layer2: FC(size:20, activation:relu)
layer3: FC(size:3, activation:softmax)
batchSize: 50000
epochs: 2500
earlyStopping: 40
learningRateStart: .0005
learningRateEnd: .00025
l2Reg: 0.00005
modelQuery: SELECT %s FROM %s WHERE bucket < 10 AND year(month) < 2010
validateQuery: SELECT %s FROM %s WHERE bucket in (10,11,12,13,14) AND year(month) < 2010
assessQuery: SELECT %s FROM %s WHERE bucket in (15,16,17,18,19)

// output locations
outDir: /home/will/goMortgage/sequential/defPp/dq2010
pass1Strat: tmp.stratDeath1
pass1Sample: tmp.sampleDeath1
pass2Strat: tmp.stratDeath2
pass2Sample: tmp.sampleDeath2
modelTable: tmp.modelDeath
//tmpDb: tmp.temp
log: log.txt

// existing models that are inputs
inputModel: mod
modLocation: /home/will/goMortgage/sequential/mod/mod2010
modTargets: pModDirect:1

inputModel: dq
dqLocation: /home/will/goMortgage/sequential/dq/dq2010
dqTargets: d120: 4,5,6,7,8,9,10,11,12; current:0

// assessment
assessAddl: aoIncome50, aoEItb50, pPen, newPti90, newPti50, newPti10, trgPti10, trgPti90, ltv, aoMaxDq12, trgPti50, newPti50, expPti50, aoUpb, trgUpbExp, state, aoIncome90, fcstMonth, servMapped

assessNameaoDqPp: Prepay
assessTargetaoDqPp: 1
assessSliceraoDqPp: aoDqCap6

assessNameaoDqDef: Default
assessTargetaoDqDef: 2
assessSliceraoDqDef: aoDqCap6

// curves
curvesNameyrQtrPp: Target Quarter, Prepay
curvesTargetyrQtrPp: 1
curvesSliceryrQtrPp: trgYrQtr

curvesNameVintagePp: Vintage, Prepay
curvesTargetVintagePp: 1
curvesSlicerVintagePp: vintage

curvesNameforecastMonthPp: Forecast Month, Prepay
curvesTargetforecastMonthPp: 1
curvesSlicerforecastMonthPp: fcstMonth

curvesNameyrQtrDef: Target Quarter, Default
curvesTargetyrQtrDef: 2
curvesSliceryrQtrDef: trgYrQtr

curvesNameVintageDef: Vintage, Default
curvesTargetVintageDef: 2
curvesSlicerVintageDef: vintage

curvesNameforecastMonthDef: Forecast Month, Default
curvesTargetforecastMonthDef: 2
curvesSlicerforecastMonthDef: fcstMonth

// general
show: no
plotHeight: 1200
plotWidth: 1600
