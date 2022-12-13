## goMortgage
[![Go Report Card](https://goreportcard.com/badge/github.com/invertedv/goMortgage)](https://goreportcard.com/report/github.com/invertedv/goMortgage)
[![godoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/mod/github.com/invertedv/goMortgage?tab=overview)

### Summary
goMortgage is an app for building models to forecast mortgage performance.  It is the culmination of many years
experience building these models. See my [linkedIn](www.linkedin.com/in/will-alexander-data-scientist) profile for 
more about me. 

goMortgage is unusual in many dimensions. It's a one-stop modeling platform that will build both the modeling
dataset and the model.

goMortgage:

   - **Is a special purpose app.**<br>
goMortgage is an app.  The specification of the model ('specs' file) is provided to it in a text file.  Changing features of
the model, the data used to build and validate the model, graphical evaluations of the model are all specified
in this file.  These are things that change often during a modeling project.  With goMortgage, one isn't hacking
into the code to do this.
<br><br>
   - **Provides multiple diagnostics**<br>
There are 3 base types of graphical diagnostics produced by goMortgage:
     - Curves.  These are the average model output and target plotted vs another field (e.g. time).
     - KS and Decile plots for categorical targets.
     - Segmented plots.  These plots slice the data by the levels of specific field, plottings the model and target
     average for each slice.  
     - Marginal plots.  These plots indicate the relation between a feature and the model output.
<br><br>
  - **produces live plots**<br>
The graphs are produced using Plotly, meaning they are *live*.  Open them in your browser and hover over anything that
looks interesting for more information.
<br><br>
- **incorporates flexible validations**<br>
You choose the slices and targets to generate diagnostics.  Want to check how the model performs state-by-state?
No problem.  Vintages? Score buckets? Also, no problem.
  <br><br>
- **builds cutting-edge models**<br>
These models are DNN models.  Helpful options such as embeddings are supported. The target can be either
categorical or continuous.  The build algorithm is the 
[gorgonia](https://pkg.go.dev/gorgonia.org/gorgonia@v0.9.17#section-readme) package.
  <br><br>
- **produces lightweight models**<br>
The models produced by goMortgage consist of only 3 text files.  Two of these specify the model and the third
the input features.
<br><br>
- **produces lightweight model builds**<br>
goMortgage is quite frugal with memory--using millions of observations is possible even in 32GB.
However, if memory does prove limiting, the data can be streamed to the model build.
<br><br>
- **consdolidates its output**<br>
The output includes the model, diagnostics, a log file and the specs file the run used.
<br><br>
- **builds its own data**<br>
goMortgage will build the modeling dataset from source files.
<br><br>
- **is written in Go**<br>
Go is a 21st century language.  This [post](https://yourbasic.org/golang/advantages-over-java-python/) has a nice
desciption of its advantages.
<br><br>
- **is open source**<br>
There are example scripts in the [scripts]() directory. The scripts are configured to work with the 
ClickHouse table for the Fannie Mae data produced by [this]() package.  Since goMortgage is open source,
it can be customized to other data sets by you.
 <br><br>
- **and more!**
  <br><br>
  Note, that all this being said, goMortgage can be used to fit really any kind of model.  The key restriction
  is that the data resides in ClickHouse.

## Documentation

For more details, see the [docse]().
