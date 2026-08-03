import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { TEMPLATES, isValidTemplateId } from "./templates.js";

interface Args {
  projectDir: string;
  template: string | undefined;
  skipInstall: boolean;
  help: boolean;
}

function parseArgs(argv: string[]): Args {
  let projectDir = "";
  let template: string | undefined;
  let skipInstall = false;
  let help = false;

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--template" || arg === "-t") {
      template = argv[++i];
    } else if (arg?.startsWith("--template=")) {
      template = arg.slice("--template=".length);
    } else if (arg === "--skip-install") {
      skipInstall = true;
    } else if (arg === "--help" || arg === "-h") {
      help = true;
    } else if (arg && !arg.startsWith("-") && !projectDir) {
      projectDir = arg;
    }
  }

  return { projectDir, template, skipInstall, help };
}

function printHelp(): void {
  console.log(`Usage: npx create-lcv-app <project-directory> --template <template>

Templates:
${TEMPLATES.map((t) => `  ${t.id.padEnd(14)} ${t.description}`).join("\n")}

Options:
  --template, -t <id>   Template to scaffold (required)
  --skip-install        Skip running "npm install" after scaffolding
  --help, -h             Show this help message
`);
}

// This package ships as ESM with dist/index.js as its bin entry, so
// import.meta.url always points at either dist/index.js (built) or
// src/index.ts (via `npm run dev` / tsx) — one level up from either is
// this package's own root, which is where the sibling `templates/`
// directory (and, in a monorepo checkout, the sibling `sdk/` package)
// live relative to it.
const cliPackageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

function resolveMonorepoSdkDir(): string | undefined {
  const candidate = path.resolve(cliPackageDir, "..", "sdk");
  return fs.existsSync(path.join(candidate, "package.json")) ? candidate : undefined;
}

function copyTemplate(templateId: string, targetDir: string): void {
  const templateDir = path.join(cliPackageDir, "templates", templateId);
  if (!fs.existsSync(templateDir)) {
    throw new Error(`Template "${templateId}" not found at ${templateDir}`);
  }
  fs.cpSync(templateDir, targetDir, { recursive: true });

  // Templates ship `gitignore` (not `.gitignore`) and `env.local.example`
  // (not dot-prefixed) because npm silently drops dotfiles named `.gitignore`
  // and `.npmignore` from published tarballs — renaming on write is the
  // standard workaround every scaffolder (create-next-app included) uses.
  const renames: Array<[string, string]> = [
    ["gitignore", ".gitignore"],
    ["env.local.example", ".env.local.example"],
  ];
  for (const [from, to] of renames) {
    const fromPath = path.join(targetDir, from);
    if (fs.existsSync(fromPath)) {
      fs.renameSync(fromPath, path.join(targetDir, to));
    }
  }
}

function writeProjectPackageJson(targetDir: string, projectName: string, monorepoSdkDir: string | undefined): void {
  const pkgPath = path.join(targetDir, "package.json");
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8")) as {
    name: string;
    dependencies?: Record<string, string>;
  };
  pkg.name = projectName;

  if (monorepoSdkDir && pkg.dependencies?.["@lastcoinvision/sdk"]) {
    // @lastcoinvision/sdk isn't published yet. When create-lcv-app is run
    // from within this monorepo checkout (dev/testing), point at the local
    // built package via an absolute file: path so `npm install` succeeds
    // today; see DECISIONS.md. A real npm publish makes this branch dead
    // code — every consumer will just get the semver range below.
    pkg.dependencies["@lastcoinvision/sdk"] = `file:${monorepoSdkDir}`;
  }

  fs.writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`);
}

function writeDefaultEnvFile(targetDir: string): void {
  const examplePath = path.join(targetDir, ".env.local.example");
  const targetPath = path.join(targetDir, ".env.local");
  if (fs.existsSync(examplePath) && !fs.existsSync(targetPath)) {
    fs.copyFileSync(examplePath, targetPath);
  }
}

function main(): void {
  const args = parseArgs(process.argv.slice(2));

  if (args.help || !args.projectDir || !args.template) {
    printHelp();
    process.exitCode = args.help ? 0 : 1;
    return;
  }

  if (!isValidTemplateId(args.template)) {
    console.error(`Unknown template "${args.template}". Available: ${TEMPLATES.map((t) => t.id).join(", ")}`);
    process.exitCode = 1;
    return;
  }

  const targetDir = path.resolve(process.cwd(), args.projectDir);
  if (fs.existsSync(targetDir) && fs.readdirSync(targetDir).length > 0) {
    console.error(`Target directory "${targetDir}" already exists and is not empty.`);
    process.exitCode = 1;
    return;
  }

  const projectName = path.basename(targetDir);
  console.log(`Scaffolding "${args.template}" into ${targetDir} ...`);
  copyTemplate(args.template, targetDir);
  writeProjectPackageJson(targetDir, projectName, resolveMonorepoSdkDir());
  writeDefaultEnvFile(targetDir);

  if (!args.skipInstall) {
    console.log("Installing dependencies with npm ...");
    const result = spawnSync("npm", ["install"], { cwd: targetDir, stdio: "inherit" });
    if (result.status !== 0) {
      console.error("npm install failed. Run it manually inside the project directory.");
      process.exitCode = result.status ?? 1;
      return;
    }
  }

  console.log(`
Done! Next steps:

  cd ${args.projectDir}
${args.skipInstall ? "  npm install\n" : ""}  npm run dev

The app defaults to a local devnet (RPC/REST on localhost, chain-id lcv-devnet-1) —
see .env.local to point it at a different network.
`);
}

main();
