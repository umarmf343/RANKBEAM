export function Footer() {
  return (
    <footer className="border-t border-white/5 bg-black/80">
      <div className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-10 text-sm text-white/50 md:flex-row md:items-center md:justify-between">
        <p>© {new Date().getFullYear()} RankBeam Labs. All rights reserved.</p>
        <nav className="flex flex-wrap gap-4">
          {[
            "Privacy",
            "Terms",
            "Security",
            "Support"
          ].map((item) => (
            <a key={item} href="#" className="hover:text-white">
              {item}
            </a>
          ))}
        </nav>
      </div>
    </footer>
  );
}
