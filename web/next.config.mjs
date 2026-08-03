import createMDX from "@next/mdx";

/** @type {import('next').NextConfig} */
const nextConfig = {
  pageExtensions: ["ts", "tsx", "mdx"],
  // Lets the Docker image (deploy/docker-compose) ship only the pruned
  // server + traced dependencies instead of the full node_modules tree.
  output: "standalone",
  images: {
    // Static-hosting-friendly: no server-side image optimization endpoint
    // required (see DECISIONS.md for the "static-exportable" caveat this
    // site actually ships with — one API route for the pilot-program form).
    unoptimized: true,
  },
};

const withMDX = createMDX({});

export default withMDX(nextConfig);
