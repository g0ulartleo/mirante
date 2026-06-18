# mirante

mirante is a monitoring system designed to watch over multiple projects and external services,
providing notifications per alarm, a simple web dashboard for real-time monitoring and a CLI for management.

## current status

this project was created initially for my own usage (and for studying golang) and it is still under development,
so production usage is not yet recommended.

## architecture

### components

- **HTTP Server:** Serves the web UI that displays alarm status and history, and an admin API for CLI usage. Located in `cmd/http-server/`.
- **Worker Server:** Processes background tasks such as writing signals and requesting sentinel checks over gRPC. See `cmd/worker-server/`.
- **Sentinel Runner:** Executes sentinels and returns check results via gRPC, isolating sentinel dependencies from Mirante core. Located in `cmd/sentinel-runner/`.
- **Scheduler:** Registers and executes periodic sentinel checks as well as cleanup tasks. Located in `cmd/scheduler/`.
- **CLI:** A command-line interface for managing alarms and signals. See `cmd/cli/`.

### builtin-sentinels

- **EndpointChecker**: Performs HTTP operations on URLs and validates responses based on configuration
- **MySQLCountChecker**: Executes SQL queries that return counts and validates them against expected values
- **SQSCountChecker**: Monitors the number of messages in an SQS queue and alerts if it exceeds a threshold
- see all built-in sentinels with configuration examples [here](docs/builtin-sentinels.md)
- sentinel runtime contract and external runner guide [here](docs/sentinel-runtime.md)


## what is working right now

- alarms management using the CLI (set-alarm, get-alarm, delete-alarm)
- notifications through email or slack
- built-in sentinels


## what is in the roadmap

- allow env variables on alarm configs
- alarm initialization/integration with github repo
- external/private sentinel runners in any language using the runtime contract
- add warning state for alarms


## setting up new alarms

   alarms are defined with code, not YAML. create an alarm runtime repo with the CLI:

   ```bash
   # scaffold a Node.js runtime repo
   mirante init repo --runtime nodejs --dir my-alarms
   cd my-alarms

   # add a new alarm
   mirante new alarm check-server-events-dlq

   # install dependencies and start
   npm install
   npm start
   ```

   alarms are loaded from `src/alarms/`. the directory structure determines the path in the dashboard.
   a file at `src/alarms/Production/DB/my-alarm.ts` gets the path `Production/DB`.

   for Go runtimes:
   ```bash
   mirante init repo --runtime go --dir my-alarms
   ```

   configuration lives in `mirante.yaml` in the runtime repo root.
   start from `examples/config/mirante.yaml` for the core server configuration.

   start with setting up authentication, then using `help` to see the available commands:
   ```bash
   # using OAuth authentication
   mirante auth <your_endpoint>

   # using API key authentication
   mirante auth-key <your_endpoint> <api_key>

   mirante help
   ```


## license

mirante is distributed under the GNU General Public License v3.
see the [LICENSE](LICENSE) file for details.
