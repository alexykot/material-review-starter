import { eq } from "drizzle-orm";
import { auth } from "../auth.ts";
import { db, pool } from "../db/client.ts";
import { user } from "../db/schema.ts";
import { ensureBucket } from "../storage.ts";

const engineers = [
  { name: "Alice Morgan", email: "alice@example.com", password: "password123" },
  { name: "Bob Chen", email: "bob@example.com", password: "password123" },
];

async function seedEngineers() {
  for (const engineer of engineers) {
    const existing = await db.query.user.findFirst({ where: eq(user.email, engineer.email) });
    if (existing) {
      console.log(`user ${engineer.email} already exists`);
      continue;
    }
    await auth.api.signUpEmail({ body: engineer });
    console.log(`created user ${engineer.email}`);
  }
}

await ensureBucket();
console.log("bucket ready");
await seedEngineers();
await pool.end();
