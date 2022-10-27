title: Model Data Through 2015
model: dq
buildData: no
buildModel: yes

// data settings
strats1: state, aoDt
sampleSize1: 30000000
strats2: fcstMonth, month, aoDqCap6
sampleSize2: 3000000
where1: AND aoAge >= 0 AND aoDq >= 0 AND aoDq <= 24 AND aoUpb > 10000 AND mon.zb='00'
where2:  AND trgZb='00' AND fcstMonth > 0 
mtgDb: mtg.fannie
mtgFields: fannie
econDb: econGo.final
econFields: zip3

// model settings
target: targetDq
targetType: cat
cat: purpose, propType, occ, amType, standard, nsDoc, nsUw, coBorr, hasSecond, aoPrior30, aoPrior60, aoPrior90p, harp, aoMod, aoBap, channel, covid, fcType, pPen36
cts: fico, trgAge, term, y20PropVal, units, dti, trgUnempRate, trgEltv, aoMonthsCur, trgPti50, trgRefiIncentive, lbrGrowth, spread, pMod
emb:  aoDq: 5, state: 4, servMapped: 4
addlCats: targetAssist, aoMaxDq12, vintage, aoDqCap6, fcstMonth, numBorr
layer1: FC(size:40, activation:relu)
layer2: FC(size:20, activation:relu)
layer3: FC(size:10, activation:relu)
layer4: FC(size:13, activation:softmax)
batchSize: 5000
epochs: 2000
earlyStopping: 40
learningRateStart: .0003
learningRateEnd: .0001
modelQuery: SELECT %s FROM %s WHERE bucket < 10 AND year(month) < 2016
validateQuery: SELECT %s FROM %s WHERE bucket in (10,11,12,13,14) AND year(month) < 2016
assessQuery: SELECT %s FROM %s WHERE bucket in (15,16,17,18,19)

// output locations
outDir: /home/will/goMortgage/sequential/dq/dq2016
pass1Strat: tmp.stratDq1
pass1Sample: tmp.sampleDq1
pass2Strat: tmp.stratDq2
pass2Sample: tmp.sampleDq2
modelTable: tmp.modelDq
log: log.txt

// existing models that are inputs
inputModel: mod
modLocation: /home/will/goMortgage/sequential/mod/mod2016
modTargets: pMod:1

// assessment
assessAddl: aoIncome50, aoEItb50,  trgPti10, trgPti90, ltv, aoMaxDq12, trgUpbExp, state, aoIncome90, msaLoc, aoPropVal, trgPropVal, vintage

assessNameaoDq: D120+
assessTargetaoDq: 4,5,6,7,8,9,10,11,12
assessSliceraoDq: aoDqCap6

assessNameOcc: D120+
assessTargetOcc: 4,5,6,7,8,9,10,11,12
assessSlicerOcc: occ

// curves
curvesNameyrQtrD120: Target Quarter, D120+
curvesTargetyrQtrD120: 4,5,6,7,8,9,10,11,12
curvesSliceryrQtrD120: trgYrQtr

curvesNamevintageD120: Vintage, D120+
curvesTargetvintageD120: 4,5,6,7,8,9,10,11,12
curvesSlicervintageD120: vintage

curvesNamefmD120: Forecast Month, D120+
curvesTargetfmD120: 4,5,6,7,8,9,10,11,12
curvesSlicerfmD120: fcstMonth

curvesNameyrQtrD30: Target Quarter, D30
curvesTargetyrQtrD30: 1
curvesSliceryrQtrD30: trgYrQtr

curvesNameyrQtrCur: Target Quarter, Current
curvesTargetyrQtrCur: 0
curvesSliceryrQtrCur: trgYrQtr

// general
show: no
plotHeight: 1200
plotWidth: 1600
