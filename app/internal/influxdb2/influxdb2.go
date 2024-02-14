package influxdb2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"rinha-be-2-app/internal/config"
	"rinha-be-2-app/internal/logger"
	"rinha-be-2-app/internal/metric"
	"time"

	"github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

type InfluxDB2ServiceT struct {
	Client  influxdb2.Client
	Writer  api.WriteAPIBlocking
	Querier api.QueryAPI
	Buckets api.BucketsAPI
	Deleter api.DeleteAPI
}

var InfluxDB2Service *InfluxDB2ServiceT

// NewInfluxDB2Service inicializa o singleton de acesso ao InfluxDB2
func NewInfluxDB2Service() {

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(60),
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	InfluxDB2Service = &InfluxDB2ServiceT{
		//Client: influxdb2.NewClient(getEndpoint(), config.AppConfig.InfluxDB2Token),
		Client: influxdb2.NewClientWithOptions(getEndpoint(), config.AppConfig.InfluxDB2Token, influxdb2.DefaultOptions().SetHTTPClient(httpClient)),
	}
	InfluxDB2Service.Writer = InfluxDB2Service.Client.WriteAPIBlocking(config.AppConfig.InfluxDB2Org, config.AppConfig.InfluxDB2Bucket)
	InfluxDB2Service.Buckets = InfluxDB2Service.Client.BucketsAPI()
	InfluxDB2Service.Querier = InfluxDB2Service.Client.QueryAPI(config.AppConfig.InfluxDB2Org)
	InfluxDB2Service.Deleter = InfluxDB2Service.Client.DeleteAPI()
	//go ErrorHandler()
}

func NewBucket(name string) error {
	rp := "1w"
	id := uuid.New().String()
	id = id[0:16]
	_, err := InfluxDB2Service.Buckets.CreateBucket(context.TODO(), &domain.Bucket{
		Name:  name,
		Id:    &id,
		OrgID: &config.AppConfig.InfluxDB2Org,
		Rp:    &rp,
	})
	if err != nil {
		logger.Info("Error creating bucket", "error", err)
		return err
	}
	return nil
}

// Write pode escrever um ou mais pontos de dados no InfluxDB2
func Write(org, bucket string, metrics ...metric.DataPointT) error {

	var points []*write.Point
	points = make([]*write.Point, 0, len(metrics))

	for _, metric := range metrics {
		measurement, tags, fields, ts := metric.GetDataPoint()
		point := write.NewPoint(measurement, tags, fields, ts)
		points = append(points, point)
	}

	InfluxDB2Service.Writer.WritePoint(context.TODO(), points...)
	return nil
}

func Query(stmt string) (*api.QueryTableResult, error) {

	tblResult, err := InfluxDB2Service.Querier.Query(context.TODO(), stmt)
	if err != nil {
		logger.Info("Error querying InfluxDB2", "error", err)
		return nil, err
	}
	return tblResult, nil
}

// ErrorHandler captura os erros de escrita no InfluxDB2
//func ErrorHandler() {
//
//	for {
//		select {
//		case err := <-writeAPI.Errors():
//			logger.Error("Error writing to InfluxDB2", "error", err)
//		}
//	}
//}

func getEndpoint() string {
	return fmt.Sprintf("http://%s:%s", config.AppConfig.InfluxDB2Host, config.AppConfig.InfluxDB2Port)
}
