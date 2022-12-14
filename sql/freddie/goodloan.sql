// requirements for a loan to be considered for the sample
has(qa.field,'state') = 0
AND has(qa.field,'fico') = 0
// AND has(qa.field,'dti') = 0 // eliminates HARP
AND has(qa.field,'ltv') = 0
AND has(allFail,'dq') = 0
AND has(qa.field,'zip3') = 0
AND has(allFail,'curRate') = 0
AND has(allFail,'upb') = 0
AND has(qa.field,'propVal') = 0
AND has(qa.field, 'numBorr') = 0
AND aoRemTerm > 0
