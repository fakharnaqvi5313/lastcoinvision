function Band({
  x,
  y,
  width,
  height,
  label,
  sublabel,
}: {
  x: number;
  y: number;
  width: number;
  height: number;
  label: string;
  sublabel?: string;
}) {
  return (
    <g>
      <rect
        x={x}
        y={y}
        width={width}
        height={height}
        rx={10}
        className="fill-white stroke-navy-200 dark:fill-navy-900 dark:stroke-navy-700"
        strokeWidth={1.5}
      />
      <text x={x + width / 2} y={y + height / 2 - (sublabel ? 6 : 0)} textAnchor="middle" dominantBaseline="middle" className="fill-navy-900 text-[13px] font-semibold dark:fill-white">
        {label}
      </text>
      {sublabel && (
        <text x={x + width / 2} y={y + height / 2 + 14} textAnchor="middle" dominantBaseline="middle" className="fill-navy-400 text-[10px] dark:fill-navy-400">
          {sublabel}
        </text>
      )}
    </g>
  );
}

function Arrow({ x, yStart, yEnd }: { x: number; yStart: number; yEnd: number }) {
  return (
    <g>
      <line x1={x} y1={yStart} x2={x} y2={yEnd - 8} className="stroke-navy-300 dark:stroke-navy-600" strokeWidth={2} />
      <polygon points={`${x - 5},${yEnd - 8} ${x + 5},${yEnd - 8} ${x},${yEnd}`} className="fill-navy-300 dark:fill-navy-600" />
    </g>
  );
}

/**
 * Hand-authored architecture SVG (not a Mermaid render) so the page stays
 * static and dependency-free — see DECISIONS.md.
 */
export function GovZoneDiagram() {
  return (
    <svg viewBox="0 0 800 460" role="img" aria-label="GovZone architecture: citizens and agencies access government portals, which call the LCV SDK, which talks to the identity, registry, voting, and taxrail chain modules, secured by CometBFT consensus." className="w-full">
      <Band x={200} y={16} width={400} height={56} label="Citizens & government agencies" sublabel="web / mobile portal access" />
      <Arrow x={400} yStart={72} yEnd={110} />

      <rect x={140} y={104} width={520} height={96} rx={12} className="fill-accent-50/50 stroke-accent-300 stroke-dasharray-4 dark:fill-accent-900/10 dark:stroke-accent-700" strokeWidth={1.5} strokeDasharray="4 3" />
      <text x={400} y={100} textAnchor="middle" className="fill-accent-700 text-[11px] font-semibold uppercase tracking-wider dark:fill-accent-400">
        GovZone applications
      </text>
      <Band x={160} y={122} width={150} height={60} label="Voting portal" />
      <Band x={325} y={122} width={150} height={60} label="Registry UI" />
      <Band x={490} y={122} width={150} height={60} label="Tax portal" />

      <Arrow x={400} yStart={200} yEnd={238} />
      <Band x={220} y={238} width={360} height={50} label="@lastcoinvision/sdk" sublabel="typed clients, wallet helpers" />

      <Arrow x={400} yStart={288} yEnd={326} />
      <rect x={80} y={320} width={640} height={80} rx={12} className="fill-navy-50/60 stroke-navy-200 dark:fill-navy-900/40 dark:stroke-navy-700" strokeWidth={1.5} strokeDasharray="4 3" />
      <text x={400} y={316} textAnchor="middle" className="fill-navy-500 text-[11px] font-semibold uppercase tracking-wider dark:fill-navy-400">
        LCV chain modules
      </text>
      <Band x={100} y={338} width={140} height={48} label="x/identity" />
      <Band x={253} y={338} width={140} height={48} label="x/registry" />
      <Band x={406} y={338} width={140} height={48} label="x/voting" />
      <Band x={559} y={338} width={140} height={48} label="x/taxrail" />

      <Arrow x={400} yStart={400} yEnd={412} />
      <Band x={200} y={412} width={400} height={40} label="CometBFT consensus — validator set" />
    </svg>
  );
}
