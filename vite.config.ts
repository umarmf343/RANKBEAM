import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import type { Plugin } from "vite";
import { scrapeAmazonKeywordData } from "./src/server/amazon";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function keywordApiPlugin(): Plugin {
  return {
    name: "rankbeam-amazon-keyword-proxy",
    configureServer(server) {
      server.middlewares.use("/api/keywords", async (req, res) => {
        if (!req.url) {
          res.statusCode = 400;
          res.end("Missing request URL");
          return;
        }

        if (req.method && req.method !== "GET") {
          res.statusCode = 405;
          res.setHeader("Allow", "GET");
          res.end("Method not allowed");
          return;
        }

        try {
          const requestUrl = new URL(req.url, "http://localhost");
          const keyword = requestUrl.searchParams.get("keyword") ?? "";
          const country = requestUrl.searchParams.get("country") ?? "US";
          const payload = await scrapeAmazonKeywordData(keyword, country);
          const response = {
            keyword,
            country,
            scrapedAt: new Date().toISOString(),
            ...payload
          };
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify(response));
        } catch (error) {
          console.error("Keyword API error", error);
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ error: "Failed to retrieve keyword intelligence" }));
        }
      });
    }
  };
}

export default defineConfig({
  plugins: [react(), keywordApiPlugin()],
  resolve: {
    alias: {
      "@": resolve(__dirname, "src")
    }
  },
  server: {
    port: 5173,
    host: "0.0.0.0"
  }
});
