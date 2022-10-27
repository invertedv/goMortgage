// pass1 is a skeleton statement that defines the table that will be sampled in pass1
// < fields > are the fields that are stratified on plus any others we need to define now.
// < mtgDb > is the ClickHouse table of mortgage loans
// < goodloan > defines a loan that has high enough data quality to be considered.
// < where > are additional conditions.
WITH d AS (
    SELECT
   <fields>
FROM
        <mtgDb> ARRAY JOIN monthly AS mon
WHERE
    <goodLoan>
    <where>
)