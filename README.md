# RankBeam

RankBeam is a concept front-end for an Amazon keyword intelligence workspace. The Vite + React application stitches together hero marketing content with interactive keyword research panels that model how authors or private-label sellers could identify new opportunities across marketplaces.

## Features

- **Immersive landing experience** – The hero and marketing panels introduce RankBeam's positioning while highlighting marketplace coverage, data-backed feature cards, and prominent calls-to-action for scanning keywords or exploring capabilities.
- **Keyword Intelligence Workbench** – Seed keyword search, marketplace selection, debounced updates, and quick-glance benefit cards help authors explore keyword ideas with minimal latency.
- **Opportunity radar dashboards** – Keyword, competitor, and international datasets are summarised into cards and tables that surface high-demand phrases, opportunity counts, competitive baselines, and localisation wins.
- **Interactive opportunity table** – Adjustable sliders and toggles drive a ranked keyword table with computed opportunity scores, emphasising long-tail phrases with high search volume and low competition.
- **Synthetic data generation** – A deterministic keyword engine powers keyword variants, international clusters, competitor lists, and trend signals without requiring network calls or external APIs.

## Tech stack

- [React](https://react.dev/) with [Vite](https://vitejs.dev/) for the development environment.
- [Tailwind CSS](https://tailwindcss.com/) for styling the marketing and dashboard sections.
- [Zustand](https://zustand-demo.pmnd.rs/) for state management across keyword insights, competitor summaries, and growth signals.
- [TypeScript](https://www.typescriptlang.org/) for strict typing.

## Getting started

The project lives inside the `frontend` workspace. Install dependencies and run the development server from that directory.

```bash
cd frontend
npm install
npm run dev
```

The Vite server starts on port `5173` by default. Visit `http://localhost:5173` to explore the RankBeam experience.

## Available scripts

From `frontend`, the following npm scripts are available:

- `npm run dev` – start the Vite development server.
- `npm run build` – type-check the project and emit a production build.
- `npm run preview` – preview the production build locally.
- `npm run lint` – run ESLint across the codebase.

## Project structure

The high-level structure focuses on components and data helpers that drive the single-page experience:

```
frontend/
├── src/
│   ├── components/      # Landing panels, dashboards, and marketing sections
│   ├── data/            # Country catalogue and helpers
│   ├── hooks/           # Shared React hooks
│   ├── lib/             # Zustand store and deterministic keyword engine
│   ├── App.tsx          # Composition of the marketing + research panels
│   └── main.tsx         # React entry point
└── package.json         # Scripts and dependencies
```

## License

This repository is provided without an explicit license. Please reach out to the maintainers before using the code in production.
