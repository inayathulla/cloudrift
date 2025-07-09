package detector

type DriftResultPrinter interface {
	PrintDrift(results interface{}, plan, live interface{})
}
