0a. Study `specs/*` with multiple parallel sub-tasks to learn the application specifications. Understand user stories, acceptance criteria, and constraints.
0b. Study @IMPLEMENTATION_PLAN.md (if present) to understand the current plan state and what may already be complete or in progress.
0c. Study `src/lib/*` with parallel sub-tasks to understand shared utilities, components, and established patterns (the project's standard library).
0d. For reference, the application source code is in `src/*`.

1. Perform gap analysis: Study @IMPLEMENTATION_PLAN.md (if present; it may be incorrect or stale) and use multiple sub-tasks to examine existing source code in `src/*` and compare it against `specs/*`. Use a high-capability sub-task to analyze findings, prioritize tasks, and create/update @IMPLEMENTATION_PLAN.md as a markdown bullet point list sorted by priority. Ultrathink. Search for:
   - TODOs, FIXMEs, and incomplete implementations
   - Minimal implementations, placeholders, and stubs
   - Skipped, flaky, or missing tests
   - Inconsistent patterns across the codebase
   - Features specified but not implemented
   - Acceptance criteria not yet verified

2. Structure the plan with clear priority sections:
   - **High Priority**: Blocking issues, core functionality, critical path items
   - **Medium Priority**: Supporting features, user-facing enhancements
   - **Low Priority / Future Work**: Nice-to-haves, technical debt, optimizations

3. Write actionable tasks that are:
   - Specific and verifiable (what "done" looks like)
   - Small enough to complete in one focused session
   - Referenced to relevant spec files when applicable (e.g., `per specs/auth.md US-001`)
   - One capability per task (if you need "and" to describe it, split it)

IMPORTANT: Plan only. Do NOT implement anything. Do NOT assume functionality is missing; confirm with code search (grep, find_code) first. Treat `src/lib` as the project's standard library for shared utilities and components. Prefer consolidated, idiomatic implementations there over ad-hoc copies.

99999. If an element is missing from specs, search first to confirm it doesn't exist in code, then author the specification at `specs/FILENAME.md` using proper PRD format with user stories and acceptance criteria.
999999. The plan is disposable — regenerate freely when wrong, stale, or after significant spec changes. Time spent planning prevents wasted building loops.
9999999. Document any architectural discoveries, patterns, or gotchas in @AGENTS.md (keep it brief and operational only).
99999999. Each spec should cover one topic of concern. Topic scope test: can you describe it in one sentence without using "and" to conjoin unrelated capabilities? If not, it's multiple topics.

ULTIMATE GOAL: Create a comprehensive, prioritized implementation roadmap that enables autonomous building. The plan should make gaps visible, dependencies clear, and progress trackable. Future building phases will execute tasks from this plan one at a time with fresh context each iteration.

# Stop Condition

After completing the gap analysis and creating/updating IMPLEMENTATION_PLAN.md:

1. Verify all specs/\* have been analyzed
2. Verify all src/\* have been examined for gaps
3. Verify IMPLEMENTATION_PLAN.md contains prioritized, actionable tasks

If the plan is comprehensive and ready for building, reply with: <promise>COMPLETE</promise>

If more analysis is needed or specs are incomplete, end your response normally (another iteration will continue the planning).
