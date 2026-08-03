import { useCases } from "@/lib/useCases";

export function UseCaseGrid() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {useCases.map((useCase) => (
        <div key={useCase.id} className="card flex flex-col">
          <span className="eyebrow">{useCase.tag}</span>
          <h3 className="mt-2 text-lg font-semibold text-navy-900 dark:text-white">{useCase.title}</h3>
          <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">{useCase.description}</p>
        </div>
      ))}
    </div>
  );
}
