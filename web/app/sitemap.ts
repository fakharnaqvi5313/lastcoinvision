import type { MetadataRoute } from "next";

import { posts } from "@/lib/posts";
import { siteConfig } from "@/lib/siteConfig";

export default function sitemap(): MetadataRoute.Sitemap {
  const staticRoutes = ["/", "/developers", "/governments", "/technology", "/token", "/roadmap", "/blog", "/brand"].map(
    (route) => ({
      url: `${siteConfig.url}${route}`,
      lastModified: new Date(),
    }),
  );

  const postRoutes = posts.map((post) => ({
    url: `${siteConfig.url}/blog/${post.slug}`,
    lastModified: new Date(post.date),
  }));

  return [...staticRoutes, ...postRoutes];
}
