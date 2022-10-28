title: Net Proceeds Model with servMapped:4 and state:4
model: netpro
buildData: no
buildModel: yes

// data settings
strats1: state
sampleSize1: 1000000
strats2: aoZb
sampleSize2: 750000
where1: AND mon.zb IN ('03', '09')
where2:  AND fcstMonth=0 
mtgDb: mtg.fannie
mtgFields: fannie
econDb: econGo.final
econFields: zip3

// model settings
target: targetNetPro
targetType: cts
cat: propType, occ, fcType, shortSale, covid
cts: trgPropVal, trgAge, trgUnempRate, units
emb: servMapped: 4, state: 4
addlCats: vintage, aoDqCap6, trgYrQtr
addlKeep: lnId, aoDt, trgZb
layer1: FC(size:20, activation:relu)
layer2: FC(size:1)
batchSize: 5000
epochs: 2000
earlyStopping: 40
learningRateStart: .0005
learningRateEnd: .0001
modelQuery: WITH d AS (SELECT %s FROM %s WHERE bucket < 10  limit 2500000) select * from d where 1=1
validateQuery: WITH d AS (SELECT %s FROM %s WHERE bucket in (10,11,12,13,14) limit 1250000) select * from d where 1=1
assessQuery: WITH d AS (SELECT %s FROM %s WHERE bucket in (15,16,17,18,19) limit 1250000) select * from d where 1=1   

// output locations
outDir: /home/will/goMortgage/netPro
pass1Strat: tmp.stratNP1
pass1Sample: tmp.sampleNP1
pass2Strat: tmp.stratNP2
pass2Sample: tmp.sampleNP2
modelTable: tmp.modelNetPro
log: log.txt

// save Assess Data + model output
saveTable: tmp.outNetPro
saveTableTargets: netProHat: 0

// assessment
assessAddl: aoIncome50, cltv, ltv, state, aoIncome90, servMapped, msaLocName, vintage, y20PropVal

assessNamePropType: Property Type
assessTargetPropType: 0
assessSlicerPropType: propType

// curves
curvesNameyrQtrPp: Target Quarter, Net Proceeds
curvesTargetyrQtrPp: 0
curvesSliceryrQtrPp: trgYrQtr

curvesNameVintagePp: Vintage, Net Proceeds
curvesTargetVintagePp: 0
curvesSlicerVintagePp: vintage

// general
show: no
plotHeight: 1200
plotWidth: 1600
