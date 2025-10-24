import { Sparkles } from "lucide-react";
import { Link, NavLink } from "react-router-dom";

const navigation = [
  { name: "Platform", to: "/platform" },
  { name: "Keyword Lab", to: "/keyword-lab" },
  { name: "Competitors", to: "/competitors" },
  { name: "Expansion", to: "/expansion" }
];

export function Header() {
  return (
    <header className="sticky top-0 z-30 border-b border-white/5 bg-night/70 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link to="/" className="flex items-center gap-2 font-display text-lg font-semibold">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-aurora-500/20 text-aurora-400">
            <Sparkles className="h-5 w-5" />
          </span>
          RankBeam
        </Link>
        <nav className="hidden items-center gap-8 text-sm font-medium text-white/70 md:flex">
          {navigation.map((item) => (
            <NavLink
              key={item.name}
              to={item.to}
              className={({ isActive }) =>
                `transition hover:text-white ${isActive ? "text-white" : "text-white/70"}`
              }
            >
              {item.name}
            </NavLink>
          ))}
        </nav>
        <div className="hidden items-center gap-3 md:flex">
          <Link
            className="rounded-full border border-white/10 px-4 py-2 text-sm font-medium text-white/70 transition hover:border-aurora-400 hover:text-white"
            to="/platform"
          >
            Explore Platform
          </Link>
          <Link
            className="rounded-full bg-gradient-to-br from-aurora-400 to-aurora-600 px-4 py-2 text-sm font-semibold text-white shadow-glow"
            to="/#cta"
          >
            Start Trial
          </Link>
        </div>
      </div>
    </header>
  );
}
