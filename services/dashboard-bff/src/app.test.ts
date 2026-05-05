import request from "supertest";
import app from "./app";

describe("Dashboard BFF", () => {
  describe("GET /health", () => {
    it("returns healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("healthy");
      expect(res.body.service).toBe("dashboard-bff");
      expect(res.body.timestamp).toBeDefined();
    });
  });

  describe("GET /dashboard/summary", () => {
    it("returns 502 when upstream services are unavailable", async () => {
      const res = await request(app).get("/dashboard/summary");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("upstream service unavailable");
    });
  });

  describe("GET /dashboard/events", () => {
    it("returns 502 when ingest API is unavailable", async () => {
      const res = await request(app).get("/dashboard/events");
      expect(res.status).toBe(502);
    });
  });

  describe("GET /dashboard/processed", () => {
    it("returns 502 when processor is unavailable", async () => {
      const res = await request(app).get("/dashboard/processed");
      expect(res.status).toBe(502);
    });
  });

  describe("unknown routes", () => {
    it("returns 404 for unknown paths", async () => {
      const res = await request(app).get("/unknown");
      expect(res.status).toBe(404);
    });
  });
});
