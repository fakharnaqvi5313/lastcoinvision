import Image from "next/image";
import Link from "next/link";

export function Logo({ size = 32, withWordmark = true }: { size?: number; withWordmark?: boolean }) {
  return (
    <Link href="/" className="flex items-center gap-2.5 font-semibold text-navy-900 dark:text-white">
      <Image src="/logo-192.png" alt="LastcoinVision logo" width={size} height={size} priority className="rounded-full" />
      {withWordmark && <span className="text-lg tracking-tight">LastcoinVision</span>}
    </Link>
  );
}
