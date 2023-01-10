package wise

type IWiseService interface {
	Convert(amount float64, source, target string) (convertedAmount float64, rate float64, error error)
	GetRate(source, target string) (rate float64, err error)
}
