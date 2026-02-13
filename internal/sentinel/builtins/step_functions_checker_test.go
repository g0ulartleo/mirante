package builtins

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSFNClient struct {
	ListExecutionsFunc func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error)
}

func (m *MockSFNClient) ListExecutions(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
	return m.ListExecutionsFunc(ctx, params, optFns...)
}

func TestStepFunctionsCheckerSentinel_Configure(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration with interval",
			config: map[string]any{
				"state_machine_arn":  "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"execution_interval": "24h",
				"aws_region":         "us-east-1",
			},
			expectError: false,
		},
		{
			name: "valid configuration with cron",
			config: map[string]any{
				"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"expected_cron":     "0 9 * * *",
				"aws_region":        "us-east-1",
			},
			expectError: false,
		},
		{
			name: "missing state_machine_arn",
			config: map[string]any{
				"execution_interval": "24h",
				"aws_region":         "us-east-1",
			},
			expectError: true,
		},
		{
			name: "missing both interval and cron",
			config: map[string]any{
				"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"aws_region":        "us-east-1",
			},
			expectError: true,
			errorMsg:    "either `execution_interval` or `expected_cron` must be provided",
		},
		{
			name: "both interval and cron provided",
			config: map[string]any{
				"state_machine_arn":  "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"execution_interval": "24h",
				"expected_cron":      "0 9 * * *",
				"aws_region":         "us-east-1",
			},
			expectError: true,
			errorMsg:    "only one of `execution_interval` or `expected_cron` can be provided",
		},
		{
			name: "invalid cron expression",
			config: map[string]any{
				"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"expected_cron":     "invalid cron",
				"aws_region":        "us-east-1",
			},
			expectError: true,
		},
		{
			name: "invalid interval duration",
			config: map[string]any{
				"state_machine_arn":  "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"execution_interval": "invalid",
				"aws_region":         "us-east-1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentinel := &StepFunctionsCheckerSentinel{}
			err := sentinel.Configure(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStepFunctionsCheckerSentinel_Check_WithInterval(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		executions     []types.ExecutionListItem
		interval       time.Duration
		expectedStatus signal.Status
		expectedMsg    string
	}{
		{
			name: "healthy - recent successful execution",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusSucceeded,
					StartDate: aws.Time(now.Add(-2 * time.Hour)),
					StopDate:  aws.Time(now.Add(-1 * time.Hour)),
				},
			},
			interval:       24 * time.Hour,
			expectedStatus: signal.StatusHealthy,
			expectedMsg:    "succeeded at",
		},
		{
			name: "unhealthy - execution too old",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusSucceeded,
					StartDate: aws.Time(now.Add(-30 * time.Hour)),
					StopDate:  aws.Time(now.Add(-26 * time.Hour)),
				},
			},
			interval:       24 * time.Hour,
			expectedStatus: signal.StatusUnhealthy,
			expectedMsg:    "last successful execution",
		},
		{
			name: "healthy - execution is running",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusRunning,
					StartDate: aws.Time(now.Add(-1 * time.Hour)),
				},
			},
			interval:       24 * time.Hour,
			expectedStatus: signal.StatusHealthy,
			expectedMsg:    "is running since",
		},
		{
			name: "unhealthy - execution failed",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusFailed,
					StartDate: aws.Time(now.Add(-1 * time.Hour)),
					StopDate:  aws.Time(now.Add(-30 * time.Minute)),
				},
			},
			interval:       24 * time.Hour,
			expectedStatus: signal.StatusUnhealthy,
			expectedMsg:    "is in status",
		},
		{
			name:           "unhealthy - no executions",
			executions:     []types.ExecutionListItem{},
			interval:       24 * time.Hour,
			expectedStatus: signal.StatusUnhealthy,
			expectedMsg:    "no executions found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSFNClient{
				ListExecutionsFunc: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: tt.executions,
					}, nil
				},
			}

			sentinel := &StepFunctionsCheckerSentinel{
				stateMachineArn:   "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				executionInterval: tt.interval,
				awsRegion:         "us-east-1",
				client:            mockClient,
			}

			sig, err := sentinel.Check(context.Background(), "test-alarm")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, sig.Status)
			assert.Equal(t, "test-alarm", sig.AlarmID)
			assert.Contains(t, sig.Message, tt.expectedMsg)
			assert.NotZero(t, sig.Timestamp)
		})
	}
}

func TestStepFunctionsCheckerSentinel_Check_WithCron(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		executions     []types.ExecutionListItem
		cronExpr       string
		expectedStatus signal.Status
		expectedMsg    string
	}{
		{
			name: "healthy - execution within expected schedule",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusSucceeded,
					StartDate: aws.Time(now.Add(-10 * time.Minute)),
					StopDate:  aws.Time(now.Add(-5 * time.Minute)),
				},
			},
			cronExpr:       "0 * * * *",
			expectedStatus: signal.StatusHealthy,
			expectedMsg:    "succeeded at",
		},
		{
			name: "unhealthy - execution missed schedule",
			executions: []types.ExecutionListItem{
				{
					Name:      aws.String("execution-1"),
					Status:    types.ExecutionStatusSucceeded,
					StartDate: aws.Time(now.Add(-5 * time.Hour)),
					StopDate:  aws.Time(now.Add(-4 * time.Hour)),
				},
			},
			cronExpr:       "0 * * * *",
			expectedStatus: signal.StatusUnhealthy,
			expectedMsg:    "expected execution by",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSFNClient{
				ListExecutionsFunc: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: tt.executions,
					}, nil
				},
			}

			sentinel := NewStepFunctionsCheckerSentinel().(*StepFunctionsCheckerSentinel)
			err := sentinel.Configure(map[string]any{
				"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"expected_cron":     tt.cronExpr,
				"aws_region":        "us-east-1",
			})
			require.NoError(t, err)

			sentinel.client = mockClient

			sig, err := sentinel.Check(context.Background(), "test-alarm")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, sig.Status)
			assert.Equal(t, "test-alarm", sig.AlarmID)
			assert.Contains(t, sig.Message, tt.expectedMsg)
			assert.NotZero(t, sig.Timestamp)
		})
	}
}

func TestStepFunctionsCheckerSentinel_GetLastExpectedExecution(t *testing.T) {
	sentinel := NewStepFunctionsCheckerSentinel().(*StepFunctionsCheckerSentinel)
	err := sentinel.Configure(map[string]any{
		"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		"expected_cron":     "0 9 * * *",
		"aws_region":        "us-east-1",
	})
	require.NoError(t, err)

	now := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	lastExpected := sentinel.getLastExpectedExecution(now)

	expectedTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedTime, lastExpected)
}

func TestStepFunctionsCheckerSentinel_Check_ComplexSchedule(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		cronExpr        string
		executionOffset time.Duration
		expectedStatus  signal.Status
		description     string
	}{
		{
			name:            "weekday hourly - execution within current hour",
			cronExpr:        "0 * * * *",
			executionOffset: -15 * time.Minute,
			expectedStatus:  signal.StatusHealthy,
			description:     "Execution 15 minutes ago (should be healthy)",
		},
		{
			name:            "weekday hourly - missed execution",
			cronExpr:        "0 * * * *",
			executionOffset: -2*time.Hour - 30*time.Minute,
			expectedStatus:  signal.StatusUnhealthy,
			description:     "Last execution 2.5 hours ago (missed last 2 hourly runs)",
		},
		{
			name:            "daily at noon - execution on schedule",
			cronExpr:        "0 12 * * *",
			executionOffset: -2 * time.Hour,
			expectedStatus:  signal.StatusHealthy,
			description:     "Execution 2 hours ago, daily at noon (should be healthy if it's still same day)",
		},
		{
			name:            "every 6 hours - on schedule",
			cronExpr:        "0 */6 * * *",
			executionOffset: -30 * time.Minute,
			expectedStatus:  signal.StatusHealthy,
			description:     "Execution 30 minutes ago, runs every 6 hours (should be healthy)",
		},
		{
			name:            "every 6 hours - missed schedule",
			cronExpr:        "0 */6 * * *",
			executionOffset: -7 * time.Hour,
			expectedStatus:  signal.StatusUnhealthy,
			description:     "Execution 7 hours ago, runs every 6 hours (should be unhealthy)",
		},
		{
			name:            "every 30 minutes - recent execution",
			cronExpr:        "*/30 * * * *",
			executionOffset: -10 * time.Minute,
			expectedStatus:  signal.StatusHealthy,
			description:     "Execution 10 minutes ago, runs every 30 minutes",
		},
		{
			name:            "every 30 minutes - missed execution",
			cronExpr:        "*/30 * * * *",
			executionOffset: -45 * time.Minute,
			expectedStatus:  signal.StatusUnhealthy,
			description:     "Execution 45 minutes ago, runs every 30 minutes (missed a run)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executionTime := now.Add(tt.executionOffset)

			mockClient := &MockSFNClient{
				ListExecutionsFunc: func(ctx context.Context, params *sfn.ListExecutionsInput, optFns ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{
								Name:      aws.String("test-execution"),
								Status:    types.ExecutionStatusSucceeded,
								StartDate: aws.Time(executionTime.Add(-10 * time.Minute)),
								StopDate:  aws.Time(executionTime),
							},
						},
					}, nil
				},
			}

			sentinel := NewStepFunctionsCheckerSentinel().(*StepFunctionsCheckerSentinel)
			err := sentinel.Configure(map[string]any{
				"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				"expected_cron":     tt.cronExpr,
				"aws_region":        "us-east-1",
			})
			require.NoError(t, err)

			sentinel.client = mockClient

			ctx := context.Background()
			sig, err := sentinel.Check(ctx, "test-alarm")
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, sig.Status,
				"Test: %s\nDescription: %s\nCron: %s\nExecution time offset: %s\nMessage: %s",
				tt.name, tt.description, tt.cronExpr,
				tt.executionOffset.String(),
				sig.Message)
			assert.Equal(t, "test-alarm", sig.AlarmID)
			assert.NotZero(t, sig.Timestamp)
		})
	}
}
