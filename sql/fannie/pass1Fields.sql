// additional fields for pass1 beyond what is stratfied on
lnId,
msa = '00000' ? state : msa AS msaLoc,
state,
mon.ageFpDt AS aoAge,
toInt32(aoAge/12) AS ageYr,
cltv > ltv ? 'Y' : 'N' AS hasSecond,
numBorr > 1 ? 'Y' : 'N' AS coBorr,
aoAge <= 36 AND pPen = 'Y' ? 'Y' : 'N' AS pPen36,

mon.month AS aoDt,
mon.dq AS aoDq,
aoDq > 6 ? 6 : aoDq AS aoDqCap6,
mon.upb AS aoUpb,
mon.mod AS aoMod,

// take the most recent servicer that is not unknown
arrayFilter((x,y)->x!='unknown' and y < aoDt, monthly.servicer, monthly.month) as notUnk,
length(notUnk) > 0 ? arrayElement(notUnk, length(notUnk)) : 'other' AS servicer,

mon.bap AS aoBap,
mon.zb AS aoZb,

mon.curRate AS aoRate,
aoRate > 0 ? aoRate / 1200.0 : 0.001 / 1200.0 AS aoR,
term - aoAge AS aoRemTerm,
aoR * aoUpb / (1.0 - pow(1.0 + aoR, (-aoRemTerm))) AS aoPayment,

dateSub(month,12,mon.month) AS lag12,
toInt32(arrayMax(arrayMap((dt,dq)->dt>=lag12 and dt < mon.month ? (dq > 6 ? 6 : dq) : 0, monthly.month, monthly.dq))) as aoMaxDq12,

toInt32(arraySum(arrayMap((dt,dq)->dt>=lag12 and dt < mon.month ? (dq = 1 ? 1 : 0) : 0, monthly.month, monthly.dq))) as aoTimes30,
toInt32(arraySum(arrayMap((dt,dq)->dt>=lag12 and dt < mon.month ? (dq = 2 ? 1 : 0) : 0, monthly.month, monthly.dq))) as aoTimes60,
toInt32(arraySum(arrayMap((dt,dq)->dt>=lag12 and dt < mon.month ? (dq >= 3 ? 1 : 0) : 0, monthly.month, monthly.dq))) as aoTimes90p,
toInt32(arraySum(arrayMap((dt, dq)->dt < mon.month AND dq=0 ? 1 : 0, monthly.month, monthly.dq))) AS aoMonthsCurUc,
aoMonthsCurUc > 36 ? 36 : aoMonthsCurUc AS aoMonthsCur,

aoTimes30 > 0 ? 'Y' : 'N' AS aoPrior30,
aoTimes60 > 0 ? 'Y' : 'N' AS aoPrior60,
aoTimes90p > 0 ? 'Y' : 'N' AS aoPrior90p
