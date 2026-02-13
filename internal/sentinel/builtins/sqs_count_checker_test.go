package builtins

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSQSClient struct {
	GetQueueAttributesFunc func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

func (m *MockSQSClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	return m.GetQueueAttributesFunc(ctx, params, optFns...)
}

func TestSQSCountCheckerSentinel_Configure(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
	}{
		{
			name: "valid configuration",
			config: map[string]any{
				"queue_url":         "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"max_message_count": 100,
				"aws_region":        "us-east-1",
			},
			expectError: false,
		},
		{
			name: "missing queue_url",
			config: map[string]any{
				"max_message_count": 100,
				"aws_region":        "us-east-1",
			},
			expectError: true,
		},
		{
			name: "missing max_message_count",
			config: map[string]any{
				"queue_url":  "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"aws_region": "us-east-1",
			},
			expectError: true,
		},
		{
			name: "invalid max_message_count type",
			config: map[string]any{
				"queue_url":         "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"max_message_count": "not a number",
				"aws_region":        "us-east-1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentinel := &SQSCountCheckerSentinel{}
			err := sentinel.Configure(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if queueURL, ok := tt.config["queue_url"]; ok {
					assert.Equal(t, queueURL.(string), sentinel.queueURL)
				}

				if maxCount, ok := tt.config["max_message_count"]; ok && !tt.expectError {
					switch v := maxCount.(type) {
					case int:
						assert.Equal(t, int64(v), sentinel.maxMessageCount)
					case float64:
						assert.Equal(t, int64(v), sentinel.maxMessageCount)
					}
				}

				if region, ok := tt.config["aws_region"]; ok {
					assert.Equal(t, region.(string), sentinel.awsRegion)
				}
			}
		})
	}
}

func TestSQSCountCheckerSentinel_Check(t *testing.T) {
	tests := []struct {
		name             string
		messageCount     string
		maxMessageCount  int64
		expectedStatus   signal.Status
		attributesResult map[string]string
		expectError      bool
	}{
		{
			name:            "healthy - under limit",
			messageCount:    "50",
			maxMessageCount: 100,
			expectedStatus:  signal.StatusHealthy,
			attributesResult: map[string]string{
				string(types.QueueAttributeNameApproximateNumberOfMessages): "50",
			},
		},
		{
			name:            "unhealthy - over limit",
			messageCount:    "150",
			maxMessageCount: 100,
			expectedStatus:  signal.StatusUnhealthy,
			attributesResult: map[string]string{
				string(types.QueueAttributeNameApproximateNumberOfMessages): "150",
			},
		},
		{
			name:            "invalid message count format",
			messageCount:    "not-a-number",
			maxMessageCount: 100,
			expectedStatus:  signal.StatusUnknown,
			attributesResult: map[string]string{
				string(types.QueueAttributeNameApproximateNumberOfMessages): "not-a-number",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSQSClient{
				GetQueueAttributesFunc: func(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
					return &sqs.GetQueueAttributesOutput{
						Attributes: tt.attributesResult,
					}, nil
				},
			}

			sentinel := &SQSCountCheckerSentinel{
				queueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				maxMessageCount: tt.maxMessageCount,
				awsRegion:       "us-east-1",
				client:          mockClient,
			}

			signal, err := sentinel.Check(context.Background(), "test-alarm")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, signal.Status)
			assert.Equal(t, "test-alarm", signal.AlarmID)
			assert.NotEmpty(t, signal.Message)
			assert.NotZero(t, signal.Timestamp)
		})
	}
}
