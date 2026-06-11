#!/usr/bin/env bash
# Simulated Claude Code session for the README demo (~30s).
# Rendered by assets/demo.tape — do not run gates against this; it is a visual script.
set -u

DIM=$'\e[2m'; RED=$'\e[31m'; GRN=$'\e[32m'; YLW=$'\e[33m'; CYN=$'\e[36m'
BLD=$'\e[1m'; RST=$'\e[0m'

p() { printf '%b\n' "$1"; }

W=74
box_top() { printf '╭'; printf '─%.0s' $(seq 1 $W); printf '╮\n'; }
box_bot() { printf '╰'; printf '─%.0s' $(seq 1 $W); printf '╯\n'; }
box_type() { # typewriter-print the user prompt inside the box
  local s="$1"
  printf '│ %b ' "${BLD}>${RST}"
  local i
  for ((i = 0; i < ${#s}; i++)); do
    printf '%s' "${s:$i:1}"
    sleep 0.03
  done
  printf '%*s│\n' $((W - 3 - ${#s})) ''
}

clear
sleep 0.6
p "${DIM}Claude Code · my-app (main) · centinela 0.11.4 hooks active${RST}"
p ""
box_top
box_type "implement password-reset: email token with expiry"
box_bot
sleep 0.8

p ""
p "⏺ I'll implement password reset — starting with the token service."
sleep 1.0
p ""
p "⏺ ${BLD}Write${RST}(internal/auth/reset_token.go)"
sleep 0.7
p "  ⎿ ${RED}⛔ Blocked by PreToolUse hook${RST}"
p "    ${RED}🛡️ 👁️  CENTINELA DIRECTIVE: no active workflow.${RST}"
p "    ${RED}Code writes require an active feature: centinela start <feature>${RST}"
sleep 2.6
p ""
p "⏺ This repo is governed by Centinela — starting the workflow first."
sleep 1.0
p ""
p "⏺ ${BLD}Bash${RST}(centinela start password-reset)"
sleep 0.6
p "  ⎿ 🛡️ 👁️  workflow ${BLD}password-reset${RST} → step ${BLD}1/5: plan${RST}"
p "    ${DIM}required: feature brief + plan + Gherkin spec${RST}"
sleep 1.8
p ""
p "⏺ ${BLD}Write${RST}(docs/plans/password-reset.md)"
sleep 0.5
p "  ⎿ ok ${DIM}[workflow: password-reset | step: plan | 1/5]${RST}"
sleep 0.7
p "⏺ ${BLD}Write${RST}(specs/password-reset.feature)"
sleep 0.5
p "  ⎿ ok ${DIM}[workflow: password-reset | step: plan | 1/5]${RST}"
sleep 1.0
p ""
p "⏺ ${BLD}Bash${RST}(centinela complete password-reset)"
sleep 0.6
p "  ⎿ ${GRN}✓ plan artifacts verified${RST} → step ${BLD}2/5: code${RST}"
sleep 1.2
p ""
p "⏺ ${BLD}Write${RST}(internal/auth/reset_token.go)"
sleep 0.5
p "  ⎿ ok ${DIM}[workflow: password-reset | step: code | 2/5]${RST}"
sleep 1.0
p "⏺ ${BLD}Write${RST}(internal/auth/reset_token_test.go)  ${DIM}· step 3/5: tests${RST}"
sleep 0.5
p "  ⎿ ok ${DIM}[workflow: password-reset | step: tests | 3/5]${RST}"
sleep 1.2
p ""
p "⏺ ${BLD}Bash${RST}(centinela validate)  ${DIM}· step 4/5${RST}"
sleep 0.8
p "  ⎿ ${GRN}✓${RST} G1 file-size      ${DIM}all files ≤ 100 lines${RST}"
sleep 0.6
p "    ${GRN}✓${RST} G2 import-graph   ${DIM}no layer violations${RST}"
sleep 0.6
p "    ${GRN}✓${RST} security          ${DIM}no secrets · no vulnerable deps${RST}"
sleep 0.6
p "    ${GRN}✓${RST} tests             ${DIM}unit + integration + acceptance green${RST}"
sleep 1.4
p ""
p "⏺ ${BLD}Bash${RST}(centinela complete password-reset)  ${DIM}· docs generated · step 5/5${RST}"
sleep 0.8
p "  ⎿ ${GRN}✓ workflow complete${RST}"
sleep 1.0
p ""
p "🛡️ 👁️  ${BLD}password-reset shipped:${RST} plan ${GRN}✓${RST} spec ${GRN}✓${RST} tests ${GRN}✓${RST} gates ${GRN}✓${RST} ${DIM}— enforced, not requested.${RST}"
sleep 3.5
