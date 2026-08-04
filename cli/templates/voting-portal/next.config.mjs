/** @type {import('next').NextConfig} */
const nextConfig = {
  // create-lcv-app patches @lastcoinvision/sdk to a file: path when
  // scaffolded from inside the monorepo (it isn't published yet — see
  // cli/src/index.ts), which npm links in as a symlink pointing outside
  // this project directory. Turbopack (Next 16's default bundler, for
  // `build` too, not just `dev`) refuses to resolve modules outside its
  // detected project root unless told to.
  experimental: {
    externalDir: true,
  },
};

export default nextConfig;
