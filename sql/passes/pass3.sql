// pass3 is a skeleton statement that defines the final table
// < with > is an additional with statement that defines the economic data
// < fields > are the fields to keep from tables b, c, d plus additional calculations
// < pass2Sample > is the table created by pass2
// < econFields > is the field to join on (e.g. zip, zip3)
<with>,
d AS (
  SELECT
    a.*,
    b.msaNameLoc AS msaLocName,
    f.pp_group = '' ? 'group3' : f.pp_group AS ppGroup,
    f.serv_mapped = '' ? 'other' : f.serv_mapped AS servMapped,
    x.fc_type AS fcType,
    x.fc_days AS fcDays,
    <fields>
  FROM
    <pass2Sample> AS a
  JOIN
    e AS b
  ON
    a.trgDt = b.month
    AND a.<econFields> = b.<econFields>
  JOIN
    e AS c
  ON
    a.aoDt = c.month
    AND a.<econFields> = c.<econFields>
  JOIN
    e AS d
  ON
    a.fpDt = d.month
    AND a.<econFields> = d.<econFields>
  JOIN
    (SELECT * FROM e WHERE e.month = toDate('2020-01-01')) AS x2020
  ON
    a.<econFields> = x2020.<econFields>
  LEFT JOIN f ON
    a.servicer=f.serv_name
// TODO: put this in econ table
  JOIN x on
    a.state = x.prop_st
)