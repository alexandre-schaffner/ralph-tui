0a. Study `specs/*` with multiple subagents to learn the application specifications.
0b. Study @IMPLEMENTATION_PLAN.md.
0c. For reference, the application source code is in `src/*`.

1. Your task is to implement functionality per the specifications using available tools and sub-tasks. Follow @IMPLEMENTATION_PLAN.md and choose the most important item to address. Before making changes, search the codebase (don't assume not implemented) using tools like grep or view, and sub-tasks if needed. You may invoke multiple sub-tasks in sequence or parallel where supported by OpenCode for searches/reads, and limit to one for build/tests. Use high-capability models via configured agents when complex reasoning is needed (debugging, architectural decisions). Leverage OpenCode's tools such as bash for execution, view/grep for reading files, write/edit/patch for modifications.
2. After implementing functionality or resolving problems, run the tests for that unit of code that was improved using the bash tool. If functionality is missing, add it as per the application specifications. Ultrathink.
3. When you discover issues, immediately update @IMPLEMENTATION_PLAN.md with your findings using a sub-task or file tools. When resolved, update and remove the item.
4. When the tests pass, update @IMPLEMENTATION_PLAN.md, then use bash to `git add -A` then `git commit` with a message describing the changes. After the commit, `git push`.

99999. Important: When authoring documentation, capture the why — tests and implementation importance.
999999. Important: Single sources of truth, no migrations/adapters. If tests unrelated to your work fail, resolve them as part of the increment.
9999999. As soon as there are no build or test errors create a git tag using bash. If there are no git tags start at 0.0.0 and increment patch by 1 for example 0.0.1 if 0.0.0 does not exist.
99999999. You may add extra logging if required to debug issues.
999999999. Keep @IMPLEMENTATION_PLAN.md current with learnings using a sub-task or tools — future work depends on this to avoid duplicating efforts. Update especially after finishing your turn.
9999999999. When you learn something new about how to run the application, update @AGENTS.md using tools but keep it brief. For example if you run commands multiple times before learning the correct command then that file should be updated.
99999999999. For any bugs you notice, resolve them or document them in @IMPLEMENTATION_PLAN.md using a sub-task even if it is unrelated to the current piece of work.
999999999999. Implement functionality completely. Placeholders and stubs waste efforts and time redoing the same work.
9999999999999. When @IMPLEMENTATION_PLAN.md becomes large periodically clean out the items that are completed from the file using a sub-task.
99999999999999. If you find inconsistencies in the specs/\* then use a high-capability sub-task (e.g., with a powerful model) with 'ultrathink' requested to update the specs.
999999999999999. IMPORTANT: Keep @AGENTS.md operational only — status updates and progress notes belong in `IMPLEMENTATION_PLAN.md`. A bloated AGENTS.md pollutes every future loop's context.

# Stop Condition

After completing a user story, check if ALL stories have passes: true.

If ALL stories are complete and passing, reply with: <promise>COMPLETE</promise>

If there are still stories with passes: false, end your response normally (another iteration will pick up the next story).
