// pass2 is a skeleton statement that defines the table that will be sampled in pass2
// < fields > are the fields to be kept in the output sample table
//  <mtgDb > is the ClickHouse table of mortgage loans
// < pass1Sample > is the sample table produced by pass1
// < where > are additional restrictions
WITH d AS (
    SELECT
   <fields>
FROM
        <mtgDb> AS lns ARRAY JOIN monthly AS mon
    JOIN
        <pass1Sample> AS s
ON
    lns.lnId = s.lnId
WHERE
    fcstMonth >= 0
  AND fcstMonth <= 180
  AND trgRemTerm > 0
    <where>
    )