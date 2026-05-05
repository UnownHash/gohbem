package gohbem

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// MasterFileURL is default remote address used to fetch MasterFile.
const MasterFileURL = "https://raw.githubusercontent.com/WatWowMap/Masterfile-Generator/master/master-latest-basics.json"

// masterFileMaxBytes caps the masterfile response size to guard against memory exhaustion.
// Real masterfile is ~1-2 MiB; 32 MiB is generous.
const masterFileMaxBytes = 32 << 20

// httpFetchTimeout is the timeout for masterfile HTTP requests.
const httpFetchTimeout = 30 * time.Second

// roundFactor5 is 10^5 for the common roundFloat(x, 5) call site.
const roundFactor5 = 100000.0

// httpClient is reused across masterfile fetches to share TCP/TLS pools.
var httpClient = &http.Client{Timeout: httpFetchTimeout}

func roundFloat(val float64, precision uint) float64 {
	if precision == 5 {
		return math.Round(val*roundFactor5) / roundFactor5
	}
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func fetchMasterFile(url string) (PokemonData, error) {
	if url == "" {
		url = MasterFileURL
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PokemonData{}, ErrMasterFileFetch
	}
	req.Header.Set("User-Agent", fmt.Sprintf("Gohbem/%s", VERSION))

	resp, err := httpClient.Do(req)
	if err != nil {
		return PokemonData{}, ErrMasterFileFetch
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PokemonData{}, ErrMasterFileFetch
	}

	var data PokemonData
	if err := json.NewDecoder(io.LimitReader(resp.Body, masterFileMaxBytes)).Decode(&data); err != nil {
		return PokemonData{}, ErrMasterFileDecode
	}
	data.Initialized = true
	return data, nil
}

// safetyCheck is hot-path: must not take the RWMutex. Initialized state is
// mirrored to o.initialized (atomic) by Load/Fetch/Watch under the write lock.
// Leagues and LevelCaps are configured at construction and not mutated after,
// so reading their length lock-free is safe.
func safetyCheck(o *Ohbem) error {
	if o == nil || !o.initialized.Load() {
		return ErrMasterFileUnloaded
	}
	if len(o.Leagues) == 0 {
		return ErrLeaguesMissing
	}
	if len(o.LevelCaps) == 0 {
		return ErrLevelCapsMissing
	}
	return nil
}

// log logs the given message using the provided logger, if available. If no logger is set, the message is ignored.
func (o *Ohbem) log(message string) {
	if o.Logger != nil {
		o.Logger.Print(message)
	}
}
