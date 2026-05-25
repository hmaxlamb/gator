package config

import (
    "os"
    "encoding/json"
)

const configFileName = ".gatorconfig.json"

type Config struct {
    Db_url string `json:"Db_url"`
    Username string `json:"Username"`
}

func Read() (Config, error) {
    path, err := getConfigFilePath()
    if err != nil {
        return Config{}, err
    }

    config_file, err := os.Open(path)
    if err != nil {
        return Config{}, err
    }

    data := make([]byte, 100)
    data_len, err := config_file.Read(data)
    if err != nil {
        return Config{}, err
    }

    data = data[:data_len]

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return Config{}, err
    }

    return config, nil
}

func getConfigFilePath() (string, error) {
dir, err := os.UserConfigDir()
if err != nil {
    return "", err
}

path := dir + "/" + "gator" + "/" + configFileName

return path, nil
}


func (config *Config) SetUser(username string) error {
    config.Username = username
    
    err := write(config)
    if err != nil {
        return err
    }

    return nil
}

func write(config *Config) error {
    path, err := getConfigFilePath()
    if err != nil {
        return err
    }

    data, err := json.Marshal(config)
    if err != nil {
        return err
    }

    err = os.WriteFile(path, data, 0666)
    if err != nil {
        return err
    }

    return nil
}
