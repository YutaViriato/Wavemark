---
title: Why a runtime
description: Serving the folder directly, with no build, changes what publishing means.
date: 2026-11-01
author: The Wavemark team
icon: bolt
---

Static site generators solve a real problem: turn text into HTML once, and serve the result from anywhere. But they bring a step nobody asked for. You save, run the build, wait, check the output folder, and only then see what you wrote.

Wavemark removes that step. The process that serves the site is the same one that reads the files. When a `.md` file changes, the server rebuilds its snapshot in memory and the next request already gets the new page.

## The snapshot

Every rebuild produces an immutable snapshot: rendered pages, navigation, the index of routes and anchors. The swap happens at once, through an atomic pointer. A request never sees half a site.

> Publishing stops being an event and becomes a state: what is in the folder is what is live.

That is only safe because the runtime is strict. A broken link never reaches production, because the server refuses to start with errors. Once it is running, a page that starts having an error keeps serving its last valid version.

<Flow caption="Two snapshots never mix: the swap is a single pointer.">
<FlowStep title="Old snapshot" icon="layers">serving requests</FlowStep>
<FlowStep title="New snapshot" icon="gear">built on the side</FlowStep>
<FlowStep title="Swap" icon="bolt">one atomic store</FlowStep>
</Flow>

## The cost

A runtime needs a running process, which is a cost a static site does not have. For internal documentation, knowledge bases and sites that change all the time, the trade pays off. For a site that changes once a year, maybe not.

## What comes next

A static export is cheap to add, because rendering is already a pure function of the snapshot. It will arrive when someone needs it.
