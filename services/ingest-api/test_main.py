from fastapi.testclient import TestClient

from main import app, events_store

client = TestClient(app)


def setup_function():
    events_store.clear()


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    body = resp.json()
    assert body["status"] == "healthy"
    assert body["service"] == "ingest-api"
    assert "timestamp" in body


def test_create_event():
    resp = client.post("/events", json={"event_type": "click", "source": "web", "data": {"page": "/home"}})
    assert resp.status_code == 201
    body = resp.json()
    assert body["event_type"] == "click"
    assert body["source"] == "web"
    assert body["data"] == {"page": "/home"}
    assert "id" in body
    assert "received_at" in body


def test_create_event_missing_fields():
    resp = client.post("/events", json={"data": {}})
    assert resp.status_code == 422


def test_list_events():
    client.post("/events", json={"event_type": "click", "source": "web"})
    client.post("/events", json={"event_type": "purchase", "source": "mobile"})
    resp = client.get("/events")
    assert resp.status_code == 200
    assert len(resp.json()) == 2


def test_list_events_filter_by_type():
    client.post("/events", json={"event_type": "click", "source": "web"})
    client.post("/events", json={"event_type": "purchase", "source": "mobile"})
    resp = client.get("/events?event_type=click")
    assert resp.status_code == 200
    events = resp.json()
    assert len(events) == 1
    assert events[0]["event_type"] == "click"


def test_list_events_invalid_limit():
    resp = client.get("/events?limit=0")
    assert resp.status_code == 400


def test_get_event():
    resp = client.post("/events", json={"event_type": "click", "source": "web"})
    event_id = resp.json()["id"]
    resp = client.get(f"/events/{event_id}")
    assert resp.status_code == 200
    assert resp.json()["id"] == event_id


def test_get_event_not_found():
    resp = client.get("/events/nonexistent-id")
    assert resp.status_code == 404
