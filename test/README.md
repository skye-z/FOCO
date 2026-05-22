# FOCO Delivery E2E

This folder contains the Edge browser delivery test for the `confirmed` app.

## What It Covers

- Data reset hook and default admin seeding.
- Admin login, overview, settings, content package import/export.
- Question bank tree, filters, question detail dialog, create-dialog smoke checks.
- User management search, role filter, and action dialogs.
- Learner login, onboarding, diagnostic, home dashboard.
- Learning-path practice session completion.
- Interactive unit completion.
- Wrong-book, home refresh, profile archive/records/portrait.
- Logout checks for both admin and learner.

## Required Services

Run production frontend services before testing:

```bash
# admin
cd /Users/zhaoguiyang/Desktop/Workspace/AI/FOCO/confirmed/frontend/admin
/Users/zhaoguiyang/.nvm/versions/node/v20.12.0/bin/node node_modules/next/dist/bin/next start -p 3001

# learner
cd /Users/zhaoguiyang/Desktop/Workspace/AI/FOCO/confirmed/frontend/learner
/Users/zhaoguiyang/.nvm/versions/node/v20.12.0/bin/node node_modules/next/dist/bin/next start -p 3000
```

Run the backend API on `http://localhost:8080`.

## Install

```bash
cd /Users/zhaoguiyang/Desktop/Workspace/AI/FOCO/confirmed/test
npm install
npm run check
npm run typecheck
```

## Run

```bash
cd /Users/zhaoguiyang/Desktop/Workspace/AI/FOCO/confirmed/test
npm run test:e2e
```

The test uses Microsoft Edge by default through Playwright `channel: msedge`.

## Environment Variables

| Name | Default | Description |
| --- | --- | --- |
| `E2E_ADMIN_URL` | `http://localhost:3001` | Admin frontend URL. |
| `E2E_LEARNER_URL` | `http://localhost:3000` | Learner frontend URL. |
| `E2E_API_URL` | `http://localhost:8080` | Backend API URL. |
| `E2E_ADMIN_EMAIL` | `skai-zhang@hotmail.com` | Delivery admin account. |
| `E2E_ADMIN_PASSWORD` | `DevAdmin@2026` | Delivery admin password. |
| `E2E_CONTENT_PACKAGE` | `../cfa.json` | Content package path relative to this folder. |
| `E2E_RESET_COMMAND` | empty | Optional shell command to clear/rebuild test data. |
| `E2E_REQUIRE_RESET` | `false` | Set to `true` to fail when `E2E_RESET_COMMAND` is empty. |
| `E2E_HEADLESS` | `0` | Set to `1` for headless runs. |
| `E2E_BROWSER_CHANNEL` | `msedge` | Browser channel. |
| `E2E_SLOW_MO_MS` | `60` | Delay between browser actions for stable recordings. |

Example with a required reset hook:

```bash
E2E_REQUIRE_RESET=true \
E2E_RESET_COMMAND='your-db-reset-command-here' \
npm run test:e2e
```

## Outputs

- HTML test report: `confirmed/test/artifacts/html-report/index.html`
- JSON test report: `confirmed/test/artifacts/results.json`
- Video recordings: `confirmed/test/artifacts/test-results/**/video.webm`
- Failure screenshots and traces: `confirmed/test/artifacts/test-results/**`
