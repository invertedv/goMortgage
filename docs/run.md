---
layout: default
title: Running Your Model
nav_order: 13
---

## Running Your Model

Great, so you have a model!  How do you actually run it?  Well, there's a package for that.
The package
<a href="https://pkg.go.dev/github.com/invertedv/moru" target="_blank" rel="noopener noreferrer" >moru</a>
was built to run these models.  

There are two functions for running a goMortgage model using moru:

- ScoreToPipe: Adds your model(s) to an existing pipeline.
- ScoreToTable: Takes an input ClickHouse table, adds the model outputs, and then writes the original table plus
the model output(s) to another ClickHouse table.  ScoreToTable can be run using concurrency.



