import { CallToAction } from "@/components/CallToAction";
import { CompetitorShowcase } from "@/components/CompetitorShowcase";
import { ExpansionPanel } from "@/components/ExpansionPanel";
import { Footer } from "@/components/Footer";
import { Header } from "@/components/Header";
import { Hero } from "@/components/Hero";
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
        <KeywordTable />
        <SignalsPanel />
        <CompetitorShowcase />
        <ExpansionPanel />
        <CallToAction />
      </main>
      <Footer />
    </div>
  );
}

export default App;
