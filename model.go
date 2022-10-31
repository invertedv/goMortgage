package main

//
// There are 3 datasets used in the building and evaluation of a model:
//
//    1. Model Build.  The data used to fit the parameters.
//    2. Validation.  The data used to assess the model during the build.  This data selects the model to use and
//         governs early stopping.
//    3. Assessment.  The data used to build graphical assessments of the final model.
import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	grob "github.com/MetalBlueberry/go-plotly/graph_objects"

	"github.com/invertedv/chutils"

	sea "github.com/invertedv/seafan"
)

// minCount is the minimum # of rows a slice must have to make the assessment graphs
const minCount = 0

// modelSpec creates the NNModel model specification from the inputs.
func modelSpec(specs specsMap) (modSpec sea.ModSpec, err error) {
	catsOh := make([]string, 0)
	for _, cat := range specs.ohFeatures() {
		field := cat + "Oh"
		catsOh = append(catsOh, field)
	}

	target := specs.target()
	if specs.targetType() == sea.FRCat {
		target += "Oh"
	}

	embsOh, e := specs.embFeatures(true)
	if e != nil {
		return nil, e
	}

	fields := append(append(specs.ctsFeatures(), catsOh...), embsOh...)
	inputs := fmt.Sprintf("input(%s)", strings.Join(fields, "+"))
	modSpec = []string{inputs}
	modSpec = append(append(modSpec, specs.layers()...), fmt.Sprintf("Target(%s)", target))

	return
}

// getModel either creates or loads the model to fit
func getModel(specs specsMap, pipe sea.Pipeline) (*sea.NNModel, error) {
	// path will be the path to a model whose values we'll use as starting values
	path, ok := specs["startFrom"]

	switch ok {
	case true:
		path = fmt.Sprintf("%smodel", slash(path))
		nnModel, e := sea.LoadNN(path, pipe, true)
		if e != nil {
			return nil, e
		}
		sea.WithCostFn(specs.costFunc())(nnModel)
		sea.WithName(specs["model"])(nnModel)
		return nnModel, nil

	case false:
		modSpec, e := modelSpec(specs)
		if e != nil {
			return nil, e
		}

		return sea.NewNNModel(modSpec, pipe, true,
			sea.WithCostFn(specs.costFunc()),
			sea.WithName(specs["model"]))
	}

	return nil, fmt.Errorf("getModel unknown error")
}

// getFTs gets the FTypes to use for all the pipelines if we're starting from an existing model.
// If we're not, nil is returned.
func getFts(specs specsMap) sea.FTypes {
	path, ok := specs["startFrom"]
	if !ok {
		return nil
	}
	path = fmt.Sprintf("%sfieldDefs.jsn", slash(path))
	fts, _ := sea.LoadFTypes(path)

	return fts
}

// model is the core model-building function.
func model(specs specsMap, conn *chutils.Connect, log *os.File) error {
	var (
		e                              error
		modelPipe, valPipe, assessPipe sea.Pipeline
		fts                            sea.FTypes
	)

	start := time.Now()
	logger(log, fmt.Sprintf("starting model build @ %s", start.Format(time.UnixDate)), true)

	batchSize, e := specs.batchSize()
	if e != nil {
		return e
	}

	epochs, e := specs.epochs()
	if e != nil {
		return e
	}

	earlyStopping, e := specs.earlyStopping()
	if e != nil {
		return e
	}

	if modelPipe, e = newPipe(specs.getQuery("model"), "Modeling data", specs,
		batchSize, getFts(specs), conn); e != nil {
		return e
	}
	obsFt := modelPipe.GetFType(specs.target()) // save this bc we may need this to un-normalize the target field

	logger(log, fmt.Sprintf("%v", modelPipe), false)

	// add defaults and restrict fts to features defined in specs append(specs.allCat(), specs.ctsFeatures()...)
	if fts, e = addDefault(modelPipe, append(specs.ctsFeatures(), specs.allCat()...)); e != nil {
		return e
	}

	if er := fts.Save(specs["modelDir"] + "fieldDefs.jsn"); er != nil {
		return er
	}

	// validation pipeline
	if valPipe, e = newPipe(specs.getQuery("validate"), "Validation data", specs, 0, fts, conn); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\n%v", valPipe), false)

	// assess pipeline
	if assessPipe, e = newPipe(specs.getQuery("assess"), "Assess data", specs, 0, fts, conn); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\n%v", assessPipe), false)

	// load model
	nnModel, e := getModel(specs, modelPipe)
	if e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\n%v", nnModel), true)

	startLR, endLR, e := specs.learnRate()
	if e != nil {
		return e
	}

	// fit model
	fit := sea.NewFit(nnModel, epochs, modelPipe,
		sea.WithValidation(valPipe, earlyStopping),
		sea.WithLearnRate(startLR, endLR),
		sea.WithOutFile(specs.modelRoot()))

	// see if there is L2 regularization
	l2, e := specs.l2()
	if e != nil {
		return e
	}
	if l2 > 0 {
		sea.WithL2Reg(l2)(fit)
	}

	sea.Verbose = true
	if e := fit.Do(); e != nil {
		return e
	}

	sea.Verbose = false

	if e := plotCosts(fit, nnModel.Cost().Name(), specs); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\nBest Epoch: %d", fit.BestEpoch()), true)

	// plot curves
	for _, curve := range specs.slicer("curves") {
		sl := curve
		if e := curves(assessPipe, specs, obsFt, &sl); e != nil {
			return e
		}
	}

	// marginal and segPlot plots
	for _, slice := range specs.slicer("assess") {
		sl := slice // bad to pass for var as a pointer

		baseFt := modelPipe.GetFType(slice.feature) // this may not be in fts
		if e := marginal(specs, &sl, baseFt, obsFt, fts, conn); e != nil {
			return e
		}
		if e := assess(assessPipe, specs, obsFt, &sl, log); e != nil {
			return e
		}
	}

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("model build run time: %0.1f minutes", elapsed), true)

	// save assess data & model values back to ClickHouse
	if e := export(assessPipe, specs, obsFt, conn); e != nil {
		return e
	}

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

	modelLoc := specs["modelDir"] + "model"

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
			modelLoc := specs["modelDir"] + "model"
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

// assess generates KS and decile plots of the features in the model plus any fields specified by the key assessAddl.
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

	modelLoc := specs["modelDir"] + "model"

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

		x := segPipe.Get("fit")
		y := segPipe.Get("obs")
		minVal := math.Min(x.Summary.DistrC.Q[2], y.Summary.DistrC.Q[2])
		maxVal := math.Max(x.Summary.DistrC.Q[len(x.Summary.DistrC.Q)-3], y.Summary.DistrC.Q[len(y.Summary.DistrC.Q)-3])
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
