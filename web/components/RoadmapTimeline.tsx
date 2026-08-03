import { roadmap, type RoadmapItem, type RoadmapStatus } from "@/lib/roadmap";

const statusStyles: Record<RoadmapStatus, string> = {
  shipped: "bg-accent-100 text-accent-800 dark:bg-accent-900/40 dark:text-accent-300",
  "in-progress": "bg-navy-100 text-navy-700 dark:bg-navy-800 dark:text-navy-200",
  planned: "border border-navy-200 text-navy-500 dark:border-navy-700 dark:text-navy-400",
};

const statusLabels: Record<RoadmapStatus, string> = {
  shipped: "Shipped",
  "in-progress": "In progress",
  planned: "Planned",
};

function StatusBadge({ status }: { status: RoadmapStatus }) {
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${statusStyles[status]}`}>
      {statusLabels[status]}
    </span>
  );
}

export function RoadmapTimeline({ detailed = false, items = roadmap }: { detailed?: boolean; items?: RoadmapItem[] }) {
  return (
    <ol className="relative space-y-8 border-l border-navy-200 pl-6 dark:border-navy-800">
      {items.map((item) => (
        <li key={item.id} className="relative">
          <span className="absolute -left-[29px] top-1 h-3 w-3 rounded-full border-2 border-white bg-navy-400 dark:border-navy-950 dark:bg-navy-600" />
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="font-semibold text-navy-900 dark:text-white">{item.title}</h2>
            <StatusBadge status={item.status} />
          </div>
          <p className="mt-1 text-sm text-navy-500 dark:text-navy-400">{item.summary}</p>
          {detailed && <p className="mt-2 text-sm text-navy-600 dark:text-navy-300">{item.detail}</p>}
        </li>
      ))}
    </ol>
  );
}
