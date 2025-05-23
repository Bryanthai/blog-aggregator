package config

import(
	"os"
	"encoding/json"
)

const configFileName = ".gatorconfig.json"

type Config struct{
	Db_url string `json: db_url`
	Username string `json: current_user_name`
}

func getFilePath () (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filePath := homeDir + "/" + configFileName

	return filePath, nil
}

func Read() (Config, error) {

	filePath, err := getFilePath()
	if err != nil {
		return Config{}, err
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(jsonData, &cfg); err != nil {
    	return Config{}, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(username string) error {

	cfg.Username = username
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	filePath, err := getFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonData, 0100664)
	if err != nil {
		return err
	}
	return nil
}
