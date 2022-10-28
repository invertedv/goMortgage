## The Specification File
The specificaton (specs) file tells goMortgage what to do.  There are examples in the scripts directory.
Some fields in the specs file are mandatory, others optional.

The format for entries is:

     <key>:<value>

There are two basic keys that drive the process.  These are:

- buildData   : yes/no
- buildModel  : yes/no

### Data Build Keys
Building the data is a three-pass process.  See {Model Building} for more details. The following keys are required for building data:

- strats1: field list<br>
These are the fields to stratify on for the first pass (which selects loans and as-of dates).
- sampleSize1: int<br>
The target sample size for pass 1;
- where1: clause<br>
a "where" clause to restrict the selection.
- pass1Strat: table name<br>
Name of ClickHouse table to create with the stratification summary.

- pass1Sample: table name<br>
Name of the ClickHouse table to create with the sampled loans.
<br><br>
- strats2: field list<br>
  These are the fields to stratify on for the second pass (which selects loans and as-of dates).
- sampleSize2: int<br>
  The target sample size for pass 2;
- where2: clause<br>
  a "where" clause to restrict the selection.
- pass2Strat: table name<br>
  Name of ClickHouse table to create with the stratification summary.
- pass2Sample: table name<br>
  Name of the ClickHouse table to create with the sampled loans.
<br><br>
- mtgDb: table name
<br>The ClickHouse table that has the loan-level detail.
- mtgFields: name<br>
The value here is a keyword.  Currently, valid values are "fannie" and "freddie". These refer to values specified
within goMortgage.  This is how goMortgage knows what fields to expect in the table.
See {Adding Sources} for details on adding a source.
- econDb:<br>
The ClickHouse table that has the economic data.
- econFields: <field><br>
Geo field that is the join field between the mortgage data and the economic data (*e.g.* zip).
<br><br>
- modelTable: table name<br>
Name of the ClickHouse table to create with the model-build sample.

### Model Build Keys

The required keys are:

- target: field name<br>
The field that is the target (dependent variable).
- cat: field list<br>
A comma-separated list of categorical (one-hot) features.
- cts: field list<br>
A comma-separated list of continuous features.
- emb: field list<br>
A comma-separated list of embedding features.  Each entry is also a key/val pair of the name of the feature
followed by the embedding dimension (field:dim).
- layer<n> : layer specification<br>
The model layers are numbered starting with 1.  The specification of the layer follows that used by the
[seafan](https://pkg.go.dev/github.com/invertedv/seafan) package.  For instance, if the first layer
after the inputs is a fully connected layer with a RELU activation and 10 outputs is specified by

      layer1: FC(10, activation=relu)
- epochs: int<br>
The maximum number of epochs to run through.
- batchsize: int<br>
The batch size for the model build optimizer.
- earlyStopping: int<br>
If the cost function evaluated on the validation data doesn't decline for 'earlyStopping' epochs, the fit is
terminated.
- learningRateStart: float<br>
The learning rate for epoch 1.
- learningRateStart: float<br>
The learning rate for the last potential epoch. 

      The learning rate declines linearly from learningRateStart at epoch 1 to learningRateEnd at epoch 'epochs'.
- modelQuery: query<br>
The query to pull the model-build data from 'modelTable'.
- validateQuery: query<br>
The query to pull the validation data from 'modelTable'.  The validation data is used only for determining
early stopping.
- assessQuery: query<br>
The query to pull the assessment data from 'modelTable'.  The assessment data is used only for post-model-build
assessment of the model fit.

      The queries have two place holders "%s".  goMortgage replaces the first %s with a list of the needed fields
      and the second with 'modelTable'.

Optional model-build keys:

- Saving the assessment data<br>
You can save the data used for the assessment along with model outputs back to ClickHouse.
There are two keys required to do this.
    - saveTable: table name<br>ClickHouse table to save the assess data to.
    - saveTableTargets: name1:target list 1; name2:target list 2.<br>
  The 'name' is the name of the field in the output field.  The target list is a list of comma-separated
  columns of the model output to sum to create the field.  For instances, if the model is a softmax with
  5 output columns, then

          first: 0; last2: 3,4

      will create a field called 'first' in the output table that is the first level of the targe 
      and another field called "last2" in the output table that is the sum of the probabilities of the target being
      its last 2 values.  If the target is continuous, then only column 0 is available.
- addlCats: field list<br>
is a comma-separated list of fields to treat as categorical. This may be needed for other input models
(see below) or because they are used as a slicer for curves.
- addlKeep: field list<br>
is a comma-separated list of additional fields to include in the 'saveTable'.  For instance, loan number.
- 