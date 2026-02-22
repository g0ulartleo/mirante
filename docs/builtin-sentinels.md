## Built-in Sentinels

Note: Alarm `config` values support environment variable placeholders with `${VAR_NAME}`. Values are resolved when the alarm is loaded/inserted, and missing variables cause validation errors.

### Endpoint Checker

The Endpoint Checker sentinel type performs HTTP operations on URLs and validates responses based on configuration.

#### Configuration

```yaml
id: providers-apis-google-health-check
name: Google Health Check
type: endpoint-checker
config:
  url: ${HEALTHCHECK_URL}
  expected_status: 200
  expected_body: "Hello, World!" # optional
```

### MySQL Count Checker

The MySQL Count Checker sentinel type executes a SQL query that returns a count and validates it against an expected value.

#### Configuration

```yaml
id: users-count-check
name: Users Count Check
type: mysql-count-checker
config:
  connection:
    host: localhost
    port: 3306
    user: root
    password: ${MYSQL_PASSWORD}
    database: myapp
  query: "SELECT COUNT(*) FROM users"
  expected: 100
```

### SQS Count Checker

The SQS Count Checker sentinel type monitors the number of messages in an Amazon SQS queue and alerts if it exceeds a specified threshold.

#### Configuration

```yaml
id: queue-backlog-monitor
name: SQS Queue Message Count Monitor
type: sqs-count-checker
config:
  queue_url: ${SQS_QUEUE_URL}
  max_message_count: 1000
  aws_region: us-east-1
```

### Step Functions Checker

The Step Functions Checker sentinel type monitors AWS Step Functions state machine executions and validates that they are running successfully according to a specified schedule.

This sentinel supports two modes:

1. **Interval-based**: Checks if there was a successful execution within a specified time window (e.g., last 24 hours)
2. **Cron-based**: Checks if there was a successful execution since the last expected execution time based on a cron schedule

#### Configuration with Execution Interval

Use this mode when you want to ensure a successful execution happened within a fixed time window:

```yaml
id: daily-etl-job-monitor
name: Daily ETL Job Monitor
type: step-functions-checker
config:
  state_machine_arn: ${STEP_FUNCTIONS_ARN}
  execution_interval: 24h
  aws_region: us-east-1
```

#### Configuration with Cron Expression

Use this mode when you want to validate executions against a cron schedule. This is useful for step functions that run at specific times or with complex schedules:

```yaml
id: hourly-sync-job-monitor
name: Hourly Sync Job Monitor
type: step-functions-checker
config:
  state_machine_arn: ${STEP_FUNCTIONS_ARN}
  expected_cron: "0 * * * *"
  aws_region: us-east-1
```

#### Cron Expression Examples

```yaml
"0 9 * * *"           # Daily at 9 AM
"0 */6 * * *"         # Every 6 hours
"0 9 * * 1-5"         # Weekdays at 9 AM
"0 9 * * 0,6"         # Weekends (Saturday and Sunday) at 9 AM
"0 * * * 1-5"         # Every hour on weekdays
"0 9-17 * * 1-5"      # Business hours (9 AM - 5 PM) on weekdays
"0 0 1 * *"           # First day of every month at midnight
"*/30 * * * *"        # Every 30 minutes
```

**Note on Complex Schedules**: For scenarios requiring different frequencies on weekdays vs weekends (e.g., hourly on weekdays, daily on weekends), you'll need to create **two separate alarms**:
- One with `expected_cron: "0 * * * 1-5"` for weekday hourly monitoring
- Another with `expected_cron: "0 9 * * 0,6"` for weekend daily monitoring

This approach ensures proper monitoring for each schedule pattern.

#### Behavior

The sentinel will report:
- **Healthy**: If a SUCCEEDED execution is found after the last expected execution time (cron mode) or within the execution interval (interval mode)
- **Healthy**: If an execution is currently RUNNING
- **Unhealthy**: If the last successful execution is older than expected
- **Unhealthy**: If the last execution FAILED or is in any status other than SUCCEEDED or RUNNING
- **Unknown**: If there's an error fetching execution data
