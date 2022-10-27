// calculated values that require the final table
trgHpi / orgHpi - 1 AS dHpi,
propVal * (dHpi + 1) AS trgPropVal,
propVal * (aoHpi / orgHpi) AS aoPropVal,
100 * aoUpb / aoPropVal AS aoELtv,
100 * trgUpbExp / trgPropVal AS trgEltv,
100 * aoIncome10 / aoUpb AS aoEItb10,
100 * aoIncome50 / aoUpb AS aoEItb50,
100 * aoIncome90 / aoUpb AS aoEItb90,

// constant dollar propVal (20200101)
propVal * (y20Hpi / orgHpi - 1) AS y20PropVal,

multiIf(a.term <= 180, 1, a.term >= 360, 0, (360 - a.term) / 180 ) AS rt15Wt,
rt15Wt * trgMortFix15 + (1-rt15Wt) * trgMortFix30 AS newRate,
rate - (rt15Wt * orgMortFix15 + (1-rt15Wt) * orgMortFix30) AS spread,
aoDq + fcstMonth >= 12 ? 'Y' : 'N' AS canBe12,
30*(aoDq + fcstMonth) / fcDays < 1.5 ? 30*(aoDq + fcstMonth) / fcDays : 1.5 AS fcTime,
fcType,

newRate / 1200.0 AS newR,
newR > 0 ? newR * trgUpbExp / (1.0 - pow(1.0 + newR, (-a.term))) : trgUpbExp / a.term AS newPayment,
12.0 * (trgPayment - newPayment) AS trgRefiIncentive,
trgEltv > 95 ? 0 : trgRefiIncentive AS trgRICapped,
(trgRate - newRate) * trgUpbExp AS trgIntIncentive,

1200.0 * newPayment / trgIncome10 AS newPti10,
1200.0 * newPayment / trgIncome25 AS newPti25,
1200.0 * newPayment / trgIncome50 AS newPti50,
1200.0 * newPayment / trgIncome75 AS newPti75,
1200.0 * newPayment / trgIncome90 AS newPti90,

1200.0 * trgPayment / trgIncome10 AS trgPti10,
1200.0 * trgPayment / trgIncome25 AS trgPti25,
toFloat64(1200.0 * trgPayment / trgIncome50 > 100.0 ? 100.0 : 1200.0 * trgPayment / trgIncome50 ) AS trgPti50,
1200.0 * trgPayment / trgIncome75 AS trgPti75,
1200.0 * trgPayment / trgIncome90 AS trgPti90,

trgPti50 < newPti50 ? trgPti50 : newPti50 AS bestPti50,
aoDq <= 1 ? bestPti50 : trgPti50 AS expPti50,
100 * trgIncome50 / trgUpbExp AS trgEItb50,

abs(1200.0 * (trgLbrForce - aoLbrForce) / (aoLbrForce * fcstMonth)) < 25 ? 1200.0 * (trgLbrForce - aoLbrForce) / (aoLbrForce * fcstMonth) : 0 AS lbrGrowth,
toFloat64(fclProNet / trgPropVal) AS targetNetPro


