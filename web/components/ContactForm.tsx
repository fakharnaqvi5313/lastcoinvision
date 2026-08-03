"use client";

import { useState, type FormEvent } from "react";

type Status = "idle" | "submitting" | "success" | "error";

export function ContactForm() {
  const [status, setStatus] = useState<Status>("idle");
  const [errorMessage, setErrorMessage] = useState<string>();

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("submitting");
    setErrorMessage(undefined);

    const form = event.currentTarget;
    const data = new FormData(form);
    const payload = {
      name: data.get("name"),
      organization: data.get("organization"),
      email: data.get("email"),
      message: data.get("message"),
    };

    try {
      const res = await fetch("/api/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as { error?: string };
        throw new Error(body.error ?? `Request failed (${res.status}).`);
      }
      setStatus("success");
      form.reset();
    } catch (err) {
      setStatus("error");
      setErrorMessage(err instanceof Error ? err.message : "Something went wrong.");
    }
  }

  if (status === "success") {
    return (
      <div className="card">
        <p className="text-navy-900 dark:text-white">Thanks — we&apos;ve received your message.</p>
        <p className="mt-1 text-sm text-navy-500 dark:text-navy-400">
          This form is a development-stage stub (see the note above); we&apos;ll follow up by email.
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="card space-y-4">
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-navy-700 dark:text-navy-200">
          Name
        </label>
        <input
          id="name"
          name="name"
          type="text"
          required
          className="mt-1 w-full rounded-md border border-navy-200 bg-white px-3 py-2 text-sm text-navy-900 focus:border-navy-500 focus:outline-none dark:border-navy-700 dark:bg-navy-900 dark:text-white"
        />
      </div>
      <div>
        <label htmlFor="organization" className="block text-sm font-medium text-navy-700 dark:text-navy-200">
          Agency / organization
        </label>
        <input
          id="organization"
          name="organization"
          type="text"
          className="mt-1 w-full rounded-md border border-navy-200 bg-white px-3 py-2 text-sm text-navy-900 focus:border-navy-500 focus:outline-none dark:border-navy-700 dark:bg-navy-900 dark:text-white"
        />
      </div>
      <div>
        <label htmlFor="email" className="block text-sm font-medium text-navy-700 dark:text-navy-200">
          Email
        </label>
        <input
          id="email"
          name="email"
          type="email"
          required
          className="mt-1 w-full rounded-md border border-navy-200 bg-white px-3 py-2 text-sm text-navy-900 focus:border-navy-500 focus:outline-none dark:border-navy-700 dark:bg-navy-900 dark:text-white"
        />
      </div>
      <div>
        <label htmlFor="message" className="block text-sm font-medium text-navy-700 dark:text-navy-200">
          What are you looking to pilot?
        </label>
        <textarea
          id="message"
          name="message"
          rows={4}
          required
          className="mt-1 w-full rounded-md border border-navy-200 bg-white px-3 py-2 text-sm text-navy-900 focus:border-navy-500 focus:outline-none dark:border-navy-700 dark:bg-navy-900 dark:text-white"
        />
      </div>

      {status === "error" && <p className="text-sm text-red-600 dark:text-red-400">{errorMessage}</p>}

      <button type="submit" disabled={status === "submitting"} className="btn-primary w-full">
        {status === "submitting" ? "Sending…" : "Send"}
      </button>
    </form>
  );
}
