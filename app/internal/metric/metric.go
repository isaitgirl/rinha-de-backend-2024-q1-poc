package metric

import "time"

type TagT map[string]string        // TagT define um tipo para as tags das métricas da aplicação
type FieldT map[string]interface{} // FieldT define um tipo para os campos das métricas da aplicação

// DataPointT define uma interface para que as métricas da aplicação sejam um datapoint do InfluxDB2
type DataPointT interface {
	GetDataPoint() (string, TagT, FieldT, time.Time)
}

//type EventOrMetricT struct {
//	Timestamp   time.Time
//	Type        string
//	Description string
//	Duration    time.Duration
//	Profile     string // Equivalente ao código da assessoria (JSADV/MDQ)
//	Value       int
//}

//func (e AppEventMetricT) GetDataPoint() (MetricTagT, MetricFieldT, time.Time) {
//	t := make(MetricTagT, 0)
//	t["type"] = e.Type
//	t["description"] = e.Description
//	t["profile"] = e.Profile
//
//	f := make(MetricFieldT, 0)
//	f["value"] = e.Value
//	f["duration"] = e.Duration.Nanoseconds()
//
//	return t, f, e.Timestamp
//}
