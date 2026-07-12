import { healthy } from "@mirante/alarms-sdk";

export const checkServerCount = {
  id: "check-server-count",
  name: "Check Server Count",
  description: "Describe what this alarm checks.",
  howToFix: "Describe how to fix failures.",
  interval: "1m",
  notifications: {
    slackWebhooks: async () => [],
    emails: async () => [],
  },
  async run() {
    return healthy("OK");
  },
};
