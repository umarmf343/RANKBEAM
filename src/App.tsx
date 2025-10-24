import { Footer } from "@/components/Footer";
import { Header } from "@/components/Header";
import { CompetitorsPage } from "@/pages/CompetitorsPage";
import { ExpansionPage } from "@/pages/ExpansionPage";
import { HomePage } from "@/pages/HomePage";
import { KeywordLabPage } from "@/pages/KeywordLabPage";
import { PlatformPage } from "@/pages/PlatformPage";
import { Route, Routes, useLocation } from "react-router-dom";
import { useEffect } from "react";

function ScrollManager() {
  const location = useLocation();
  const { pathname, hash } = location;

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "auto" });
  }, [pathname]);

  useEffect(() => {
    if (!hash) return;

    const id = hash.replace("#", "");
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }, [hash, pathname]);

  return null;
}

function App() {
  return (
    <div className="flex min-h-screen flex-col bg-night text-white">
      <Header />
      <main className="flex-1">
        <ScrollManager />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/platform" element={<PlatformPage />} />
          <Route path="/keyword-lab" element={<KeywordLabPage />} />
          <Route path="/competitors" element={<CompetitorsPage />} />
          <Route path="/expansion" element={<ExpansionPage />} />
          <Route path="*" element={<HomePage />} />
        </Routes>
      </main>
      <Footer />
    </div>
  );
}

export default App;
