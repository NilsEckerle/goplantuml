# Golden-file integration tests

Each case under `cases/<name>/` parses `input/` and compares the rendered
PlantUML against `expected.puml`.

## Run

```sh
go test ./goldentests
```

## Add a new edge case

1. Create the fixture. The code **must** live in a subdirectory named
   `input/` — that's how the harness discovers cases. Nested dirs inside
   `input/` are fine (and are how you test multi-level package layouts):

   ```sh
   mkdir -p goldentests/cases/my_case/input
   # drop .go files into input/ (and nested dirs if you need depth)
   ```

2. (Optional) Add `goldentests/cases/my_case/config.json` to set flags.
   Omit it for a plain `goplantuml <dir>` run. Keys match the CLI flag
   names; all are optional and default to the zero value:

   ```json
   {
     "recursive": false,
     "ignore": [],
     "max-depth": 0,
     "show-aggregations": false,
     "hide-fields": false,
     "hide-methods": false,
     "hide-connections": false,
     "show-compositions": false,
     "show-implementations": false,
     "show-aliases": false,
     "show-connection-labels": false,
     "aggregate-private-members": false,
     "hide-private-members": false,
     "title": ""
   }
   ```

   Note: `show-compositions`, `show-implementations`, and `show-aliases`
   only take effect when `hide-connections` is `true`, matching the
   behavior of `cmd/goplantuml/main.go`. An unknown or misspelled key
   makes the test fail loudly rather than being silently ignored.

3. Generate the golden file from the real parser output:

   ```sh
   go test ./goldentests -run TestGolden -update
   ```

4. Inspect `git diff` on the new `expected.puml`. If it looks right, commit
   the input files, `config.json`, and `expected.puml` together.

Never hand-write `expected.puml` — always generate it with `-update`, then
review the diff. That keeps the golden honest about what the parser produces.

## When a test fails

A mismatch prints the case name. Regenerate with `-update`, then use
`git diff goldentests/cases/<name>/expected.puml` to see exactly what
changed. If the change is intended, commit the new golden; if not, you found
a regression.
