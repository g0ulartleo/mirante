package builtins

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/g0ulartleo/mirante/internal/sentinel"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type SQSClient interface {
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

type SQSCountCheckerSentinel struct {
	queueURL        string
	maxMessageCount int64
	awsRegion       string
	client          SQSClient
}

func NewSQSCountCheckerSentinel() sentinel.Sentinel {
	return &SQSCountCheckerSentinel{}
}

func (s *SQSCountCheckerSentinel) Configure(config map[string]any) error {
	for _, field := range []string{"queue_url", "max_message_count", "aws_region"} {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	queueURL, ok := config["queue_url"].(string)
	if !ok {
		return fmt.Errorf("can't convert `queue_url` to string: %v", config["queue_url"])
	}
	s.queueURL = queueURL

	switch v := config["max_message_count"].(type) {
	case int:
		s.maxMessageCount = int64(v)
	case float64:
		s.maxMessageCount = int64(v)
	default:
		return fmt.Errorf("can't convert `max_message_count` to int64: %v", v)
	}

	awsRegion, ok := config["aws_region"].(string)
	if !ok {
		return fmt.Errorf("can't convert `aws_region` to string: %v", config["aws_region"])
	}
	s.awsRegion = awsRegion

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(s.awsRegion))
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %v", err)
	}

	s.client = sqs.NewFromConfig(cfg)
	return nil
}

func (s *SQSCountCheckerSentinel) Check(ctx context.Context, alarmID string) (signal.Signal, error) {
	result, err := s.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	})

	if err != nil {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("failed to get queue attributes: %v", err),
		}, nil
	}

	var messageCount int64
	if countStr, ok := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]; ok {
		var parseErr error
		messageCount, parseErr = strconv.ParseInt(countStr, 10, 64)
		if parseErr != nil {
			return signal.Signal{
				AlarmID:   alarmID,
				Status:    signal.StatusUnknown,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("failed to parse message count: %v", parseErr),
			}, nil
		}
	} else {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   "ApproximateNumberOfMessages attribute not found in response",
		}, nil
	}

	if messageCount <= s.maxMessageCount {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusHealthy,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("queue has %d messages, which is within the limit of %d", messageCount, s.maxMessageCount),
		}, nil
	}

	return signal.Signal{
		AlarmID:   alarmID,
		Status:    signal.StatusUnhealthy,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("queue has %d messages, which exceeds the limit of %d", messageCount, s.maxMessageCount),
	}, nil
}
