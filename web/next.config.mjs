import createMDX from "@next/mdx";

/** @type {import('next').NextConfig} */
const nextConfig = {
  pageExtensions: ["ts", "tsx", "mdx"],
  images: {
    // Static-hosting-friendly: no server-side image optimization endpoint
    // required (see DECISIONS.md for the "static-exportable" caveat this
    // site actually ships with — one API route for the pilot-program form).
    unoptimized: true,
  },
};

const withMDX = createMDX({});

export default withMDX(nextConfig);
