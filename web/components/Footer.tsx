import Link from "next/link";

import { siteConfig } from "@/lib/siteConfig";
import { Logo } from "./Logo";

export function Footer() {
  return (
    <footer className="border-t border-navy-100 bg-navy-50/50 dark:border-navy-800 dark:bg-navy-950">
      <div className="container-page grid gap-10 py-12 md:grid-cols-[2fr_1fr_1fr_1fr_1fr]">
        <div>
          <Logo />
          <p className="mt-4 max-w-xs text-sm text-navy-500 dark:text-navy-400">{siteConfig.description}</p>
        </div>

        {Object.entries(siteConfig.footerLinks).map(([heading, links]) => (
          <div key={heading}>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-navy-500 dark:text-navy-400">{heading}</h3>
            <ul className="mt-3 space-y-2">
              {links.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-sm text-navy-600 hover:text-navy-900 dark:text-navy-300 dark:hover:text-white"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="border-t border-navy-100 py-6 dark:border-navy-800">
        <div className="container-page flex flex-col gap-2 text-xs text-navy-500 sm:flex-row sm:items-center sm:justify-between dark:text-navy-400">
          <p>© {new Date().getFullYear()} LastcoinVision. Testnet/devnet software — not an offer or solicitation of securities.</p>
          <p>Bech32 prefix <code className="rounded bg-navy-100 px-1 py-0.5 dark:bg-navy-800">lcv</code></p>
        </div>
      </div>
    </footer>
  );
}
