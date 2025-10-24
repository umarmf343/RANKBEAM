import { useRankBeamStore } from "@/lib/state";
import { BadgeCheck, CheckCircle2, Copy, CopyCheck } from "lucide-react";
import { useMemo, useState } from "react";

function buildBlueprintCopy(title: string, subtitle: string, bullets: string[]): string {
  const bulletText = bullets.map((item, index) => `${index + 1}. ${item}`).join("\n");
  return `Title: ${title}\nSubtitle: ${subtitle}\nKey bullets:\n${bulletText}`;
}

export function ListingBlueprint() {
  const { headlineIdeas, keywordInsights, keyword, country } = useRankBeamStore();
  const [copied, setCopied] = useState(false);

  const blueprint = useMemo(() => {
    const primaryHeadline = headlineIdeas[0] ?? `Essential ${keyword ? keyword.trim() : "Amazon"} Blueprint`;
    const subtitleKeyword = keywordInsights[0]?.keyword ?? keyword;
    const supportingKeywords = keywordInsights.slice(0, 4).map((item) => item.keyword);

    const subtitle = subtitleKeyword
      ? `Launch a ${subtitleKeyword.toLowerCase()} listing optimised for ${country.label}`
      : `Launch with algorithm-ready metadata across ${country.label}`;

    const bullets = [
      supportingKeywords[0]
        ? `Aligns with top intent phrase "${supportingKeywords[0]}" while signalling premium positioning`
        : "Aligns with the highest intent reader keyword to boost conversion",
      supportingKeywords[1]
        ? `Builds in secondary demand for "${supportingKeywords[1]}" to diversify ad groups`
        : "Stacks secondary keyword clusters so ads and organic rank reinforce each other",
      supportingKeywords[2]
        ? `Adds value proof tied to ${keywordInsights[2]?.searchVolume.toLocaleString()} monthly searches`
        : "Showcases tangible reader outcomes to raise click-through rate",
      `CTA: Activate a ${country.currency} ${country.code === "US" ? "7-day" : "launch"} promo to spike initial velocity`
    ];

    return {
      title: primaryHeadline,
      subtitle,
      supportingKeywords,
      bullets
    };
  }, [headlineIdeas, keywordInsights, keyword, country]);

  if (headlineIdeas.length === 0 && keywordInsights.length === 0) {
    return null;
  }

  const copyPayload = buildBlueprintCopy(blueprint.title, blueprint.subtitle, blueprint.bullets);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(copyPayload);
      setCopied(true);
      setTimeout(() => setCopied(false), 2200);
    } catch (error) {
      console.error("Failed to copy listing blueprint", error);
    }
  };

  return (
    <section className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">AI listing blueprint</h2>
            <p className="mt-2 max-w-3xl text-sm text-white/70">
              Transform keyword intelligence into launch-ready copy. RankBeam drafts a headline, positioning angle and bullet
              architecture tuned for {country.label} so you can brief designers and writers instantly.
            </p>
          </div>
          <button
            type="button"
            onClick={handleCopy}
            className="inline-flex items-center gap-2 rounded-full border border-aurora-500/40 bg-aurora-500/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-100 transition hover:border-aurora-400 hover:text-aurora-50"
          >
            {copied ? (
              <>
                <CopyCheck className="h-4 w-4" /> Blueprint copied
              </>
            ) : (
              <>
                <Copy className="h-4 w-4" /> Copy blueprint
              </>
            )}
          </button>
        </div>
        <div className="mt-10 grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
          <div className="space-y-6 rounded-3xl border border-white/10 bg-black/40 p-8 shadow-[0_30px_100px_-60px_rgba(76,102,241,0.45)]">
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-wide text-white/50">Hero title</p>
              <h3 className="font-display text-2xl font-semibold text-white">{blueprint.title}</h3>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-white/50">Subtitle direction</p>
              <p className="mt-2 text-sm text-white/70">{blueprint.subtitle}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-white/50">Bullet architecture</p>
              <ul className="mt-3 space-y-3 text-sm text-white/70">
                {blueprint.bullets.map((bullet) => (
                  <li key={bullet} className="flex items-start gap-3">
                    <CheckCircle2 className="mt-1 h-4 w-4 flex-none text-aurora-300" />
                    <span>{bullet}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <aside className="space-y-5 rounded-3xl border border-white/10 bg-gradient-to-br from-black/60 via-black/30 to-aurora-500/10 p-6">
            <div className="flex items-center gap-3">
              <BadgeCheck className="h-5 w-5 text-aurora-300" />
              <p className="text-sm font-semibold text-white">Keyword anchors</p>
            </div>
            <ul className="space-y-2 text-sm text-white/70">
              {blueprint.supportingKeywords.map((keywordValue) => (
                <li key={keywordValue} className="rounded-full border border-white/10 bg-black/40 px-3 py-1">
                  {keywordValue}
                </li>
              ))}
              {blueprint.supportingKeywords.length === 0 && <li>No supporting keywords available yet.</li>}
            </ul>
            <div className="rounded-2xl border border-white/10 bg-black/50 p-4 text-xs text-white/60">
              <p className="font-semibold uppercase tracking-wide text-white/50">Launch guidance</p>
              <p className="mt-2">
                Pair this blueprint with a {country.currency} 100 ad test split across broad and exact campaigns. Monitor click
                through rate for three days and feed winners back into RankBeam keyword batches.
              </p>
            </div>
          </aside>
        </div>
      </div>
    </section>
  );
}
