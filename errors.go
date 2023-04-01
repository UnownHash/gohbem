package gohbem

import "errors"

// ErrNilChannel is returned when o.watcherChan is uninitialized.
var ErrNilChannel = errors.New("can't close nil channel")

// ErrMasterFileUnloaded is returned when MasterFile wasn't loaded but there was a need to use it.
var ErrMasterFileUnloaded = errors.New("masterFile unloaded")

// ErrMasterFileOpen is returned when MasterFile can't be open.
var ErrMasterFileOpen = errors.New("can't open MasterFile")

// ErrMasterFileSave is returned when MasterFile can't be saved.
var ErrMasterFileSave = errors.New("can't save MasterFile")

// ErrMasterFileMarshall is returned when Marshal of MasterFile fail.
var ErrMasterFileMarshall = errors.New("can't marshal MasterFile")

// ErrMasterFileUnmarshall is returned when UnMarshal of MasterFile fail.
var ErrMasterFileUnmarshall = errors.New("can't unmarshal MasterFile")

// ErrMasterFileFetch is returned when remote fetch of MasterFile fail.
var ErrMasterFileFetch = errors.New("can't fetch remote MasterFile")

// ErrMasterFileDecode is returned when decode of MasterFile fail.
var ErrMasterFileDecode = errors.New("can't decode remote MasterFile")

// ErrWatcherStarted is returned when MasterFile Watcher is already running.
var ErrWatcherStarted = errors.New("MasterFile Watcher Already Started")

// ErrQueryInputOutOfRange is returned when wrong arguments are passed to QueryPvPRank function.
var ErrQueryInputOutOfRange = errors.New("one of input arguments 'Attack, Defense, Stamina, Level' is out of range")

// ErrMissingPokemon is returned when Pokemon is missing in MasterFile.
var ErrMissingPokemon = errors.New("missing pokemonID in MasterFile")

// ErrPvpStatBestCp is returned when BestCP > Cap in calculatePvPStat function.
var ErrPvpStatBestCp = errors.New("bestCP > cap")

// ErrLeaguesMissing is returned when Leagues configuration is empty.
var ErrLeaguesMissing = errors.New("leagues configuration is empty")

// ErrLevelCapsMissing is returned when levelCaps configuration is empty.
var ErrLevelCapsMissing = errors.New("levelCaps configuration is empty")
