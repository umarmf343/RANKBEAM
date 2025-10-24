import { CallToAction } from "@/components/CallToAction";
import { CompetitorShowcase } from "@/components/CompetitorShowcase";
import { CreativeStudio } from "@/components/CreativeStudio";
import { ExpansionPanel } from "@/components/ExpansionPanel";
import { Footer } from "@/components/Footer";
import { Header } from "@/components/Header";
import { Hero } from "@/components/Hero";
import { InsightSummary } from "@/components/InsightSummary";
import { KeywordTable } from "@/components/KeywordTable";
import { KeywordWorkbench } from "@/components/KeywordWorkbench";
import { SignalsPanel } from "@/components/SignalsPanel";

function App() {
  return (
    <div className="flex min-h-screen flex-col bg-night text-white">
      <Header />
      <main className="flex-1">
        <Hero />
        <KeywordWorkbench />
        <InsightSummary />
        <KeywordTable />
        <SignalsPanel />
        <CompetitorShowcase />
        <ExpansionPanel />
        <CreativeStudio />
        <CallToAction />
      </main>
      <Footer />
    </div>
  );
}

export default App;
