## goMortgage
[![Go Report Card](https://goreportcard.com/badge/github.com/invertedv/goMortgage)](https://goreportcard.com/report/github.com/invertedv/goMortgage)
[![godoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/mod/github.com/invertedv/goMortgage?tab=overview)

## A Self-Service Program to Build Mortgage Models

### What is a mortgage model?

A mortgage model is a predictive model that forecasts some aspect of a mortgage's performance.
Common performance metrics are delinquency, default, severity and prepayment.

Such models are built on historical data and may be either at a loan level or pool level.
goMortgage builds loan-level models.

### About goMortgage

goMortgage is an app that builds mortgage forecasting models.

What aspects of mortgage performance can be modeled? Really, anything you can think of.  The software
is agnostic about the model target and features.

goMortgage takes a text file (*.gom) you create to direct all aspects of the modeling process
from building datasets to model assessment.

Since goMortgage is open source, goMortgage can be modified to suit your needs.
Out of the box, it is configured to work with Freddie and Fannie data. There are
instructions to set up your own data sources.

### Why goMortgage?

I wrote goMortgage for my own use. I made it open source because I thought others might find it useful, too.
Who?
- People who are more interested in the output of the model than building the infrastructure --
  say people on a trading desk or academics.  Especially those who want to focus on Fannie and Freddie,
  since I have packages ([Fannie](https://pkg.go.dev/github.com/invertedv/fannie),
  [Freddie](https://pkg.go.dev/github.com/invertedv/freddie)) to build this data.
- Those who don't have the bandwidth to start from scratch but enough to make the modifications for new data sources.


### Requirements

You need to be able to put up an instance of [ClickHouse](https://clickhouse.com/clickhouse).

Currently, goMortgage doesn't support building DNNs on a GPU, but the run speed has been fine. You'll want
64GB of RAM, though you could get by with less probably.

The raw files for Freddie and Fannie are quite large, so you'll want a few TBs of disk.

And, of course, you need to be able to compile Go.

### Documentation

For details, see the [docs](https://invertedv.github.io/goMortgage).
