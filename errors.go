package gohbem

import "errors"

// User input errors — caller passed something invalid.
var (
	// ErrQueryInputOutOfRange is returned when wrong arguments are passed to QueryPvPRank function.
	ErrQueryInputOutOfRange = errors.New("one of input arguments 'Attack, Defense, Stamina, Level' is out of range")

	// ErrMissingPokemon is returned when Pokemon is missing in MasterFile.
	ErrMissingPokemon = errors.New("missing pokemonID in MasterFile")
)

// Runtime state errors — Ohbem instance is not in a usable state.
var (
	// ErrMasterFileUnloaded is returned when MasterFile wasn't loaded but there was a need to use it.
	ErrMasterFileUnloaded = errors.New("masterFile unloaded")

	// ErrLeaguesMissing is returned when Leagues configuration is empty.
	ErrLeaguesMissing = errors.New("leagues configuration is empty")

	// ErrLevelCapsMissing is returned when levelCaps configuration is empty.
	ErrLevelCapsMissing = errors.New("levelCaps configuration is empty")

	// ErrNilChannel is returned when o.watcherChan is uninitialized.
	ErrNilChannel = errors.New("can't close nil channel")

	// ErrWatcherStarted is returned when MasterFile Watcher is already running.
	ErrWatcherStarted = errors.New("MasterFile Watcher Already Started")
)

// I/O errors — fetching, loading, or saving the MasterFile failed.
var (
	// ErrMasterFileOpen is returned when MasterFile can't be open.
	ErrMasterFileOpen = errors.New("can't open MasterFile")

	// ErrMasterFileSave is returned when MasterFile can't be saved.
	ErrMasterFileSave = errors.New("can't save MasterFile")

	// ErrMasterFileMarshall is returned when Marshal of MasterFile fail.
	ErrMasterFileMarshall = errors.New("can't marshal MasterFile")

	// ErrMasterFileUnmarshall is returned when UnMarshal of MasterFile fail.
	ErrMasterFileUnmarshall = errors.New("can't unmarshal MasterFile")

	// ErrMasterFileFetch is returned when remote fetch of MasterFile fail.
	ErrMasterFileFetch = errors.New("can't fetch remote MasterFile")

	// ErrMasterFileDecode is returned when decode of MasterFile fail.
	ErrMasterFileDecode = errors.New("can't decode remote MasterFile")
)

// Internal computation errors.
var (
	// ErrPvpStatBestCp is returned when BestCP > Cap in calculatePvPStat function.
	ErrPvpStatBestCp = errors.New("bestCP > cap")
)
