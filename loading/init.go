package loading

import (
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

var (
	loadingSpinner        = spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	loadingSpinnerWrapper = &spinnerWrapper{
		Spinner: loadingSpinner,
	}
)

type spinnerWrapper struct {
	*spinner.Spinner
}

func (s *spinnerWrapper) UpdateSuffix(suffix string) {
	if strings.HasSuffix(suffix, "...") {
		suffix += "..."
	}
	s.Spinner.Suffix = " " + suffix
	s.Spinner.Restart()
}

func Loading(text string) *spinnerWrapper {
	loadingSpinner.Suffix = " " + text
	loadingSpinner.Restart()
	return loadingSpinnerWrapper
}

func Spinner() *spinner.Spinner {
	return loadingSpinner
}
