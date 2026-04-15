package storage

// SchemaVersion is the current highest migration number.
// The Rust dashboard uses this as the minimum required version to ensure
// compatibility between the CLI and the desktop app.
const SchemaVersion = 20
