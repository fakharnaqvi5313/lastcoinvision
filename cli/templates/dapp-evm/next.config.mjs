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
};

export default nextConfig;
