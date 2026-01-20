#!/bin/bash
set -euo pipefail

# Ralph Loop Script
# Based on: https://github.com/ghuntley/how-to-ralph-wiggum
#
# Usage:
#   ./loop.sh [plan] [max_iterations]       # Plan/build on current branch
#   ./loop.sh plan-work "work description"  # Create scoped plan on current branch
#
# Examples:
#   ./loop.sh                               # Build mode, unlimited
#   ./loop.sh 20                            # Build mode, max 20
#   ./loop.sh plan                          # Full planning, unlimited
#   ./loop.sh plan 5                        # Full planning, max 5
#   ./loop.sh plan-work "user auth"         # Scoped planning
#

# Parse arguments
MODE="build"
PROMPT_FILE="PROMPT_build.md"
MAX_ITERATIONS=0

if [ "${1:-}" = "plan" ]; then
    # Full planning mode
    MODE="plan"
    PROMPT_FILE="PROMPT_plan.md"
    MAX_ITERATIONS=${2:-0}
elif [ "${1:-}" = "plan-work" ]; then
    # Scoped planning mode
    if [ -z "${2:-}" ]; then
        echo "Error: plan-work requires a work description"
        echo "Usage: ./loop.sh plan-work \"description of the work\""
        exit 1
    fi
    MODE="plan-work"
    WORK_DESCRIPTION="$2"
    PROMPT_FILE="PROMPT_plan_work.md"
    MAX_ITERATIONS=${3:-5}  # Default 5 for work planning
elif [[ "${1:-}" =~ ^[0-9]+$ ]]; then
    # Build mode with max iterations
    MAX_ITERATIONS=$1
fi

ITERATION=0
CURRENT_BRANCH=$(git branch --show-current)

# Model configuration (can be overridden via environment variable)
MODEL="opencode/claude-opus-4-5"

# Validate branch for plan-work mode
if [ "$MODE" = "plan-work" ]; then
    if [ "$CURRENT_BRANCH" = "main" ] || [ "$CURRENT_BRANCH" = "master" ]; then
        echo "Error: plan-work should be run on a work branch, not main/master"
        echo "Create a work branch first: git checkout -b ralph/your-work"
        exit 1
    fi

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Mode:    plan-work"
    echo "Branch:  $CURRENT_BRANCH"
    echo "Work:    $WORK_DESCRIPTION"
    echo "Prompt:  $PROMPT_FILE"
    echo "Model:   $MODEL"
    echo "Plan:    Will create scoped IMPLEMENTATION_PLAN.md"
    [ "$MAX_ITERATIONS" -gt 0 ] && echo "Max:     $MAX_ITERATIONS iterations"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Warn about uncommitted changes to IMPLEMENTATION_PLAN.md
    if [ -f "IMPLEMENTATION_PLAN.md" ] && ! git diff --quiet IMPLEMENTATION_PLAN.md 2>/dev/null; then
        echo "Warning: IMPLEMENTATION_PLAN.md has uncommitted changes that will be overwritten"
        read -p "Continue? [y/N] " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && exit 1
    fi

    # Export work description for PROMPT_plan_work.md
    export WORK_SCOPE="$WORK_DESCRIPTION"
else
    # Normal plan/build mode
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Mode:   $MODE"
    echo "Branch: $CURRENT_BRANCH"
    echo "Prompt: $PROMPT_FILE"
    echo "Model:  $MODEL"
    echo "Plan:   IMPLEMENTATION_PLAN.md"
    [ "$MAX_ITERATIONS" -gt 0 ] && echo "Max:    $MAX_ITERATIONS iterations"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
fi

# Verify prompt file exists
if [ ! -f "$PROMPT_FILE" ]; then
    echo "Error: $PROMPT_FILE not found"
    exit 1
fi

# Main loop
while true; do
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo "Reached max iterations: $MAX_ITERATIONS"

        if [ "$MODE" = "plan-work" ]; then
            echo ""
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            echo "Scoped plan created: $WORK_DESCRIPTION"
            echo "To build, run:"
            echo "  ./loop.sh 20"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        fi
        break
    fi

    # Run Ralph iteration with selected prompt using opencode
    # opencode run: Non-interactive mode for scripting/automation
    # --model: Model configuration to use
    # --agent: Agent configuration to use
    #
    # Output is captured while also being displayed in real-time via tee
    # The || true prevents script exit on command failure (due to set -e)

    # For plan-work mode, substitute ${WORK_SCOPE} in prompt before passing
    if [ "$MODE" = "plan-work" ]; then
        PROMPT_CONTENT=$(envsubst < "$PROMPT_FILE")
    else
        PROMPT_CONTENT=$(cat "$PROMPT_FILE")
    fi

    OUTPUT=$(opencode run "$PROMPT_CONTENT" \
        --model "$MODEL" \
        --agent "build" 2>&1 | tee /dev/stderr) || true

    # Check for completion signal
    if echo "$OUTPUT" | grep -q "<promise>COMPLETE</promise>"; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "✓ All tasks complete!"
        echo "  Completed at iteration $((ITERATION + 1))"
        echo "  Branch: $CURRENT_BRANCH"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        exit 0
    fi

    # Push changes after each iteration
    CURRENT_BRANCH=$(git branch --show-current)
    git push origin "$CURRENT_BRANCH" || {
        echo "Failed to push. Creating remote branch..."
        git push -u origin "$CURRENT_BRANCH"
    }

    ITERATION=$((ITERATION + 1))
    echo -e "\n\n======================== LOOP $ITERATION ========================\n"
done
