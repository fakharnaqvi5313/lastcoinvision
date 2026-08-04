/** @type {import('next').NextConfig} */
const nextConfig = {
  // Only relevant if @lastcoinvision/sdk is linked in from outside this
  // project directory (e.g. a monorepo-local `file:` dependency during
  // dev/testing, before the package is published) — harmless no-op for a
  // normal npm-installed dependency. See create-lcv-app's own patching of
  // this file when scaffolding from within the monorepo.
  experimental: {
    externalDir: true,
  },
  webpack: (config) => {
    // wagmi/connectors barrel-exports Coinbase's `baseAccount` connector
    // (which this template doesn't use — only `injected`), and that
    // connector pulls in @coinbase/cdp-sdk's optional x402 payment
    // integration, which references packages not meant to be bundled
    // client-side. Alias them out rather than installing dead weight.
    config.resolve.alias = {
      ...config.resolve.alias,
      "@x402/core/client": false,
      "@x402/svm/exact/client": false,
      "@x402/evm": false,
    };
    return config;
  },
};

export default nextConfig;
