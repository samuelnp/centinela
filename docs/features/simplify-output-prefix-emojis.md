# Feature Brief: Simplify Output Prefix to Emoji Pair

## Problem

Centinela output headers include verbose persona text and tone-specific faces. Users
want a simpler, consistent identity prefix.

## Goal

Replace the current prefix with exactly `🛡️👁️` across CLI and hook output lines.

## Scope

- Update persona label rendering to use a fixed emoji prefix.
- Keep channel/title metadata unchanged.
- Update tests that assert old label and face content.
