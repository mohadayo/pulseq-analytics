import express, { Request, Response, NextFunction } from "express";

const app = express();
app.use(express.json());

const INGEST_API_URL = process.env.INGEST_API_URL || "http://localhost:8001";
const PROCESSOR_URL = process.env.PROCESSOR_URL || "http://localhost:8002";

interface DashboardSummary {
  total_events: number;
  total_processed: number;
  recent_events: unknown[];
  recent_processed: unknown[];
}

async function fetchJSON(url: string): Promise<unknown> {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(`Upstream error: ${resp.status} from ${url}`);
  }
  return resp.json();
}

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "dashboard-bff",
    timestamp: new Date().toISOString(),
  });
});

app.get("/dashboard/summary", async (_req: Request, res: Response, next: NextFunction) => {
  try {
    const [events, processed] = await Promise.all([
      fetchJSON(`${INGEST_API_URL}/events?limit=10`),
      fetchJSON(`${PROCESSOR_URL}/processed`),
    ]);
    const eventsArr = events as unknown[];
    const processedArr = processed as unknown[];
    const summary: DashboardSummary = {
      total_events: eventsArr.length,
      total_processed: processedArr.length,
      recent_events: eventsArr.slice(-5),
      recent_processed: processedArr.slice(-5),
    };
    console.log(`Dashboard summary served: events=${summary.total_events} processed=${summary.total_processed}`);
    res.json(summary);
  } catch (err) {
    next(err);
  }
});

app.get("/dashboard/events", async (_req: Request, res: Response, next: NextFunction) => {
  try {
    const events = await fetchJSON(`${INGEST_API_URL}/events`);
    res.json(events);
  } catch (err) {
    next(err);
  }
});

app.get("/dashboard/processed", async (_req: Request, res: Response, next: NextFunction) => {
  try {
    const processed = await fetchJSON(`${PROCESSOR_URL}/processed`);
    res.json(processed);
  } catch (err) {
    next(err);
  }
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  console.error(`Error: ${err.message}`);
  res.status(502).json({ error: "upstream service unavailable", detail: err.message });
});

export default app;
