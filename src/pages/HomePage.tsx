import { CallToAction } from "@/components/CallToAction";
import { Hero } from "@/components/Hero";
import { InsightSummary } from "@/components/InsightSummary";

export function HomePage() {
  return (
    <>
      <Hero />
      <InsightSummary />
      <CallToAction />
    </>
  );
}
