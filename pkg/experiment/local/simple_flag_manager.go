package local

import (
	"time"

	"github.com/LIVEauctioneers/lafeature/internal/evaluation"
	"github.com/LIVEauctioneers/lafeature/internal/logger"
)

// SimpleFlagManagerConfig holds configuration for SimpleFlagManager
type SimpleFlagManagerConfig struct {
	// Timeout for API requests (default: 2s)
	Timeout time.Duration
	// Enable debug logging (default: false)
	Debug bool
}

// SimpleFlagManager provides a simple interface to check if flags are enabled
type SimpleFlagManager struct {
	flagConfigStorage flagConfigStorage
	flagConfigUpdater flagConfigUpdater
	flagConfigApi     flagConfigApi
	log               *logger.Log
}

// NewSimpleFlagManager creates a new flag manager that maintains a map[string]bool of enabled flags
func NewSimpleFlagManager(apiKey string, config SimpleFlagManagerConfig) *SimpleFlagManager {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 2 * time.Second
	}

	// Build internal config
	internalConfig := fillConfigDefaults(&Config{
		FlagConfigPollerRequestTimeout: config.Timeout,
		StreamFlagConnTimeout:          config.Timeout,
		Debug:                          config.Debug,
	})

	// Force streaming to be enabled
	internalConfig.StreamUpdates = true

	// Wrap storage with release flag filter
	flagConfigStorage := newReleaseFlagStorage(newInMemoryFlagConfigStorage())
	log := logger.New(internalConfig.Debug)

	// Setup streamer with poller fallback
	flagConfigStreamApi := newFlagConfigStreamApiV2(apiKey, internalConfig.StreamServerUrl, internalConfig.StreamFlagConnTimeout)
	flagConfigApi := newFlagConfigApiV2(apiKey, internalConfig.ServerUrl, internalConfig.FlagConfigPollerRequestTimeout)

	pollerFallback := newflagConfigFallbackRetryWrapper(
		newFlagConfigPoller(flagConfigApi, internalConfig, flagConfigStorage, nil, nil),
		nil,
		internalConfig.FlagConfigPollerInterval,
		updaterRetryMaxJitter,
		0,
		0,
		internalConfig.Debug,
	)

	flagConfigUpdater := newflagConfigFallbackRetryWrapper(
		newFlagConfigStreamer(flagConfigStreamApi, internalConfig, flagConfigStorage, nil, nil),
		pollerFallback,
		streamUpdaterRetryDelay,
		updaterRetryMaxJitter,
		internalConfig.FlagConfigPollerInterval,
		0,
		internalConfig.Debug,
	)

	return &SimpleFlagManager{
		flagConfigStorage: flagConfigStorage,
		flagConfigUpdater: flagConfigUpdater,
		flagConfigApi:     flagConfigApi,
		log:               log,
	}
}

// Start fetches initial flags via FlagsV2 API, then starts streaming updates
func (m *SimpleFlagManager) Start() error {
	// First, fetch flags synchronously via FlagsV2 API to populate initial state
	m.log.Debug("Fetching initial flags via FlagsV2 API")
	flagConfigs, err := m.flagConfigApi.getFlagConfigs()
	if err != nil {
		return err
	}

	// Populate storage with initial flags (only release flags)
	count := 0
	for _, flagConfig := range flagConfigs {
		if isReleaseFlag(flagConfig) {
			m.flagConfigStorage.putFlagConfig(flagConfig)
			count++
		}
	}
	m.log.Debug("Loaded %d initial release flags", count)

	// Now start streaming for real-time updates
	m.log.Debug("Starting flag stream updater")
	return m.flagConfigUpdater.Start(func(err error) {
		m.log.Error("Flag updater error: %v", err)
	})
}

// Stop stops the flag updater
func (m *SimpleFlagManager) Stop() {
	m.flagConfigUpdater.Stop()
}

// MustEnabled returns true if the flag is enabled, false otherwise
func (m *SimpleFlagManager) MustEnabled(flagKey string) (enabled bool) {
	flag := m.flagConfigStorage.getFlagConfig(flagKey)
	if flag == nil {
		return false
	}
	return isFlagEnabled(flag)
}

// Enabled returns true if the flag is enabled, false otherwise
func (m *SimpleFlagManager) Enabled(flagKey string) (enabled bool, ok bool) {
	flag := m.flagConfigStorage.getFlagConfig(flagKey)
	if flag == nil {
		return false, false
	}
	return isFlagEnabled(flag), true
}

// GetAll returns a copy of all flags as map[string]bool
func (m *SimpleFlagManager) GetAll() map[string]bool {
	flagConfigs := m.flagConfigStorage.getFlagConfigs()

	result := make(map[string]bool, len(flagConfigs))
	for key, flag := range flagConfigs {
		result[key] = isFlagEnabled(flag)
	}
	return result
}

// releaseFlagStorage wraps flagConfigStorage to filter and only store release flags
type releaseFlagStorage struct {
	storage flagConfigStorage
}

func newReleaseFlagStorage(storage flagConfigStorage) flagConfigStorage {
	return &releaseFlagStorage{storage: storage}
}

func (r *releaseFlagStorage) getFlagConfig(key string) *evaluation.Flag {
	return r.storage.getFlagConfig(key)
}

func (r *releaseFlagStorage) getFlagConfigs() map[string]*evaluation.Flag {
	return r.storage.getFlagConfigs()
}

func (r *releaseFlagStorage) getFlagConfigsArray() []*evaluation.Flag {
	return r.storage.getFlagConfigsArray()
}

func (r *releaseFlagStorage) putFlagConfig(flagConfig *evaluation.Flag) {
	// Only store release flags
	if isReleaseFlag(flagConfig) {
		r.storage.putFlagConfig(flagConfig)
	}
}

func (r *releaseFlagStorage) removeIf(condition func(*evaluation.Flag) bool) {
	r.storage.removeIf(condition)
}

// isReleaseFlag checks if a flag is a release flag
func isReleaseFlag(flag *evaluation.Flag) bool {
	if flag == nil || flag.Metadata == nil {
		return false
	}
	flagType, ok := flag.Metadata["flagType"].(string)
	return ok && flagType == "release"
}

// isFlagEnabled checks if a flag is enabled based on metadata
func isFlagEnabled(flag *evaluation.Flag) bool {
	// Try to access Metadata field directly
	if flag == nil || flag.Metadata == nil {
		return false
	}
	// Check if deployed is explicitly false
	if deployed, ok := flag.Metadata["deployed"].(bool); !ok || !deployed {
		return false
	}

	// Default to true if deployed is not set or is true
	return true
}
