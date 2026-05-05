---
id: "disc-2"
title: "Fix: DecisionLog handler double-resolution bug"
priority: "P0"
dependencies: []
status:
breaking: true
---

# disc-2: Fix: DecisionLog handler double-resolution bug

DecisionLog handler resolves URL bizKey parameter to internal mainItem ID via ResolveBizKey, then passes that internal ID to service methods that treat it as a bizKey again (calling FindByBizKey with the ID). This double-resolution causes all Create and List operations to fail. Fix: use ParseBizKeyParam to pass raw bizKey directly, update service and repo interfaces to accept int64 (bizKey) instead of uint (internal ID).
