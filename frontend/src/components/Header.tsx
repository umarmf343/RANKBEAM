import { Sparkles } from "lucide-react";

const navigation = [
  { name: "Platform", href: "#platform" },
  { name: "Keyword Lab", href: "#keywords" },
  { name: "Competitors", href: "#competitors" },
  { name: "Expansion", href: "#expansion" }
];

export function Header() {
  return (
    <header className="sticky top-0 z-30 border-b border-white/5 bg-night/70 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <a href="#top" className="flex items-center gap-2 font-display text-lg font-semibold">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-aurora-500/20 text-aurora-400">
            <Sparkles className="h-5 w-5" />
          </span>
          RankBeam
        </a>
        <nav className="hidden items-center gap-8 text-sm font-medium text-white/70 md:flex">
          {navigation.map((item) => (
            <a key={item.name} href={item.href} className="transition hover:text-white">
              {item.name}
            </a>
          ))}
        </nav>
        <div className="hidden items-center gap-3 md:flex">
          <a
            className="rounded-full border border-white/10 px-4 py-2 text-sm font-medium text-white/70 transition hover:border-aurora-400 hover:text-white"
            href="#platform"
          >
            Explore Platform
          </a>
          <a
            className="rounded-full bg-gradient-to-br from-aurora-400 to-aurora-600 px-4 py-2 text-sm font-semibold text-white shadow-glow"
            href="#cta"
          >
            Start Trial
          </a>
        </div>
      </div>
    </header>
  );
}
