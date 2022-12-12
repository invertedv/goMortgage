// econ table is at the zip level.  This aggregates to the zip3 level
WITH e AS (
    SELECT
        zip3,
    month,
    avg(zip3Hpi) AS hpi,
    sum(unempRate * lbrForce) / sum(lbrForce) AS unempRate,
    sum(lbrForce) AS lbrForceTot,
    max(mortFix30) AS mortFix30,
    max(mortFix15) AS mortFix15,
    max(treas10) AS treas10,
    max(q10) AS income10,
    max(q25) AS income25,
    max(q50) AS income50,
    max(q75) AS income75,
    max(q90) AS income90,
    max(msaName) = '' ? max(state) : max(msaName) AS msaNameLoc,
    max(fcDays) AS fcDays,
    max(fcType) AS fcType
FROM
    econGo.final
GROUP BY zip3, month)
