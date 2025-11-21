package builtins

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/g0ulartleo/mirante-alerts/internal/sentinel"
	"github.com/g0ulartleo/mirante-alerts/internal/signal"
	"github.com/robfig/cron/v3"
)

type SFNClient interface {
	ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
}

type StepFunctionsCheckerSentinel struct {
	stateMachineArn   string
	executionInterval time.Duration
	expectedCron      string
	cronSchedule      cron.Schedule
	awsRegion         string
	client            SFNClient
}

func NewStepFunctionsCheckerSentinel() sentinel.Sentinel {
	return &StepFunctionsCheckerSentinel{}
}

func (s *StepFunctionsCheckerSentinel) Configure(config map[string]any) error {
	for _, field := range []string{"state_machine_arn", "aws_region"} {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	stateMachineArn, ok := config["state_machine_arn"].(string)
	if !ok {
		return fmt.Errorf("can't convert `state_machine_arn` to string: %v", config["state_machine_arn"])
	}
	s.stateMachineArn = stateMachineArn

	if executionIntervalVal, ok := config["execution_interval"]; ok {
		intervalStr, ok := executionIntervalVal.(string)
		if !ok {
			return fmt.Errorf("can't convert `execution_interval` to string: %v", executionIntervalVal)
		}
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("can't parse `execution_interval` as duration: %v", err)
		}
		s.executionInterval = interval
	}

	if expectedCronVal, ok := config["expected_cron"]; ok {
		expectedCron, ok := expectedCronVal.(string)
		if !ok {
			return fmt.Errorf("can't convert `expected_cron` to string: %v", expectedCronVal)
		}
		s.expectedCron = expectedCron

		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(expectedCron)
		if err != nil {
			return fmt.Errorf("can't parse `expected_cron` expression: %v", err)
		}
		s.cronSchedule = schedule
	}

	if s.executionInterval == 0 && s.expectedCron == "" {
		return fmt.Errorf("either `execution_interval` or `expected_cron` must be provided")
	}

	if s.executionInterval != 0 && s.expectedCron != "" {
		return fmt.Errorf("only one of `execution_interval` or `expected_cron` can be provided")
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

	s.client = sfn.NewFromConfig(cfg)
	return nil
}

func (s *StepFunctionsCheckerSentinel) Check(ctx context.Context, alarmID string) (signal.Signal, error) {
	result, err := s.client.ListExecutions(ctx, &sfn.ListExecutionsInput{
		StateMachineArn: aws.String(s.stateMachineArn),
	})

	if err != nil {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusUnknown,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("failed to list executions: %v", err),
		}, nil
	}

	if len(result.Executions) == 0 {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusUnhealthy,
			Timestamp: time.Now(),
			Message:   "no executions found",
		}, nil
	}

	lastExecution := result.Executions[0]

	if lastExecution.Status == types.ExecutionStatusRunning {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusHealthy,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("execution '%s' is running since %s", aws.ToString(lastExecution.Name), lastExecution.StartDate.Format(time.RFC3339)),
		}, nil
	}

	if lastExecution.Status == types.ExecutionStatusSucceeded {
		if lastExecution.StopDate == nil {
			return signal.Signal{
				AlarmID:   alarmID,
				Status:    signal.StatusUnknown,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("execution '%s' succeeded but stop date is missing", aws.ToString(lastExecution.Name)),
			}, nil
		}

		stopDate := *lastExecution.StopDate
		timeSinceStop := time.Since(stopDate)

		if s.expectedCron != "" {
			return s.checkWithCron(alarmID, lastExecution, stopDate, timeSinceStop)
		}

		if timeSinceStop <= s.executionInterval {
			return signal.Signal{
				AlarmID:   alarmID,
				Status:    signal.StatusHealthy,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("execution '%s' succeeded at %s (%.1f hours ago)", aws.ToString(lastExecution.Name), stopDate.Format(time.RFC3339), timeSinceStop.Hours()),
			}, nil
		}

		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusUnhealthy,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("last successful execution '%s' was %.1f hours ago (limit: %.1f hours)", aws.ToString(lastExecution.Name), timeSinceStop.Hours(), s.executionInterval.Hours()),
		}, nil
	}

	return signal.Signal{
		AlarmID:   alarmID,
		Status:    signal.StatusUnhealthy,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("execution '%s' is in status %s", aws.ToString(lastExecution.Name), lastExecution.Status),
	}, nil
}

func (s *StepFunctionsCheckerSentinel) checkWithCron(alarmID string, lastExecution types.ExecutionListItem, stopDate time.Time, timeSinceStop time.Duration) (signal.Signal, error) {
	now := time.Now()
	lastExpectedExecution := s.getLastExpectedExecution(now)

	if stopDate.After(lastExpectedExecution) || stopDate.Equal(lastExpectedExecution) {
		return signal.Signal{
			AlarmID:   alarmID,
			Status:    signal.StatusHealthy,
			Timestamp: now,
			Message:   fmt.Sprintf("execution '%s' succeeded at %s (%.1f hours ago, expected by %s)", aws.ToString(lastExecution.Name), stopDate.Format(time.RFC3339), timeSinceStop.Hours(), lastExpectedExecution.Format(time.RFC3339)),
		}, nil
	}

	timeSinceExpected := now.Sub(lastExpectedExecution)
	return signal.Signal{
		AlarmID:   alarmID,
		Status:    signal.StatusUnhealthy,
		Timestamp: now,
		Message:   fmt.Sprintf("last successful execution '%s' at %s, but expected execution by %s (%.1f hours overdue)", aws.ToString(lastExecution.Name), stopDate.Format(time.RFC3339), lastExpectedExecution.Format(time.RFC3339), timeSinceExpected.Hours()),
	}, nil
}

func (s *StepFunctionsCheckerSentinel) getLastExpectedExecution(now time.Time) time.Time {
	prev := now.Add(-48 * time.Hour)
	var lastExpected time.Time

	for prev.Before(now) {
		next := s.cronSchedule.Next(prev)
		if next.After(now) {
			break
		}
		lastExpected = next
		prev = next
	}

	return lastExpected
}
