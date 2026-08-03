/** @type {import('next').NextConfig} */
const nextConfig = {
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
