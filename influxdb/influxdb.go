package influx

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Client struct {
	cli influxdb2.Client
}

func New(url, authToken string) *Client {
	httpClient := &http.Client{
		Timeout: time.Second * 1000,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 60 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 30 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	opts := influxdb2.DefaultOptions().SetHTTPClient(httpClient)
	opts.SetLogLevel(1)
	client := influxdb2.NewClientWithOptions(url, authToken, opts)
	return &Client{cli: client}
}

func (cli *Client) Close() {
	cli.cli.Close()
}

func (cli *Client) Client() influxdb2.Client {
	return cli.cli
}

func (cli *Client) Set(org, bucket string) *ClientAPI {
	return &ClientAPI{
		cli:    cli.cli,
		org:    org,
		bucket: bucket,
	}
}

type ClientAPI struct {
	cli         influxdb2.Client
	ctx         context.Context
	w           api.WriteAPI
	d           api.DeleteAPI
	q           api.QueryAPI
	org         string
	bucket      string
	measurement string
}

func (api *ClientAPI) WithContext(ctx context.Context) *ClientAPI {
	api.ctx = ctx
	return api
}

func (api *ClientAPI) Measurement(measurement string) *ClientAPI {
	api.measurement = measurement
	return api
}

type ModelFace interface {
	TableName() string
}

func (api *ClientAPI) Model(dest ModelFace) *ClientAPI {
	api.measurement = dest.TableName()
	return api
}

func (api *ClientAPI) Delete(start, stop time.Time, predicate string) error {
	if api.d == nil {
		api.d = api.cli.DeleteAPI()
	}

	if api.measurement != "" {
		predicate = fmt.Sprintf("_measurement = \"%s\" AND %s", api.measurement, predicate)
	}

	return api.d.DeleteWithName(api.ctx, api.org, api.bucket, start, stop, predicate)
}

func (api *ClientAPI) Query(obj interface{}, cmd string) error {
	if api.q == nil {
		api.q = api.cli.QueryAPI(api.org)
	}

	result, err := api.q.Query(api.ctx, cmd)
	if err != nil {
		return err
	}

	if err = Scan(result, obj); err != nil {
		return err
	}

	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (api *ClientAPI) QueryFlux(obj interface{}, args ...Args) error {
	if api.q == nil {
		api.q = api.cli.QueryAPI(api.org)
	}

	var query string
	var before []Args
	before = append(before, From(api.bucket))

	for _, b := range before {
		b(&query)
	}
	for _, arg := range args {
		arg(&query)
	}
	log.Default().Printf("influxdb::query.exec: %v\n", query)
	result, err := api.q.Query(api.ctx, query)
	if err != nil {
		return err
	}

	if err = Scan(result, obj); err != nil {
		return err
	}

	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (api *ClientAPI) Add(value interface{}) (*ClientAPI, error) {
	if api.w == nil {
		api.w = api.cli.WriteAPI(api.org, api.bucket)
	}
	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() == reflect.Interface {
		reflectValue = reflectValue.Elem()
	}

	switch reflectValue.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < reflectValue.Len(); i++ {
			value := reflectValue.Index(i)
			point, err := parse(value)
			if err != nil {
				return api, err
			}
			api.w.WritePoint(point)
		}
	case reflect.Struct, reflect.Ptr:
		point, err := parse(reflectValue)
		if err != nil {
			return api, err
		}
		api.w.WritePoint(point)
	}
	return api, nil
}

// NewPointFromStruct creates new Point instance from provided struct
func parse(s reflect.Value) (*write.Point, error) {
	// t := reflect.TypeOf(s)
	// if t.Kind() == reflect.Ptr {
	// 	t = t.Elem()
	// 	s = reflect.ValueOf(s).Elem().Interface()
	// }
	// if t.Kind() != reflect.Struct {
	// 	return nil, fmt.Errorf("struct type required, got %s", t.Kind())
	// }
	t := s.Type()
	var p *write.Point = &write.Point{}
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" {
			tag = t.Field(i).Name
		}
		value := reflect.ValueOf(s).Field(i)

		tp := t.Field(i).Tag.Get("influx")
		if tp == "tag" {
			p.AddTag(tag, value.String())
			continue
		}

		if tag == "time" {
			p.SetTime(value.Interface().(time.Time))
			continue
		}

		switch value.Kind() {
		case reflect.Bool:
			p.AddField(tag, value.Bool())
		case reflect.Float32, reflect.Float64:
			p.AddField(tag, value.Float())
		case reflect.String:
			p.AddField(tag, value.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p.AddField(tag, value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			p.AddField(tag, value.Uint())
		default:
			return nil, fmt.Errorf("Unsupported type: %s", value.Kind())
		}
	}
	return p, nil
}

func (api *ClientAPI) AddPoint(time time.Time, tags map[string]string, fields map[string]interface{}) *ClientAPI {
	if api.w == nil {
		api.w = api.cli.WriteAPI(api.org, api.bucket)
	}
	p := influxdb2.NewPointWithMeasurement(api.measurement)
	for k, v := range tags {
		p = p.AddTag(k, v)
	}
	p = p.SetTime(time)

	for k, v := range fields {
		p = p.AddField(k, v)
	}
	api.w.WritePoint(p)
	return api
}

func (api *ClientAPI) Flush() {
	if api.w == nil {
		return
	}
	api.w.Flush()
}
