import Anthropic from "@anthropic-ai/sdk";
import { readFileSync, writeFileSync } from "node:fs";
import { env, rootEnvPath } from "../env.ts";

const client = new Anthropic();

const agentConfig = {
  name: "Document Reviewer",
  model: "claude-opus-5-5",
  system: "You are a helpful assistant.",
  tools: [
    {
      type: "agent_toolset_20260401" as const,
      default_config: { enabled: true, permission_policy: { type: "always_allow" as const } },
      configs: [
        { name: "web_fetch" as const, enabled: false },
        { name: "web_search" as const, enabled: false },
      ],
    },
  ],
};

const environmentConfig = {
  name: "document-review",
  config: { type: "cloud" as const, networking: { type: "limited" as const } },
};

async function ensureEnvironment(): Promise<string> {
  if (env.environmentId) {
    const existing = await client.beta.environments.retrieve(env.environmentId);
    console.log(`environment ${existing.id} already exists`);
    return existing.id;
  }
  const created = await client.beta.environments.create(environmentConfig);
  console.log(`created environment ${created.id}`);
  return created.id;
}

async function ensureAgent(): Promise<string> {
  if (env.agentId) {
    const updated = await client.beta.agents.update(env.agentId, agentConfig);
    console.log(`agent ${updated.id} is at version ${updated.version}`);
    return updated.id;
  }
  const created = await client.beta.agents.create(agentConfig);
  console.log(`created agent ${created.id} (version ${created.version})`);
  return created.id;
}

function writeEnvValues(values: Record<string, string>) {
  const lines = readFileSync(rootEnvPath, "utf8").split("\n");
  for (const [name, value] of Object.entries(values)) {
    const index = lines.findIndex((line) => line.startsWith(`${name}=`));
    if (index >= 0) lines[index] = `${name}=${value}`;
    else lines.splice(lines.at(-1) === "" ? lines.length - 1 : lines.length, 0, `${name}=${value}`);
  }
  writeFileSync(rootEnvPath, lines.join("\n"));
}

const environmentId = await ensureEnvironment();
const agentId = await ensureAgent();
writeEnvValues({ ENVIRONMENT_ID: environmentId, AGENT_ID: agentId });
console.log("wrote AGENT_ID and ENVIRONMENT_ID to .env");
