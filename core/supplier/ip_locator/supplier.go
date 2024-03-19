package ip_locator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Supplier struct{}

type IpLocation struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`

	Country     string  `json:"country,omitempty"`
	CountryCode string  `json:"countryCode,omitempty"`
	Region      string  `json:"region,omitempty"`
	RegionName  string  `json:"regionName,omitempty"`
	City        string  `json:"city,omitempty"`
	Zip         string  `json:"zip,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lon         float64 `json:"lon,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
	Isp         string  `json:"isp,omitempty"`
	Org         string  `json:"org,omitempty"`
	As          string  `json:"as,omitempty"`
	Query       string  `json:"query,omitempty"`
}

func (r *Supplier) GetLocation(ctx context.Context, ip string) (IpLocation, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("http://ip-api.com/json/%s", ip),
		nil,
	)
	if err != nil {
		return IpLocation{}, errors.WithStack(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return IpLocation{}, errors.WithStack(err)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return IpLocation{}, errors.WithStack(err)
	}

	var location IpLocation

	err = json.Unmarshal(bodyBytes, &location)
	if err != nil {
		return IpLocation{}, errors.WithStack(err)
	}

	if location.Status == "fail" {
		return location, errors.Errorf("request failed with error: %s", location.Message)
	}

	return location, nil
}

func (r *Supplier) Ping(ctx context.Context) error {
	_, err := r.GetLocation(ctx, "misis.ru")
	if err != nil {
		return errors.Wrap(err, "failed to call locator")
	}

	return nil
}
