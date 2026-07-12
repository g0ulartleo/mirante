import { serveRuntime } from "@mirante/alarms";

await serveRuntime({
  alarmsDir: new URL("./alarms", import.meta.url).pathname,
  addr: process.env.ALARM_RUNTIME_ADDR ?? "127.0.0.1:50051",
});
