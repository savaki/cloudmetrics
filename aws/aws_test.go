package awscloudmetrics

//	Copyright 2016 Matt Ho
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAWSRegion(t *testing.T) {
	lookupFunc := func(context.Context) (io.ReadCloser, error) {
		return nil, errors.New("failure")
	}

	t.Run("OK - Use AWS_DEFAULT_REGION", func(t *testing.T) {
		env := "AWS_DEFAULT_REGION"
		val := "us-west-1"
		os.Setenv(env, val)
		assert.Equal(t, val, findAWSRegion(lookupFunc))
		os.Setenv(env, "")
	})

	t.Run("OK - Use AWS_REGION", func(t *testing.T) {
		env := "AWS_REGION"
		val := "us-west-2"
		os.Setenv(env, val)
		assert.Equal(t, val, findAWSRegion(lookupFunc))
		os.Setenv(env, "")
	})

	t.Run("OK - Use lookupAZ function", func(t *testing.T) {
		lookupFunc := func(context.Context) (io.ReadCloser, error) {
			return ioutil.NopCloser(strings.NewReader("eu-west-1b")), nil
		}
		assert.Equal(t, "eu-west-1", findAWSRegion(lookupFunc))
	})

	t.Run("OK - Default", func(t *testing.T) {
		assert.Equal(t, "us-east-1", findAWSRegion(lookupFunc))
	})

}
