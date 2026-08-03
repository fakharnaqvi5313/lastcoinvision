"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

import { siteConfig } from "@/lib/siteConfig";
import { Logo } from "./Logo";
import { ThemeToggle } from "./ThemeToggle";

export function Nav() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 border-b border-navy-100 bg-white/90 backdrop-blur dark:border-navy-800 dark:bg-navy-950/90">
      <div className="container-page flex h-16 items-center justify-between">
        <Logo />

        <nav className="hidden items-center gap-1 lg:flex">
          {siteConfig.nav.map((item) => {
            const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  active
                    ? "text-navy-900 dark:text-white"
                    : "text-navy-500 hover:text-navy-900 dark:text-navy-300 dark:hover:text-white"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="hidden items-center gap-3 lg:flex">
          <ThemeToggle />
          <Link href="/developers" className="btn-primary">
            Get started
          </Link>
        </div>

        <div className="flex items-center gap-2 lg:hidden">
          <ThemeToggle />
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            aria-label="Toggle navigation menu"
            aria-expanded={open}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-navy-200 text-navy-700 dark:border-navy-700 dark:text-navy-200"
          >
            <svg viewBox="0 0 24 24" fill="none" className="h-5 w-5" aria-hidden>
              {open ? (
                <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
              ) : (
                <path d="M4 7h16M4 12h16M4 17h16" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
              )}
            </svg>
          </button>
        </div>
      </div>

      {open && (
        <nav className="border-t border-navy-100 bg-white px-4 py-3 dark:border-navy-800 dark:bg-navy-950 lg:hidden">
          <ul className="flex flex-col gap-1">
            {siteConfig.nav.map((item) => (
              <li key={item.href}>
                <Link
                  href={item.href}
                  onClick={() => setOpen(false)}
                  className="block rounded-md px-3 py-2 text-sm font-medium text-navy-700 hover:bg-navy-50 dark:text-navy-200 dark:hover:bg-navy-900"
                >
                  {item.label}
                </Link>
              </li>
            ))}
            <li className="pt-2">
              <Link href="/developers" onClick={() => setOpen(false)} className="btn-primary w-full">
                Get started
              </Link>
            </li>
          </ul>
        </nav>
      )}
    </header>
  );
}
