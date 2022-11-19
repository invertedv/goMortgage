package main

import (
	"fmt"
	"math"
	"os"
	"time"

	grob "github.com/MetalBlueberry/go-plotly/graph_objects"
	"github.com/invertedv/chutils"
	sea "github.com/invertedv/seafan"
)

// assessModel drives the model assessment based on the user specs in the .gom file.
func assessModel(specs specsMap, conn *chutils.Connect, log *os.File) error {
	var (
		fts        sea.FTypes
		assessPipe sea.Pipeline
		e          error
	)

	start := time.Now()
	logger(log, fmt.Sprintf("starting assessment @ %s", start.Format(time.UnixDate)), true)

	if fts, e = sea.LoadFTypes(specs.modelDir() + "fieldDefs.jsn"); e != nil {
		return e
	}

	obsFt := fts.Get(specs.target())

	// assess pipeline
	if assessPipe, e = newPipe(specs.getQuery("assess"), "Assess data", specs, 0, fts, conn); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\n%v", assessPipe), false)

	// plot curves
	for _, curve := range specs.slicer("curves") {
		sl := curve
		if e := curves(assessPipe, specs, obsFt, &sl); e != nil {
			return e
		}
	}

	// Marginal and KS/Decile/SegPlot plots
	for _, slice := range specs.slicer("assess") {
		sl := slice // bad to pass for var as a pointer

		baseFt := assessPipe.GetFType(slice.feature) // this may not be in fts
		if e := marginal(specs, &sl, baseFt, obsFt, fts, conn); e != nil {
			return e
		}
		if e := assess(assessPipe, specs, obsFt, &sl, log); e != nil {
			return e
		}
	}

	// save assess data & model values back to ClickHouse
	if e := export(assessPipe, specs, obsFt, conn); e != nil {
		return e
	}

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("assessment run time: %0.1f minutes", elapsed), true)

	return nil
}

// curves outputs curves of fitted & actual versus the values of another feature.  Often, the other feature
// will be related to time.
func curves(pipe sea.Pipeline, specs specsMap, obsFt *sea.FType, curveSpec *slices) error {
	if pipe.GetFType(curveSpec.feature) == nil {
		return fmt.Errorf("feature %s not in pipeline", curveSpec.name)
	}

	pd := &sea.PlotDef{
		Show:     specs.plotShow(),
		Title:    fmt.Sprintf("%s<br>%s", specs.title(), curveSpec.name),
		XTitle:   curveSpec.feature,
		YTitle:   "Fit and Actual",
		STitle:   "",
		Legend:   true,
		Height:   specs.plotHeight(),
		Width:    specs.plotWidth(),
		FileName: fmt.Sprintf("%s%s.html", specs["curvesDir"], curveSpec.shortName),
	}

	modelLoc := specs.modelDir() + "model"

	nnP, e := sea.PredictNN(modelLoc, pipe, false)
	if e != nil {
		return e
	}

	nCat := nnP.OutputCols()

	baseSl, e := sea.NewSlice(curveSpec.feature, minCount, pipe, nil)
	if e != nil {
		return e
	}

	xVals := make([]any, 0)
	obs := make([]float64, 0)
	fit := make([]float64, 0)

	for baseSl.Iter() {
		xVals = append(xVals, baseSl.Value())
		baseSlicer := baseSl.MakeSlicer()

		fitSlice, e := sea.Coalesce(nnP.FitSlice(), nCat, curveSpec.target, false, false, baseSlicer)
		if e != nil {
			return e
		}

		obsSlice, e := sea.Coalesce(nnP.ObsSlice(), nCat, curveSpec.target, obsFt.Role == sea.FRCat, false, baseSlicer)
		if e != nil {
			return e
		}

		desc, e := sea.NewDesc(nil, "temp")
		if e != nil {
			return e
		}

		desc.Populate(sea.UnNormalize(obsSlice, obsFt), false, nil)
		obs = append(obs, desc.Mean)

		desc.Populate(sea.UnNormalize(fitSlice, obsFt), false, nil)
		fit = append(fit, desc.Mean)
	}

	trAct := &grob.Scatter{
		Type: grob.TraceTypeScatter,
		X:    xVals,
		Y:    obs,
		Name: "Actual",
		Mode: grob.ScatterModeLines,
		Line: &grob.ScatterLine{Color: "black"},
	}
	fig := &grob.Fig{Data: grob.Traces{trAct}}

	trFit := &grob.Scatter{
		Type: grob.TraceTypeScatter,
		X:    xVals,
		Y:    fit,
		Name: "Fitted",
		Mode: grob.ScatterModeLines,
		Line: &grob.ScatterLine{Color: "red"},
	}
	fig.AddTraces(trFit)

	return sea.Plotter(fig, nil, pd)
}

// marginal generates the marginal response plots of the features in the model.
//
//   - exist: directory of existing models
//   - val : feature slice on which to segment the output
//   - baseFts: FType of the feature specified by val
//   - fts : Ftypes of features in model.
func marginal(specs specsMap, valSpec *slices, baseFt, obsFt *sea.FType, fts sea.FTypes, conn *chutils.Connect) error {
	pd := &sea.PlotDef{
		Show:     specs.plotShow(),
		Title:    "",
		XTitle:   "",
		YTitle:   "",
		STitle:   "",
		Legend:   false,
		Height:   specs.plotHeight(),
		Width:    specs.plotWidth(),
		FileName: "",
	}

	var pathMarg string
	var pipe sea.Pipeline
	var e error

	graphDir, e := specs.gDir("marginal", valSpec)
	if e != nil {
		return e
	}

	for lvl := range baseFt.FP.Lvl {
		subDir := fmt.Sprintf("%s%v", valSpec.feature, lvl)

		if pathMarg, e = makeSubDir(graphDir, subDir); e != nil {
			return e
		}
		qry := fmt.Sprintf("%s AND %s=", specs.getQuery("assess"), valSpec.feature)

		switch lvl.(type) {
		case string:
			qry = fmt.Sprintf("%s '%s' ORDER BY rand32(10) LIMIT 10000", qry, lvl)
		default:
			qry = fmt.Sprintf("%s %v ORDER BY rand32(10) LIMIT 10000", qry, lvl)
		}

		if pipe, e = newPipe(qry, "marginal", specs, 0, fts, conn); e != nil {
			return e
		}

		for _, fld := range specs.allFeatures() {
			pd.Title = fmt.Sprintf("%s<br>metric %s restrict %s = %v", specs.title(), valSpec.name, valSpec.feature, lvl)
			pd.FileName = fmt.Sprintf("%s%s.html", pathMarg, fld)
			modelLoc := specs.modelDir() + "model"
			fldName := fld

			if pipe.GetFType(fld).Role == sea.FRCat {
				fldName = fld + "Oh"
			}

			if e := sea.Marginal(modelLoc, fldName, valSpec.target, pipe, pd, obsFt); e != nil {
				return e
			}
		}
	}

	return nil
}

// assess generates KS, Decile and SegPlot plots of the features in the model plus any fields specified by the key assessAddl.
//
//   - segSpec : feature to segment the output on.
//   - obsFT: sea.FType of the target field
func assess(pipe sea.Pipeline, specs specsMap, obsFt *sea.FType, segSpec *slices, log *os.File) error {
	if pipe.GetFType(segSpec.feature) == nil {
		return fmt.Errorf("feature %s not in pipeline", segSpec.feature)
	}

	pd := &sea.PlotDef{
		Show:     specs.plotShow(),
		Title:    "",
		XTitle:   "",
		YTitle:   "",
		STitle:   "",
		Legend:   false,
		Height:   specs.plotHeight(),
		Width:    specs.plotWidth(),
		FileName: "",
	}

	graphDir, e := specs.gDir("assess", segSpec)
	if e != nil {
		return e
	}

	modelLoc := specs.modelDir() + "model"

	nnP, e := sea.PredictNN(modelLoc, pipe, false)
	if e != nil {
		return e
	}

	nCat := nnP.OutputCols()

	fit, e := sea.Coalesce(nnP.FitSlice(), nCat, segSpec.target, false, false, nil)
	if e != nil {
		return e
	}

	fit = sea.UnNormalize(fit, obsFt)

	obs, e := sea.Coalesce(nnP.ObsSlice(), nCat, segSpec.target, obsFt.Role == sea.FRCat, false, nil)
	if e != nil {
		return e
	}

	obs = sea.UnNormalize(obs, obsFt)

	if e1 := pipe.GData().AppendField(sea.NewRawCast(fit, nil), "fit", sea.FRCts); e1 != nil {
		return e1
	}

	if e1 := pipe.GData().AppendField(sea.NewRawCast(obs, nil), "obs", sea.FRCts); e1 != nil {
		return e1
	}

	xy, e := sea.NewXY(fit, obs)
	if e != nil {
		return e
	}

	// overall assessment
	switch obsFt.Role {
	case sea.FRCat:
		pd.Title, pd.FileName = fmt.Sprintf("%s<br>KS-%s", specs.title(), segSpec.name), graphDir+"ksAll.html"
		ks, _, _, e1 := sea.KS(xy, pd)
		if e1 != nil {
			return e1
		}
		logger(log, fmt.Sprintf("\n\nModel Assessment\nKS - %s: %0.1f%%\n\n", segSpec.name, ks), true)

	case sea.FRCts:
		logger(log, fmt.Sprintf("\n\nModel Assessment\n R-Squared %0.1f%%\n\n", sea.R2(obs, fit)), true)
	}

	pd.Title, pd.FileName = fmt.Sprintf("%s<br>Decile-%s", specs.title(), segSpec.name), graphDir+"decileAll.html"
	e = sea.Decile(xy, pd)
	if e != nil {
		return e
	}

	baseSl, e := sea.NewSlice(segSpec.feature, minCount, pipe, nil)
	if e != nil {
		return e
	}

	// run through the values we're slicing on
	for baseSl.Iter() {
		var pathVal string
		baseSlicer := baseSl.MakeSlicer()
		segPipe, e := pipe.Slice(baseSlicer)
		if e != nil {
			return e
		}

		subDir := fmt.Sprintf("%s%v", segSpec.feature, baseSl.Value())
		if pathVal, e = makeSubDir(graphDir, subDir); e != nil {
			return e
		}

		pltFile := fmt.Sprintf("%sdecileALL.html", pathVal)
		pltTitle := fmt.Sprintf("%s<br>%s<br>restrict %s", specs.title(), segSpec.name, baseSl.Title())
		pd.YTitle, pd.XTitle, pd.STitle = "", "", ""

		// get fitted and observed values
		x := segPipe.Get("fit")
		y := segPipe.Get("obs")

		// ranges for graphs
		minVal := math.Min(x.Summary.DistrC.Q[2], y.Summary.DistrC.Q[2])
		maxVal := math.Max(x.Summary.DistrC.Q[len(x.Summary.DistrC.Q)-3], y.Summary.DistrC.Q[len(y.Summary.DistrC.Q)-3])

		// structure needed for KS and Decile plots
		xy, e = sea.NewXY(x.Data.([]float64), y.Data.([]float64))
		if e != nil {
			return e
		}

		if obsFt.Role == sea.FRCat {
			ksFile := fmt.Sprintf("%sksALL.html", pathVal)
			ksTitle := fmt.Sprintf("%s<br>%s<br>restrict %s", specs.title(), segSpec.name, baseSl.Title())
			pd.FileName, pd.Title = ksFile, ksTitle

			if _, _, _, e = sea.KS(xy, pd); e != nil {
				return e
			}
		}

		pd.FileName, pd.Title, pd.STitle = pltFile, pltTitle, ""
		if e1 := sea.Decile(xy, pd); e1 != nil {
			return e1
		}

		// run through the fields we're making SegPlots for
		for _, fld := range specs.assessFields() {
			ft := pipe.GetFType(fld)
			if ft == nil {
				return fmt.Errorf("assess: feature %s not in pipeline", fld)
			}

			pltFile = fmt.Sprintf("%sdecile%s.html", pathVal, fld)
			pltTitle = fmt.Sprintf("%s<br>Field %s Metric %s<br>Where %s", specs.title(), fld, segSpec.name, baseSl.Title())
			pd.YTitle, pd.XTitle = "", ""
			pd.FileName, pd.Title, pd.STitle = pltFile, pltTitle, ""
			if e := sea.SegPlot(segPipe, "obs", "fit", fld, pd, &minVal, &maxVal); e != nil {
				return e
			}
		}
	}

	return nil
}

// TODO: add implementation check after save

// TODO: implement trim option in pipeline
// TODO: in seafan build a set of queries to do CIs by strats
// TODO: pass1Fields need to also keep stratify fields

// TODO: remove all direct access to specs -- use methods
// TODO: unified.serv_map -- build in serv_map

// fc_type, 1yrHPI at trg, fcstMonth, trgFcCompletion

// servicer, weighting, batchsize,

// TODO: figure out why fcstMonth has to be a cat for curves to work

// TODO: think about fcstMonth curve for mods...is there a reason it misses?
