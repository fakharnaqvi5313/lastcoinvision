import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";
import { themes as prismThemes } from "prism-react-renderer";

const config: Config = {
  title: "LastcoinVision Docs",
  tagline: "Documentation for the LastcoinVision (LCV) Layer-1",
  favicon: "img/favicon.png",

  url: "https://docs.lastcoinvision.com",
  baseUrl: "/",

  organizationName: "lastcoinvision",
  projectName: "lastcoinvision-docs",

  // "warn" rather than "throw": the auto-included modules/*/SPEC.md files
  // (see the "modules" plugin instance below) intentionally use
  // GitHub-relative links like ../../DECISIONS.md and ../../chain/x/foo,
  // which resolve correctly on GitHub but point outside this docs root and
  // can never resolve to a Docusaurus route — see DECISIONS.md.
  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "warn",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          path: "docs",
          routeBasePath: "/",
          sidebarPath: "./sidebars.ts",
          editUrl: undefined,
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    [
      "@docusaurus/plugin-content-docs",
      {
        id: "modules",
        path: "../modules",
        routeBasePath: "modules",
        sidebarPath: "./sidebarsModules.ts",
        editUrl: undefined,
        exclude: ["README.md", "x/**"],
      },
    ],
  ],

  themeConfig: {
    image: "img/social-card.png",
    navbar: {
      title: "LastcoinVision Docs",
      logo: {
        alt: "LastcoinVision logo",
        src: "img/logo.png",
      },
      items: [
        { type: "docSidebar", sidebarId: "docsSidebar", position: "left", label: "Docs" },
        { to: "/modules/", label: "Module Reference", position: "left" },
        { to: "/api-reference", label: "API Reference", position: "left" },
        { href: "https://lastcoinvision.com", label: "lastcoinvision.com", position: "right" },
        { href: "https://faucet.lastcoinvision.com", label: "Faucet", position: "right" },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Docs",
          items: [
            { label: "Getting Started", to: "/getting-started/run-a-node" },
            { label: "Tutorials", to: "/tutorials/build-a-voting-portal" },
            { label: "Module Reference", to: "/modules/" },
            { label: "API Reference", to: "/api-reference" },
          ],
        },
        {
          title: "Network",
          items: [
            { label: "Faucet", href: "https://faucet.lastcoinvision.com" },
            { label: "Explorer", href: "https://explorer.lastcoinvision.com" },
          ],
        },
        {
          title: "More",
          items: [
            { label: "lastcoinvision.com", href: "https://lastcoinvision.com" },
            { label: "Brand assets", href: "https://lastcoinvision.com/brand" },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} LastcoinVision. Testnet/devnet software.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ["bash", "solidity", "toml", "protobuf"],
    },
    colorMode: {
      defaultMode: "light",
      respectPrefersColorScheme: true,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
