package abi

// abigen emits shared Solidity structs (e.g. KeyValue) once per contract file, so
// generating several contracts into this one package double-declares them; the
// dedupe step strips every duplicate after the first file alphabetically.

//go:generate go run ./gen/abigen_jsons.go . abi
//go:generate go run ./gen/dedupe_shared_types.go .
//go:generate go run ./gen/add_raw_methods.go .
//go:generate gofmt -w .
