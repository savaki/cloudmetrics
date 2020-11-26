package awscloudmetrics

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// NewCloudWatchClient creates a CloudWatch client
func NewCloudWatchClient() *cloudwatch.CloudWatch {
	cfg := &aws.Config{Region: aws.String(findAWSRegion(lookupAvailabilityZone))}

	return cloudwatch.New(session.New(cfg))
}

func findAWSRegion(lookupAZ func(ctx context.Context) (io.ReadCloser, error)) string {
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	if region == "" {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()

		body, err := lookupAZ(ctx)
		if err == nil {
			defer body.Close()

			data, err := ioutil.ReadAll(body)
			if err == nil {
				region = strings.TrimSpace(string(data))
				if len(region) > 0 {
					region = region[0 : len(region)-1]
				}
			}
		}
	}

	if region == "" {
		region = "us-east-1"
	}

	return region
}

func lookupAvailabilityZone(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://169.254.169.254/latest/meta-data/placement/availability-zone", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
