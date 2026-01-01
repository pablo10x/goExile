# RedEye: Neural Threat Intelligence & Integrated Firewall

RedEye has been upgraded from a basic script to a high-performance, event-driven security engine that provides real-time protection, reputation tracking, and automated defense for the Exile Master Server.

## 1. Real-Time Architecture (The Event Bus)
*   **Zero-Latency Ingestion:** Replaced periodic database polling with a high-capacity, non-blocking internal **Signal Bus** (`signalChan`).
*   **Streaming Analysis:** A dedicated background worker (the "Brain") processes every security signal (HTTP requests, authentication failures, anti-cheat reports) as they happen.
*   **Batched Persistence:** To ensure maximum performance during attacks, security logs are batched and written to the database every 2 seconds, preventing I/O bottlenecks.

## 2. Smart Reputation Engine (In-Memory Tracking)
*   **IP Badness Scoring:** RedEye maintains an active map of all client behavior in memory (`ipScores`).
*   **Weighted Violations:** Different events impact reputation differently:
    *   **Normal Traffic:** Counted for statistics, 0 severity.
    *   **Rate Limit Hits:** Minor impact (+5).
    *   **Firewall Blocks:** Medium impact (+10).
    *   **Anti-Cheat Reports:** High impact (Severity × 2).
*   **Temporal Decay:** To prevent "false positives" over long periods, reputation scores **decay automatically** (halving every 60 seconds). Only active, aggressive threats reach the ban threshold.

## 3. Automated Defense & Banning
*   **Instant Threshold Trigger:** As soon as an IP's reputation score exceeds the threshold (default: 100), RedEye triggers an **immediate ban**.
*   **Multi-Layer Enforcement:**
    *   **Application Level:** IP is added to `BannedIPCache` for instant request rejection.
    *   **OS Level:** Engine automatically executes a system-level block via **Linux UFW** to drop packets at the network stack.
    *   **Persistence:** The ban is recorded in the `redeye_ip_reputation` table for long-term tracking.

## 4. Integrated Signal Sources
*   **Master Firewall:** Middleware feeds the engine with every incoming request, identifying scanning patterns and brute-force attempts.
*   **Game Instance Reports:** Integrated with the Anti-Cheat handler. Reports from game servers (e.g., speedhacks, memory manipulation) feed directly into the IP's reputation, allowing RedEye to ban cheaters at the gateway before they can reconnect.

## 5. High-Fidelity Observability
*   **Deep Telemetry:** The engine now tracks and exposes:
    *   `rt_active_trackers`: Number of IPs currently being monitored for behavior.
    *   `rt_queue_depth`: Current load on the security processing bus.
    *   `rt_cached_bans`: Number of bans currently enforced in high-speed memory.
*   **Granular Logging:** Every block and violation is stored with detailed metadata (Method, Path, DestPort, Protocol, and specific Violation Details).