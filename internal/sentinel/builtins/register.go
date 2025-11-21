package builtins

import "github.com/g0ulartleo/mirante-alerts/internal/sentinel"

func Register(f *sentinel.SentinelFactory) {
	f.Register("endpoint-checker", NewEndpointCheckerSentinel)
	f.Register("mysql-count-checker", NewMySQLCountCheckerSentinel)
	f.Register("sqs-count-checker", NewSQSCountCheckerSentinel)
	f.Register("postgres-count-checker", NewPostgresCountCheckerSentinel)
	f.Register("step-functions-checker", NewStepFunctionsCheckerSentinel)
}
