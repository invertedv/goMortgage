### Data


#### Sampling

In sampling data for this kind of model, there are two key points in time.

- The as-of date.  This is the date at which we are forecasting from.  We know the status of the loan at this point.
- The target date.  This is the date for which we want a forecast of the loan's status.

The source loan-level dataset is sampled in two stages.  The first stage sample selects as-of dates.
The second stage selects the target dates.

Conceptually, there are at least three ways to choose the as-of date.

1. Select at most one date for each loan.<br>
The potential advantage here is that a loan will appear with at most one as-of date, so we
needn't think about issues of correlation (as if loans weren't cross-correlated!).  However, care must
be taken in the sampling to prevent length-biased sampling.
2. Randomly sample the time series of each loan to choose 1 or more as-of dates.<br>
This approach avoids length-biased sampling but will result in loans appearing more than once in the data.
3. Stratified.

goMortage uses method 2. A similar issue applies to choosing the target date.

goMortgage does this sampling in two stages. The first stage generates a table of loans sampled to the
as-of date.  This table is then sampled to pick target dates.

One can also specify other fields on which to stratify.
For instance, to avoid building a model dominated by loans in California, one can stratify on state.
Or, you could build a California-only model using the WHERE1 key in the specs file.

There is a third stage to the data build. This joins the sampled loans to other (economic) data.  The table
is joined by geo (e.g. zip3, state, zip) at four time periods:

1. The origination date.
2. The as-of date.
3. The target date.
4. January 2020.

The utility of the first 3 is self-evident.  Why January 2020? So that we have a baseline to normalize
values so that the model isn't confused by trends in (say) house prices.


