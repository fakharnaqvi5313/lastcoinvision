import type { MDXComponents } from "mdx/types";
import type { ReactNode } from "react";

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    wrapper: ({ children }: { children: ReactNode }) => (
      <article className="container-page prose max-w-3xl py-16 dark:prose-invert sm:py-20">{children}</article>
    ),
    ...components,
  };
}
