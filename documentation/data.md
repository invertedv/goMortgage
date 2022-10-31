### Data


#### Sampling

In sampling data for this kind of model, there are two key points in time.

- The as-of age.  This is the loan age from which we are forecasting.  We know the status of the loan at this point.
- The forecast month.  This is number of months into the future we are forecasting the loan's status.

Conceptually, there are at least three ways to choose the as-of age of a loan.

1. Select at most one as-of age and forecast month for each loan.<br>
The potential advantage here is that a loan will appear with at most once, so we
needn't think about issues of correlation (as if loans weren't cross-correlated!).  However, care must
be taken in the sampling to prevent length-biased sampling as would result if one randomly chose an age from the
loan's history.
2. Create a table that has one row for each loan for each month it is on the books then randomly sample this table. Then
create a second table that has all possible forecast months from the age of each sampled loan and sample this table<br>
This approach avoids length-biased sampling but will result in loans appearing more than once in the data.
3. Stratify the sample based on loan age and then forecast month.  This approach avoids length-biased sampling.  It also
evens out the sample on these dimensions which should provide more stable estimates of the effects.

goMortage uses method 3. A similar issue applies to choosing the target date.

Stratifying the sample also opens up the opportunity to stratify along other dimensions.
For instance, to avoid building a model dominated by loans in California, one can stratify on state.
Or, you could build a California-only model using the WHERE1 key in the specs file.  Stratifying along
loan age and forecast month alone may produce a sample that is concentrated on certain vintages or performance
periods.




goMortgage does this sampling in two stages. The first stage generates a table of loans sampled to the
as-of date.  This table is then sampled to pick target dates.

One can also specify other fields on which to stratify.

There is a third stage to the data build. This joins the sampled loans to other (economic) data.  The table
is joined by geo (e.g. zip3, state, zip) at four time periods:

1. The origination date.
2. The as-of date.
3. The target date.
4. January 2020.

The utility of the first 3 is self-evident.  Why January 2020? So that we have a baseline to normalize
values so that the model isn't confused by trends in (say) house prices.


