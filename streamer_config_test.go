// Copyright 2021 OTA Insight Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bqwriter

import (
	"testing"
	"time"
)

func deepCloneStreamerConfig(cfg *StreamerConfig) *StreamerConfig {
	if cfg == nil {
		return nil
	}
	copy := new(StreamerConfig)
	*copy = *cfg
	if cfg.InsertAllClient != nil {
		copy.InsertAllClient = new(InsertAllClientConfig)
		*(copy.InsertAllClient) = *(cfg.InsertAllClient)
	}
	if cfg.StorageClient != nil {
		copy.StorageClient = new(StorageClientConfig)
		*(copy.StorageClient) = *(cfg.StorageClient)
	}
	return copy
}

var (
	expectedDefaultStreamerConfig = StreamerConfig{
		WorkerCount:     DefaultWorkerCount,
		WorkerQueueSize: DefaultWorkerQueueSize,
		MaxBatchDelay:   DefaultMaxBatchDelay,
		Logger:          stdLogger{},
		InsertAllClient: &InsertAllClientConfig{
			BatchSize:              DefaultBatchSize,
			MaxRetryDeadlineOffset: DefaultMaxRetryDeadlineOffset,
		},
	}

	expecredDefaultStorageClient = StorageClientConfig{
		MaxRetries:             DefaultMaxRetries,
		InitialRetryDelay:      DefaultInitialRetryDelay,
		MaxRetryDeadlineOffset: DefaultMaxRetryDeadlineOffset,
		RetryDelayMultiplier:   DefaultRetryDelayMultiplier,
	}
)

func assertStreamerConfig(t *testing.T, inputCfg *StreamerConfig, expectedOuputCfg *StreamerConfig) {
	// deep clone input param so that we can really test if we do not mutate our variables
	inputPassedCfg := deepCloneStreamerConfig(inputCfg)
	if inputCfg == nil {
		assertNil(t, inputPassedCfg)
	} else {
		assertNotEqualShallow(t, inputPassedCfg, inputCfg)
		// insertAll client is either nil in output or it should be a different pointer
		if inputCfg.InsertAllClient == nil {
			assertNil(t, inputCfg.InsertAllClient)
		} else {
			assertNotEqualShallow(t, inputPassedCfg.InsertAllClient, inputCfg.InsertAllClient)
		}
		// storage client is either nil in output or it should be a different pointer
		if inputCfg.StorageClient == nil {
			assertNil(t, inputCfg.StorageClient)
		} else {
			assertNotEqualShallow(t, inputPassedCfg.StorageClient, inputCfg.StorageClient)
		}
	}

	// sanitize the cfg
	outputCfg := sanitizeStreamerConfig(inputPassedCfg)

	// ensure a new pointer is returned
	assertNotEqualShallow(t, inputPassedCfg, outputCfg)
	if inputPassedCfg != nil {
		// also ensure our child cfgs are either nil or other pointers
		assertNotEqualShallow(t, inputPassedCfg.InsertAllClient, outputCfg.InsertAllClient)
		// storage client is either nil in output or it should be a different pointer
		if inputPassedCfg.StorageClient == nil {
			assertNil(t, outputCfg.StorageClient)
		} else {
			assertNotEqualShallow(t, inputPassedCfg.StorageClient, outputCfg.StorageClient)
		}
	}

	// ensure that our output is as expected
	assertEqual(t, outputCfg.StorageClient, expectedOuputCfg.StorageClient)
	// overwrite so our global deep equal check also passes
	outputCfg.StorageClient = expectedOuputCfg.StorageClient
	assertEqual(t, outputCfg.InsertAllClient, expectedOuputCfg.InsertAllClient)
	// overwrite so our global deep equal check also passes
	outputCfg.InsertAllClient = expectedOuputCfg.InsertAllClient
	// final global deep equal
	assertEqual(t, outputCfg, expectedOuputCfg)
}

func TestSanitizeStreamerConfigFullDefault(t *testing.T) {
	testCases := []struct {
		InputCfg          *StreamerConfig
		ExpectedOutputCfg *StreamerConfig
	}{
		{nil, &expectedDefaultStreamerConfig},
		{new(StreamerConfig), &expectedDefaultStreamerConfig},
		{
			&StreamerConfig{
				InsertAllClient: new(InsertAllClientConfig),
			},
			&expectedDefaultStreamerConfig,
		},
		// test defaults of StreamerCfg
	}
	for _, testCase := range testCases {
		assertStreamerConfig(t, testCase.InputCfg, testCase.ExpectedOutputCfg)
	}
}

func TestSanitizeStreamerConfigSharedDefaults(t *testing.T) {
	testCases := []struct {
		InputWorkerCount    int
		ExpectedWorkerCount int

		InputWorkerQueueSize    int
		ExpectedWorkerQueueSize int

		InputMaxBatchDelay    time.Duration
		ExpectedMaxBatchDelay time.Duration

		InputLogger    Logger
		ExpectedLogger Logger
	}{
		// full default
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		// worker count cases
		{
			InputWorkerCount: -1,
			// need at least 1 worker, otherwise how can any work be done?
			ExpectedWorkerCount:     1,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			InputWorkerCount:        0,
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			InputWorkerCount:        42,
			ExpectedWorkerCount:     42,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		// worker queue size cases
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			InputWorkerQueueSize:    -1,
			ExpectedWorkerQueueSize: 0,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			InputWorkerQueueSize:    0,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			InputWorkerQueueSize:    42,
			ExpectedWorkerQueueSize: 42,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		// max batch delay cases
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			InputMaxBatchDelay:      0,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			InputMaxBatchDelay:      1,
			ExpectedMaxBatchDelay:   1,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		// loger cases
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			InputLogger:             nil,
			ExpectedLogger:          expectedDefaultStreamerConfig.Logger,
		},
		{
			ExpectedWorkerCount:     expectedDefaultStreamerConfig.WorkerCount,
			ExpectedWorkerQueueSize: expectedDefaultStreamerConfig.WorkerQueueSize,
			ExpectedMaxBatchDelay:   expectedDefaultStreamerConfig.MaxBatchDelay,
			InputLogger:             testLogger{},
			ExpectedLogger:          testLogger{},
		},
	}
	for _, testCase := range testCases {
		// we create our input & expected out cfg,
		// which we can do by starting from the default cfg and simply
		// assigning our fresh variables
		// ... input
		inputCfg := new(StreamerConfig)
		inputCfg.WorkerCount = testCase.InputWorkerCount
		inputCfg.WorkerQueueSize = testCase.InputWorkerQueueSize
		inputCfg.MaxBatchDelay = testCase.InputMaxBatchDelay
		inputCfg.Logger = testCase.InputLogger
		// ... expected output
		expectedOutputCfg := deepCloneStreamerConfig(&expectedDefaultStreamerConfig)
		expectedOutputCfg.WorkerCount = testCase.ExpectedWorkerCount
		expectedOutputCfg.WorkerQueueSize = testCase.ExpectedWorkerQueueSize
		expectedOutputCfg.MaxBatchDelay = testCase.ExpectedMaxBatchDelay
		expectedOutputCfg.Logger = testCase.ExpectedLogger
		// and finally piggy-back on our other logic
		assertStreamerConfig(t, inputCfg, expectedOutputCfg)
	}
}

func TestSanitizeStreamerConfigInsertAllDefaults(t *testing.T) {
	testCases := []struct {
		InputFailOnInvalidRows    bool
		ExpectedFailOnInvalidRows bool

		InputFailForUnknownValues    bool
		ExpectedFailForUnknownValues bool

		InputBatchSize    int
		ExpectedBatchSize int

		InputMaxRetryDeadlineOffset    time.Duration
		ExpectedMaxRetryDeadlineOffset time.Duration
	}{
		// full default
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		// fail on invalid rows cases
		{
			InputFailOnInvalidRows:         false,
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		{
			InputFailOnInvalidRows:         true,
			ExpectedFailOnInvalidRows:      true,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		// fail on unknown values cases
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			InputFailForUnknownValues:      false,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			InputFailForUnknownValues:      true,
			ExpectedFailForUnknownValues:   true,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		// batch size cases
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			InputBatchSize:                 -1,
			ExpectedBatchSize:              1,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			InputBatchSize:                 0,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			InputBatchSize:                 42,
			ExpectedBatchSize:              42,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		// max retry deadline offset cases
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			InputMaxRetryDeadlineOffset:    0,
			ExpectedMaxRetryDeadlineOffset: expectedDefaultStreamerConfig.InsertAllClient.MaxRetryDeadlineOffset,
		},
		{
			ExpectedFailOnInvalidRows:      expectedDefaultStreamerConfig.InsertAllClient.FailOnInvalidRows,
			ExpectedFailForUnknownValues:   expectedDefaultStreamerConfig.InsertAllClient.FailForUnknownValues,
			ExpectedBatchSize:              expectedDefaultStreamerConfig.InsertAllClient.BatchSize,
			InputMaxRetryDeadlineOffset:    1,
			ExpectedMaxRetryDeadlineOffset: 1,
		},
	}
	for _, testCase := range testCases {
		// we create our input & expected out cfg,
		// which we can do by starting from the default cfg and simply
		// assigning our fresh variables
		// ... input
		inputCfg := &StreamerConfig{
			InsertAllClient: new(InsertAllClientConfig),
		}
		inputCfg.InsertAllClient.FailOnInvalidRows = testCase.InputFailOnInvalidRows
		inputCfg.InsertAllClient.FailForUnknownValues = testCase.InputFailForUnknownValues
		inputCfg.InsertAllClient.BatchSize = testCase.InputBatchSize
		inputCfg.InsertAllClient.MaxRetryDeadlineOffset = testCase.InputMaxRetryDeadlineOffset
		// ... expected output
		expectedOutputCfg := deepCloneStreamerConfig(&expectedDefaultStreamerConfig)
		expectedOutputCfg.InsertAllClient.FailOnInvalidRows = testCase.ExpectedFailOnInvalidRows
		expectedOutputCfg.InsertAllClient.FailForUnknownValues = testCase.ExpectedFailForUnknownValues
		expectedOutputCfg.InsertAllClient.BatchSize = testCase.ExpectedBatchSize
		expectedOutputCfg.InsertAllClient.MaxRetryDeadlineOffset = testCase.ExpectedMaxRetryDeadlineOffset
		// ensure to configure out streamer config correctly,
		// for the worker queue size, in case the batch size is defined
		inputCfg.WorkerQueueSize = (testCase.ExpectedBatchSize + 1) / 2
		expectedOutputCfg.WorkerQueueSize = inputCfg.WorkerQueueSize
		// and finally piggy-back on our other logic
		assertStreamerConfig(t, inputCfg, expectedOutputCfg)
	}
}

func TestSanitizeStreamerConfigStorageDefaults(t *testing.T) {
	testCases := []struct {
		InputMaxRetries    int
		ExpectedMaxRetries int

		InputInitialRetryDelay    time.Duration
		ExpectedInitialRetryDelay time.Duration

		InputMaxRetryDeadlineOffset    time.Duration
		ExpectedMaxRetryDeadlineOffset time.Duration

		InputRetryDelayMultiplier    float64
		ExpectedRetryDelayMultiplier float64
	}{
		// full default
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		// max retry cases
		{
			InputMaxRetries:                -1,
			ExpectedMaxRetries:             0,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			InputMaxRetries:                0,
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			InputMaxRetries:                42,
			ExpectedMaxRetries:             42,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		// initial retry delay cases
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			InputInitialRetryDelay:         0,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			InputInitialRetryDelay:         42,
			ExpectedInitialRetryDelay:      42,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		// max retry deadline offset cases
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			InputMaxRetryDeadlineOffset:    0,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			InputMaxRetryDeadlineOffset:    2021,
			ExpectedMaxRetryDeadlineOffset: 2021,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		// retry delay multiplier cases
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      -1,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      -0.5,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      0,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      0.5,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      1,
			ExpectedRetryDelayMultiplier:   expecredDefaultStorageClient.RetryDelayMultiplier,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      1.05,
			ExpectedRetryDelayMultiplier:   1.05,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      1.5,
			ExpectedRetryDelayMultiplier:   1.5,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      5,
			ExpectedRetryDelayMultiplier:   5,
		},
		{
			ExpectedMaxRetries:             expecredDefaultStorageClient.MaxRetries,
			ExpectedInitialRetryDelay:      expecredDefaultStorageClient.InitialRetryDelay,
			ExpectedMaxRetryDeadlineOffset: expecredDefaultStorageClient.MaxRetryDeadlineOffset,
			InputRetryDelayMultiplier:      42,
			ExpectedRetryDelayMultiplier:   42,
		},
	}
	for _, testCase := range testCases {
		// we create our input & expected out cfg,
		// which we can do by starting from the default cfg and simply
		// assigning our fresh variables
		// ... input
		inputCfg := new(StreamerConfig)
		inputCfg.StorageClient = &StorageClientConfig{
			MaxRetries:             testCase.InputMaxRetries,
			InitialRetryDelay:      testCase.InputInitialRetryDelay,
			MaxRetryDeadlineOffset: testCase.InputMaxRetryDeadlineOffset,
			RetryDelayMultiplier:   testCase.InputRetryDelayMultiplier,
		}
		// ... expected output
		expectedOutputCfg := deepCloneStreamerConfig(&expectedDefaultStreamerConfig)
		expectedOutputCfg.StorageClient = &StorageClientConfig{
			MaxRetries:             testCase.ExpectedMaxRetries,
			InitialRetryDelay:      testCase.ExpectedInitialRetryDelay,
			MaxRetryDeadlineOffset: testCase.ExpectedMaxRetryDeadlineOffset,
			RetryDelayMultiplier:   testCase.ExpectedRetryDelayMultiplier,
		}
		// and finally piggy-back on our other logic
		assertStreamerConfig(t, inputCfg, expectedOutputCfg)
	}
}

type testLogger struct{}

func (testLogger) Debug(_args ...interface{})                    {}
func (testLogger) Debugf(_template string, _args ...interface{}) {}
func (testLogger) Error(_args ...interface{})                    {}
func (testLogger) Errorf(_template string, _args ...interface{}) {}
