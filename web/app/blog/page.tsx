import type { Metadata } from "next";
import Link from "next/link";

import { posts } from "@/lib/posts";

export const metadata: Metadata = {
  title: "Blog",
  description: "Notes on what's actually being built, including the parts that aren't finished yet.",
};

export default function BlogIndexPage() {
  const sorted = [...posts].sort((a, b) => b.date.localeCompare(a.date));

  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Blog</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Notes from the build</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          Technical write-ups on what&apos;s shipped — including the parts that are intentionally incomplete.
        </p>
      </div>

      <div className="mt-12 max-w-2xl divide-y divide-navy-100 dark:divide-navy-800">
        {sorted.map((post) => (
          <Link key={post.slug} href={`/blog/${post.slug}`} className="block py-6 first:pt-0">
            <time className="text-xs font-medium uppercase tracking-wider text-navy-500 dark:text-navy-400">
              {new Date(post.date).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" })}
            </time>
            <h2 className="mt-1 text-xl font-semibold text-navy-900 dark:text-white">{post.title}</h2>
            <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">{post.description}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
