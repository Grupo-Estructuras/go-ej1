package fileconfig

import (
	"os"

	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type FileConfigStore struct {
	logger       zerolog.Logger
	filename     string
	swapFilename string
}

func NewFileConfigstore(logger zerolog.Logger, filename string) *FileConfigStore {
	var store FileConfigStore
	store.logger = logger
	store.filename = filename
	store.swapFilename = filename + ".tmp~"
	return &store
}

func (store *FileConfigStore) Load(configData interface{}) error {
	store.logger.Trace().Str("method", "Load").Msg("ENTRY")

	store.logger.Info().Str("method", "Load").Str("file", store.filename).Msg("Loading config from file")
	if store.fileExists(store.filename) {
		if err := store.canAccessFile(store.filename, os.O_RDWR); err != nil {
			store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
			return err
		}

		if err := store.canAccessFile(store.swapFilename, os.O_RDWR|os.O_CREATE); err != nil {
			store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
			return err
		}
	} else {
		store.logger.Info().Str("method", "Load").Str("file", store.filename).Msg("Config file does not exist; creating using dafaults.")
		if err := store.Save(configData); err != nil {
			store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
			return err
		}
	}

	file, err := os.OpenFile(store.filename, os.O_RDONLY, 0644)
	if err != nil {
		store.logger.Error().Str("method", "Load").Str("file", store.filename).Msg("Failed to read config-file")
		store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(configData); err != nil {
		store.logger.Error().Str("method", "Load").Str("file", store.filename).Msg("Failed to decode config-file")
		store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
		return err
	}

	store.logger.Info().Str("method", "Load").Str("file", store.filename).Msg("Writing back config-file to add new parameters.")
	if err := store.Save(configData); err != nil {
		store.logger.Trace().Str("method", "Load").Err(err).Msg("EXIT")
		return err
	}

	store.logger.Trace().Str("method", "Load").Msg("EXIT")
	return nil
}

func (store *FileConfigStore) Save(configData interface{}) error {
	store.logger.Trace().Str("method", "Save").Msg("ENTRY")

	path := filepath.Dir(store.filename)
	if err := os.MkdirAll(path, 0755); err != nil {
		store.logger.Error().Str("method", "Save").Str("file", store.filename).Err(err).Msg("Failed to create directory to save config-file")
		store.logger.Trace().Str("method", "Save").Err(err).Msg("EXIT")
		return err
	}

	file, err := os.OpenFile(store.swapFilename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		store.logger.Error().Str("method", "Save").Str("file", store.swapFilename).Err(err).Msg("Failed to open temporary config-file")
		store.logger.Trace().Str("method", "Save").Err(err).Msg("EXIT")
		return err
	}

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(configData); err != nil {
		store.logger.Error().Str("method", "Save").Str("file", store.swapFilename).Err(err).Msg("Failed to write temporary config-file")
		store.logger.Trace().Str("method", "Save").Err(err).Msg("EXIT")
		return err
	}

	if err := file.Close(); err != nil {
		store.logger.Error().Str("method", "Save").Str("file", store.swapFilename).Err(err).Msg("Failed to close temporary config-file")
		store.logger.Trace().Str("method", "Save").Err(err).Msg("EXIT")
		return err
	}

	if err := os.Rename(store.swapFilename, store.filename); err != nil {
		store.logger.Error().Str("method", "Save").Str("file", store.filename).Err(err).Msg("Failed to rename temporary config-file")
		store.logger.Trace().Str("method", "Save").Err(err).Msg("EXIT")
		return err
	}

	store.logger.Trace().Str("method", "Save").Msg("EXIT")
	return nil
}

func (store *FileConfigStore) fileExists(filename string) bool {
	store.logger.Trace().Str("method", "fileExists").Msg("ENTRY")

	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			store.logger.Trace().Str("method", "fileExists").Bool("return-value", false).Msg("EXIT")
			return false
		} else {
			store.logger.Error().Str("method", "fileExists").Str("filename", filename).Err(err).Msg("Failed to access config file")
		}

		store.logger.Trace().Str("method", "fileExists").Bool("return-value", false).Msg("EXIT")
		return false
	}

	store.logger.Trace().Str("method", "Save").Bool("return-value", true).Msg("EXIT")
	return true
}

func (store *FileConfigStore) canAccessFile(filename string, flags int) error {
	store.logger.Trace().Str("method", "canAccessFile").Msg("ENTRY")

	file, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		if err := file.Close(); err == nil {
			store.logger.Trace().Str("method", "canAccessFile").Msg("EXIT")
			return nil
		}
	}

	store.logger.Trace().Str("method", "canAccessFile").Err(err).Msg("EXIT")
	return err
}
