---
title: Web dashboard
description: The supervisor hosts a built-in web dashboard for all your cities.
---

The Gas City supervisor hosts a built-in web dashboard. It is a single-page app
compiled into the `gc` binary and served by the supervisor on its own listener,
so there is nothing extra to install or run.

## Open it

Start the supervisor, then open the URL it prints:

```bash
gc supervisor start
# Supervisor API listening on http://127.0.0.1:8372
# Dashboard:  http://127.0.0.1:8372/
```

If the supervisor is already running, `gc dashboard` opens it in your browser
and prints the URL too:

```bash
gc dashboard
# Opened the dashboard in your browser: http://127.0.0.1:8372
```

Pass `--no-open` to print the URL without launching a browser (useful over SSH
or in scripts):

```bash
gc dashboard --no-open
# The dashboard is served by the gc supervisor at http://127.0.0.1:8372
```

`gc dashboard` does not start a server — it points your browser at the running
supervisor. The supervisor is the host, and one supervisor serves every
registered city. Pick the city you want from the switcher in the dashboard
header. If the supervisor is not running, `gc dashboard` prints how to start it
instead of opening a (dead) URL.

## What it shows

The dashboard reads the supervisor's typed API directly (same origin), so it
reflects live state: agents and their sessions, beads, mail, formula runs, and a
health view (system, local tools, per-rig store health, and the dolt store
trend).

## Filter the Beads board

Choose a rig to restrict the board to that project's ledger. The parent filter
selects all beads, beads without a parent, beads with a parent, or children of
one exact parent ID.

Use the metadata filter for your own intake conventions. Enter one condition
per line: `source=chat` requires an exact value; a bare key such as `thread_id`
requires that key to exist. Apply the conditions to combine them with the
parent and rig filters. Filters are kept in the URL so you can bookmark or share
the view.

For a requests-only view, choose **No parent** and filter on the metadata your
intake writes on request beads. Thread metadata copied onto workflow steps is
not, by itself, proof that a bead is an original request. Closed requests
require the **closed** status control.

The server filters before pagination, so newer workflow steps cannot crowd
matching requests out of the board's fetch window. See
[Filter beads](/reference/api#filter-beads) for the API parameters.

## Security posture

The dashboard is served on the supervisor's bind address, which defaults to
loopback (`127.0.0.1`). It is intended for local, single-operator use:

- It is same-origin with the API; browser mutations carry the supervisor's
  `X-GC-Request` CSRF header.
- When the supervisor binds a non-localhost address without
  `allow_mutations`, it runs read-only and the dashboard disables its mutating
  controls.

## Turn it off

Set `GC_SUPERVISOR_DASHBOARD=0` before starting the supervisor to run a
typed-API-only supervisor with no embedded dashboard.
