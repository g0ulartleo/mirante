# mirante

mirante is a monitoring system designed to watch over multiple projects and external services,
providing notifications per alarm, a simple web dashboard for real-time monitoring and a CLI for management.

## current status

this project was created initally for my own usage (and for studying golang) and it is still under development,
so production usage is not yet recommended.

## architecture

### components

- **HTTP Server:** Serves the web UI that displays alarm status and history, and an admin API for CLI usage. Located in `cmd/http-server/`.
- **Worker Server:** Processes background tasks such as writing signals and executing sentinel checks. See `cmd/worker-server/`.
- **Scheduler:** Registers and executes periodic sentinel checks as well as cleanup tasks. Located in `cmd/scheduler/`.
- **CLI:** A command-line interface for managing alarms and signals. See `cmd/cli/`.

### builtin-sentinels

- **EndpointChecker**: Performs HTTP operations on URLs and validates responses based on configuration
- **MySQLCountChecker**: Executes SQL queries that return counts and validates them against expected values
- **SQSCountChecker**: Monitors the number of messages in an SQS queue and alerts if it exceeds a threshold
- see all built-in sentinels with configuration examples [here](docs/builtin-sentinels.md)


## what is working right now

- alarms management using the CLI (set-alarm, get-alarm, delete-alarm)
- notifications through email or slack
- built-in sentinels


## what is in the roadmap

- allow env variables on alarm configs
- alarm initialization/integration with github repo
- custom sentinel runtime, allowing private/external sentinels (make worker talk with sentinel runner using RPC)
- add warning state for alarms


## setting up new alarms

   alarms are configured via YAML files in the `config/alarms` directory.
   the directory structure reflects the URL path for an alarm's dashboard (if no `path` is defined for the alarm).

   example:
   ```yaml
   id: my-alarm
   name: My Custom Alarm
   description: "Expects a 200 status code from some API"
   type: endpoint-checker
   interval: "30s"        # or specify a cron expression in the `cron` field
   path: ['Project', 'APIs']
   config:
     url: "https://example.com"
     expected_status: 200
   notifications:
      email:
         to:
            - "test@example.com"
            - "test2@example.com"
      slack:
         webhook_url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXX"
   ```

   you can also manage alarms using the CLI.

   start with setting up authentication, then using `help` to see the available commands
   ```bash
   # using OAuth authentication
   $ mirante auth <your_endpoint>

   # using API key authentication
   $ mirante auth-key <your_endpoint> <api_key>

   $ mirante help
   ```


## license

mirante is distributed under the GNU General Public License v3.
see the [LICENSE](LICENSE) file for details.
