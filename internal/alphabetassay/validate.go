package alphabetassay

import "fmt"

// RunDirNameShape is the response-dir name shape ParseRunDir accepts, named in
// the mismatch error so a caller sees why nothing parsed.
const RunDirNameShape = "asy-<class>-s<scenario>-<model>/out/responses"

// CheckCollectInputs guards alphabet-collect-assay against a success-shaped empty
// result. dirsFound is the number of response dirs the per-class globs matched;
// dirsParsed is how many of them ParseRunDir accepted; totalItems is the
// responses collected across all classes. It returns a named, non-nil error when:
//
//   - the globs matched response dirs but ParseRunDir matched none of them
//     (naming the dirs-found-vs-parsed counts);
//   - zero responses were collected (nothing was scored).
//
// A study with at least one collected response returns nil, so a correct collect
// run is unchanged.
func CheckCollectInputs(dirsFound, dirsParsed, totalItems int) error {
	if dirsFound > 0 && dirsParsed == 0 {
		return fmt.Errorf("alphabet-collect-assay: glob matched %d response dir(s) but ParseRunDir matched none of them (expected %q); nothing was collected", dirsFound, RunDirNameShape)
	}
	if totalItems == 0 {
		return fmt.Errorf("alphabet-collect-assay: collected 0 responses (%d response dir(s) found, %d parsed); nothing was collected", dirsFound, dirsParsed)
	}
	return nil
}
