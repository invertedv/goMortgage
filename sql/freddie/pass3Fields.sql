// calculated values that require the final table
100.0 * (trgHpi / orgHpi - 1) AS trgdHpi,
propVal * (trgHpi / orgHpi) AS trgPropVal,
propVal * (aoHpi / orgHpi) AS aoPropVal,
// constant dollar propVal (20200101)
propVal * (y20Hpi / orgHpi) AS y20PropVal,

100 * aoUpb / aoPropVal AS aoEltv,
100 * trgUpbExp / trgPropVal AS trgEltv,

rate - (rt15Wt * orgMortFix15 + (1-rt15Wt) * orgMortFix30) AS orgSpread,
toInt32(aoDq + fcstMonth > 12 ? 12 : aoDq + fcstMonth) AS trgDqMax,
30*(aoDq + fcstMonth) / trgFcDays < 1.5 ? 30*(aoDq + fcstMonth) / trgFcDays : 1.5 AS trgFcTime,

multiIf(a.term <= 180, 1, a.term >= 360, 0, (360 - a.term) / 180 ) AS rt15Wt,
rt15Wt * trgMortFix15 + (1-rt15Wt) * trgMortFix30 AS newRate,

newRate / 1200.0 AS newR,
newR > 0 ? newR * trgUpbExp / (1.0 - pow(1.0 + newR, (-a.term))) : trgUpbExp / a.term AS newPayment,
12.0 * (trgPayment - newPayment) AS trgRefiIncentive,

toFloat64(1200.0 * trgPayment / trgIncome50 > 100.0 ? 100.0 : 1200.0 * trgPayment / trgIncome50 ) AS trgPti50,

abs(1200.0 * (trgLbrForce - orgLbrForce) / (orgLbrForce * (aoAge+fcstMonth))) < 25 ? 1200.0 * (trgLbrForce - orgLbrForce) / (orgLbrForce * (aoAge+fcstMonth)) : 25 AS trgLbrGrowth,
toFloat64(fclProNet / trgPropVal) AS targetNetPro


