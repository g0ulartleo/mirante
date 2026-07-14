import { healthy, type AlarmDefinition } from "@mirante/alarms";

export const checkServerCount: AlarmDefinition = {
  id: "check-server-count",
  name: "Check Server Count",
  description: "Describe what this alarm checks.",
  howToFix: "Describe how to fix failures.",
  interval: "1m",
  notifications: {
    critical: {},
    warnings: {},
  },
  async run() {
    return healthy("OK");
  },
};
