# DriftMap roadmap

DriftMap compares YAML configuration across files, directories, and Git revisions.
The next phase focuses on trustworthy comparisons, automation, and a more usable
terminal summary. These estimates describe planned work, not completed features.

## 1. Report incomplete directory comparisons (3–5 hours)

- Return actionable errors for unreadable or invalid YAML instead of silently
  skipping files and potentially reporting no differences.
- Include the filename and side of the comparison in diagnostics.
- Cover malformed files, read failures, and mixed valid/invalid directories.

## 2. Preserve YAML comparison semantics (5–8 hours)

- Detect additions, removals, and changes involving empty maps and lists.
- Distinguish literal dotted/bracketed keys from nested paths and list indices.
- Define behavior for nulls, type changes, and multi-document YAML, with regression
  tests for the supported behavior.

## 3. Support CI and structured output (4–6 hours)

- Add an opt-in exit-code policy that distinguishes drift from execution errors.
- Produce consistent machine-readable directory results, including added and
  removed files, without mixing prose into structured output.
- Validate incompatible flags and document automation examples.

## 4. Improve the terminal summary (4–6 hours)

The initial two-file summary dashboard is implemented. Remaining work:

- Adapt to narrow terminals and long paths, labels, and section names.
- Improve Unicode display-width handling and terminal capability detection.
- Make color behavior consistent across all CLI output formats.
- Refine and test the deployment-impact heuristic, including added configuration;
  it is an estimate, not a prediction of actual rollout behavior.

## 5. Strengthen the HTTP interface (3–5 hours)

- Limit request bodies and configure server timeouts.
- Validate request shape and reject trailing or unexpected input consistently.
- Add handler tests for successful comparisons and invalid requests.

## 6. Improve installation and release documentation (3–5 hours)

- Provide a straightforward installation path for a binary named `driftmap`.
- Verify release artifacts and add common ARM64 targets.
- Update usage examples, explain comparison limitations, and clearly label
  synthetic sample configuration.

Estimated remaining work: **22–35 hours**. Priorities and estimates may change as
implementation uncovers additional requirements.
