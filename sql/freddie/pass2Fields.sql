// this field list is used if there is no window key in the .gom file
// these are fields from pass1 that carry over to the sample2
dateDiff('month', s.aoDt, trgDt) AS fcstMonth,
s.msaLoc,
s.aoDt,
s.aoAge,
s.aoDq,
s.aoDqCap6,
s.aoUpb,
s.aoMod,
s.hasSecond,
s.coBorr,
s.pPen36,
s.aoMaxDq12,
s.aoMonthsCur,
s.aoTimes30,
s.aoTimes60,
s.aoTimes90p,
s.aoPrior30,
s.aoPrior60,
aoPrior90p,
s.aoPayment,
s.aoRate,
s.aoZb,
s.noGroups,

trgRate > 0 ? trgRate / 1200.0 : 0.01 / 1200.0 AS trgR,
term - trgAge AS trgRemTerm,

s.aoUpb * pow(1.0 + s.aoR, fcstMonth) - s.aoPayment * (pow(1.0 + s.aoR, fcstMonth) - 1.0) / s.aoR AS trgUpbExp,

trgRate > 0 ? trgR * trgUpbExp / (1.0 - pow(1.0 + trgR, (-trgRemTerm))) : aoR * trgUpbExp / (1.0 - pow(1.0 + s.aoR, (-trgRemTerm)))  AS trgPayment,

concat(toString(year(trgDt)), 'Q', toString(quarter(trgDt))) AS trgYrQtr,
year(trgDt)>=2019 ? toString(year(trgDt)) : 'Before 2019' AS periods,
dateDiff('month', toDate('2020-04-01'),trgDt) >= 0 AND dateDiff('month', toDate('2022-04-01'),trgDt) <= 0 ? 'Y' : 'N' AS covid,
trgZb = '03' ? 'Y' : 'N' AS shortSale,

// Note, Freddie mod flag means modified this month
toInt32(trgMod=='Y' ? 1 : 0) AS targetMod,
toInt32(trgMod=='Y' OR trgPayPl in ['Y']) AS targetAssist,
toInt32(multiIf(trgDq < 0, 0, trgDq > 12, 12, trgDq)) AS targetDq,
toInt32(multiIf(trgZb='00', 0, trgZb='01', 1, 2)) AS targetDeath,
toInt32(multiIf(trgZb='00', targetDq, trgZb='01', 13, 14)) AS targetStatus