import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy — LCV Wallet",
  description: "How the LastcoinVision wallet apps handle keys and data. Non-custodial: your keys never leave your device.",
};

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mt-10">
      <h2 className="text-xl font-semibold text-navy-900 dark:text-white">{title}</h2>
      <div className="mt-3 space-y-3 text-navy-600 dark:text-navy-300">{children}</div>
    </section>
  );
}

export default function PrivacyPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Legal</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Privacy Policy</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          This page covers the LCV Wallet apps (desktop and mobile). It does not cover
          lastcoinvision.com&apos;s own web analytics, if any are added separately.
        </p>
        <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">Last updated 2026-08-07.</p>
      </div>

      <div className="mt-4 max-w-2xl">
        <Section title="This is a non-custodial wallet">
          <p>
            LastcoinVision never has access to your private keys, recovery phrase, or funds. Key
            generation, encryption, and transaction signing all happen locally on your device.
            Nothing sensitive is ever sent to us or to any third party — there is no server in
            this wallet&apos;s architecture that could see or store your keys even if it wanted
            to.
          </p>
        </Section>

        <Section title="What's stored, and where">
          <p>
            Your recovery phrase is encrypted at rest and stored only on your device, using your
            operating system&apos;s native secure storage: Keychain on macOS/iOS, Credential
            Manager on Windows, Secret Service on Linux, or Keystore on Android. We do not
            operate a backend that stores wallet data. If you delete the app or use &quot;forget
            this device,&quot; that local copy is gone — we cannot recover it, and neither can
            anyone else. Back up your recovery phrase yourself when the app shows it to you.
          </p>
        </Section>

        <Section title="What the app sends over the network">
          <p>
            To check balances and broadcast transactions, the app talks directly to public
            LastcoinVision blockchain nodes (RPC/REST endpoints) — the same kind of public
            infrastructure any blockchain wallet uses. This means the endpoint operator can see
            your IP address and the (public, on-chain) transactions you send, the same as any
            node operator can for any blockchain traffic. We do not add any analytics, tracking,
            or crash-reporting SDK to the wallet apps as of this writing; if that changes in a
            future release, this page will be updated first.
          </p>
        </Section>

        <Section title="No account, no personal data collection">
          <p>
            There is no sign-up, no email address, and no password tied to your identity — the
            only &quot;password&quot; is the one you choose to encrypt your local keystore, and
            we never see it. We do not collect names, emails, or other personal information
            through the wallet apps.
          </p>
        </Section>

        <Section title="Testnet software">
          <p>
            The wallet apps currently point at LastcoinVision&apos;s public testnet by default.
            Testnet tokens have no monetary value. Treat any testnet keys as disposable — don&apos;t
            reuse a testnet recovery phrase for anything you&apos;d consider valuable.
          </p>
        </Section>

        <Section title="Questions">
          <p>
            Contact us via the channels listed on{" "}
            <a href="/" className="text-accent-700 underline underline-offset-4 dark:text-accent-500">
              lastcoinvision.com
            </a>
            .
          </p>
        </Section>
      </div>
    </div>
  );
}
