# mirante

mirante is a monitoring system designed to watch over multiple projects and external services,
providing notifications per alarm, a simple web dashboard for real-time monitoring and a CLI for management.

## current status

this project was created initially for my own usage (and for studying golang) and it is still under development,
so production usage is not yet recommended.

## architecture

### components

- **Core API:** Serves mirante core API and a simple web interface that displays alarm signals and history.
- **Worker:** Processes background tasks such as writing signals and requesting sentinel checks over gRPC.
- **Scheduler:** Registers and executes periodic alarm checks as well as cleanup tasks.
- **Alarm Runtime:** A client managed server that lists and executes alarms via gRPC, isolating alarm dependencies and secrets.
- **CLI:** A command-line interface for managing alarms. Type `mirante help` to get started.
- **TUI:** A terminal based dashboard for managing alarms. Accessible via `mirante tui`

## what is working right now

- alarms management using the your our repo
- any gRPC supported language can serve the alarm runtime, with Node.js and Go having CLI scaffold and SDK support.
- notifications through email or slack
- cli and TUI for alarm management

## what is in the roadmap

-

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
