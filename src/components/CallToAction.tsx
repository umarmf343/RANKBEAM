import { ArrowRightCircle } from "lucide-react";
import { Link } from "react-router-dom";

export function CallToAction() {
  return (
    <section id="cta" className="bg-gradient-to-b from-night via-night to-black">
      <div className="mx-auto max-w-5xl px-6 py-20 text-center">
        <div className="inline-flex rounded-full border border-aurora-500/30 bg-aurora-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-wide text-aurora-100">
          Launch faster with RankBeam
        </div>
        <h2 className="mt-6 font-display text-3xl font-semibold text-white sm:text-4xl">
          Ready to turn insights into top-ranking listings?
        </h2>
        <p className="mx-auto mt-3 max-w-2xl text-sm text-white/60">
          Start a 7-day growth sprint and unlock unlimited keyword scans, competitor playbooks and AI-assisted listing
          blueprints. Cancel anytime.
        </p>
        <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
          <a
            href="#"
            className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-aurora-400 to-aurora-600 px-6 py-3 font-semibold text-white shadow-glow"
          >
            Start 7-day trial <ArrowRightCircle className="h-5 w-5" />
          </a>
          <Link to="/platform" className="text-sm font-semibold text-white/70 hover:text-white">
            Download product overview ↗
          </Link>
        </div>
      </div>
    </section>
  );
}
