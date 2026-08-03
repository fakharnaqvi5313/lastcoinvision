import { NextResponse } from "next/server";

interface ContactPayload {
  name?: unknown;
  organization?: unknown;
  email?: unknown;
  message?: unknown;
}

/**
 * Stub endpoint for the /governments pilot-program form — validates the
 * payload and logs it server-side. Not wired to a real inbox, CRM, or
 * database yet; see DECISIONS.md and the note on the governments page.
 */
export async function POST(request: Request) {
  let body: ContactPayload;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Request body must be JSON." }, { status: 400 });
  }

  const { name, organization, email, message } = body;
  if (typeof name !== "string" || !name.trim() || typeof email !== "string" || !email.trim() || typeof message !== "string" || !message.trim()) {
    return NextResponse.json({ error: "name, email, and message are required." }, { status: 400 });
  }

  console.log("[governments/contact stub]", {
    name,
    organization: typeof organization === "string" ? organization : undefined,
    email,
    message,
    receivedAt: new Date().toISOString(),
  });

  return NextResponse.json({ ok: true });
}
