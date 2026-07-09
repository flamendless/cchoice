# Grafana Dashboards for C-Choice

Starter dashboards for monitoring C-Choice via Prometheus.

## Prerequisites

### Local development

1. Run the app (`mage serve`) so `/metrics` is available.
2. Start Prometheus: `mage Prom` (uses [`prometheus/prometheus.yml`](../prometheus/prometheus.yml)).
3. Run Grafana locally (default `http://localhost:3000`).
4. Add a Prometheus data source in Grafana pointing at `http://localhost:9090` (UID `prometheus` works; dashboards use a **Data source** variable and pick the default Prometheus source on import).

### Grafana Cloud

Dashboards use a **Data source** template variable (regex excludes usage/ML metrics sources) and **job** / **scrape_job** filters for `integrations/metrics_endpoint/*` scrapes. On import, select your stack Prometheus source (e.g. `grafanacloud-prom`).

## Import dashboards

1. Grafana → **Dashboards** → **New** → **Import**.
2. Upload a JSON file from [`grafana/dashboards/`](dashboards/), or paste its contents.
3. Select the Prometheus data source when prompted (local) or confirm the **Data source** variable (Grafana Cloud).

## Dashboards

| File | UID | Purpose |
|------|-----|---------|
| `overview.json` | `cchoice-overview` | Promo products, client events, HTTP requests (matches Grafana Cloud Metrics Overview) |
| `user-activity.json` | `cchoice-user-activity` | Page visits, actions, search, product clicks |
| `commerce.json` | `cchoice-commerce` | Orders, checkout, auth, shopping signals |

## Key PromQL queries

### Overview

- Promo impressions: `sum without (datetime) (cchoice_promo_product_impressions_total)`
- Client events: `cchoice_client_client_event`
- HTTP request series: `count(cchoice_http_requests_total)`, `count(cchoice_http_routes_skipped_total)`, `rate(cchoice_http_errors_total[$__rate_interval])`

### User activity

- Events by type: `sum by (event) (rate(cchoice_client_client_event[5m]))`
- Page visits: `sum by (value) (rate(cchoice_client_client_event{event=~"admin_visit|customer_visit|anon_visit"}[5m]))`
- Actions: `sum by (value) (rate(cchoice_client_client_event{event=~".*_exec"}[5m]))`

### Commerce

- Orders created: `sum by (payment_method) (rate(cchoice_orders_created_total[1h]))`
- Orders paid: `rate(cchoice_orders_paid_total[1h])`
- Login attempts: `sum by (user_type, result) (rate(cchoice_auth_login_attempts_total[5m]))`
- Add to cart: `sum(rate(cchoice_client_client_event{event="add_to_cart"}[$__rate_interval]))`
