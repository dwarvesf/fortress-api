package googlesheet

var ()

type IService interface {
	FetchSheetContent(fromIdx int) ([]DeliveryMetricRawData, error)
}
