import logging
import os
import uuid
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s %(levelname)s %(name)s %(message)s",
)
logger = logging.getLogger("ingest-api")

app = FastAPI(title="PulseQ Ingest API", version="1.0.0")

events_store: list[dict] = []


class EventPayload(BaseModel):
    event_type: str = Field(..., min_length=1, max_length=128)
    source: str = Field(..., min_length=1, max_length=256)
    data: dict = Field(default_factory=dict)


class EventResponse(BaseModel):
    id: str
    event_type: str
    source: str
    data: dict
    received_at: str


@app.get("/health")
def health():
    return {"status": "healthy", "service": "ingest-api", "timestamp": datetime.now(timezone.utc).isoformat()}


@app.post("/events", response_model=EventResponse, status_code=201)
def create_event(payload: EventPayload):
    event = {
        "id": str(uuid.uuid4()),
        "event_type": payload.event_type,
        "source": payload.source,
        "data": payload.data,
        "received_at": datetime.now(timezone.utc).isoformat(),
    }
    events_store.append(event)
    logger.info("Event ingested: id=%s type=%s source=%s", event["id"], event["event_type"], event["source"])
    return event


@app.get("/events", response_model=list[EventResponse])
def list_events(event_type: str | None = None, limit: int = 100):
    if limit < 1 or limit > 1000:
        raise HTTPException(status_code=400, detail="limit must be between 1 and 1000")
    filtered = events_store
    if event_type:
        filtered = [e for e in filtered if e["event_type"] == event_type]
    return filtered[-limit:]


@app.get("/events/{event_id}", response_model=EventResponse)
def get_event(event_id: str):
    for event in events_store:
        if event["id"] == event_id:
            return event
    raise HTTPException(status_code=404, detail=f"Event {event_id} not found")
