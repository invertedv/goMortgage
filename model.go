package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	if path := specs.startFrom(); path != "" {
		path = fmt.Sprintf("%smodel", slash(path))
		nnModel, e := sea.LoadNN(path, pipe, true)
		if e != nil {
			return nil, e
		}

		sea.WithCostFn(specs.costFunc())(nnModel)
		sea.WithName(specs["model"])(nnModel)
		return nnModel, nil
	}

	modSpec, e := modelSpec(specs)
	if e != nil {
		return nil, e
	}

	return sea.NewNNModel(modSpec, pipe, true,
		sea.WithCostFn(specs.costFunc()),
		sea.WithName(specs["model"]))
}

// getFTs gets the FTypes to use for all the pipelines if we're starting from an existing model.
// If we're not, nil is returned.
func getFts(specs specsMap) (sea.FTypes, error) {
	if path := specs.startFrom(); path != "" {
		path = fmt.Sprintf("%sfieldDefs.jsn", slash(path))

		return sea.LoadFTypes(path)
	}

	return nil, nil
}

// model is the core model-building function.
func model(specs specsMap, conn *chutils.Connect, log *os.File) error {
	var (
		e                  error
		modelPipe, valPipe sea.Pipeline
		fts                sea.FTypes
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

	// get FTypes if startFrom: key is used, o.w. this is nil
	startFts, e := getFts(specs)
	if e != nil {
		return e
	}

	if modelPipe, e = newPipe(specs.getQuery("model"), "Modeling data", specs,
		batchSize, startFts, conn); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("%v", modelPipe), false)

	// add defaults and restrict fts to features defined in specs append(specs.allCat(), specs.ctsFeatures()...)
	if fts, e = addDefault(modelPipe, append(specs.ctsFeatures(), specs.allCat()...)); e != nil {
		return e
	}

	if er := fts.Save(specs.modelDir() + "fieldDefs.jsn"); er != nil {
		return er
	}

	// validation pipeline
	if valPipe, e = newPipe(specs.getQuery("validate"), "Validation data", specs, 0, fts, conn); e != nil {
		return e
	}

	logger(log, fmt.Sprintf("\n\n%v", valPipe), false)

	// assess pipeline
	//	if assessPipe, e = newPipe(specs.getQuery("assess"), "Assess data", specs, 0, fts, conn); e != nil {
	//		return e
	//	}

	//	logger(log, fmt.Sprintf("\n\n%v", assessPipe), false)

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

	// model fit struct
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

	elapsed := time.Since(start).Minutes()
	logger(log, fmt.Sprintf("model build run time: %0.1f minutes", elapsed), true)

	return nil
}
